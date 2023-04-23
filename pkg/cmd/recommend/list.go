package recommend

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	analysisv1alpha1 "github.com/gocrane/api/analysis/v1alpha1"

	"github.com/gocrane/kubectl-crane/pkg/cmd/options"
	"github.com/gocrane/kubectl-crane/pkg/utils"
)

var (
	recommendListExample = `
# view all recommend result with all namespace
%[1]s recommend list

# view all recommend result with kube-system namespace
%[1]s recommend list --namespace kube-system

# view Resource type recommend result with kube-system namespace
%[1]s recommend list --namespace kube-system --type Resource
`
)

const (
	RecommendationRuleNameLabel          = "analysis.crane.io/recommendation-rule-name"
	RecommendationRuleUidLabel           = "analysis.crane.io/recommendation-rule-uid"
	RecommendationRuleRecommenderLabel   = "analysis.crane.io/recommendation-rule-recommender"
	RecommendationRuleTargetKindLabel    = "analysis.crane.io/recommendation-target-kind"
	RecommendationRuleTargetVersionLabel = "analysis.crane.io/recommendation-target-version"
	RecommendationRuleTargetNameLabel    = "analysis.crane.io/recommendation-target-name"
	RunNumberAnnotation                  = "analysis.crane.io/run-number"
)

type RecommendListOptions struct {
	CommonOptions *options.CommonOptions

	Name       string
	Type       string
	TargetKind string
	TargetName string
	RuleName   string
}

func NewRecommendListOptions() *RecommendListOptions {
	return &RecommendListOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

func NewCmdRecommendList() *cobra.Command {
	o := NewRecommendListOptions()

	command := &cobra.Command{
		Use:     "list",
		Short:   "view recommend result",
		Example: fmt.Sprintf(recommendListExample, "kubectl-crane"),
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				klog.Infof(fmt.Sprintf("\nExample:\n"+recommendListExample, "kubectl-crane"))
				return err
			}

			if err := o.Run(); err != nil {
				return err
			}

			return nil
		},
	}
	o.CommonOptions.AddCommonFlag(command)
	o.AddFlags(command)

	return command
}

func (o *RecommendListOptions) Validate() error {
	if err := o.CommonOptions.Validate(); err != nil {
		return err
	}

	if len(o.Type) > 0 {
		typeExist := false
		for _, recommenderType := range analysisv1alpha1.AllRecommenderType {
			if recommenderType == o.Type {
				typeExist = true
			}
		}
		if !typeExist {
			return fmt.Errorf("the recommender type not supported %s", o.Type)
		}
	}

	return nil
}

func (o *RecommendListOptions) Complete(cmd *cobra.Command, args []string) error {
	if err := o.CommonOptions.Complete(cmd, args); err != nil {
		return err
	}

	return nil
}

func (o *RecommendListOptions) Run() error {
	query := utils.NewQuery()
	if len(o.Name) > 0 {
		query.Filters[utils.FieldName] = utils.Value(o.Name)
	}

	if len(o.Type) > 0 {
		query.LabelSelector[RecommendationRuleRecommenderLabel] = o.Type
	}

	if len(o.TargetKind) > 0 {
		query.LabelSelector[RecommendationRuleTargetKindLabel] = o.TargetKind
	}

	if len(o.TargetName) > 0 {
		query.LabelSelector[RecommendationRuleTargetNameLabel] = o.TargetName
	}

	if len(o.RuleName) > 0 {
		query.LabelSelector[RecommendationRuleNameLabel] = o.RuleName
	}

	namespace := ""
	if len(*o.CommonOptions.ConfigFlags.Namespace) > 0 {
		namespace = *o.CommonOptions.ConfigFlags.Namespace
	}

	selector := ""
	for label, value := range query.LabelSelector {
		selector += label + "=" + value + ","
	}
	// remove the last ","
	if len(selector) > 0 {
		selector = selector[:len(selector)-1]
	}
	listOptions := metav1.ListOptions{
		LabelSelector: selector,
	}
	recommendResult, err := o.CommonOptions.CraneClient.AnalysisV1alpha1().Recommendations(namespace).List(context.TODO(), listOptions)
	if err != nil {
		klog.Errorf("Failed to get recommend result, %v.", err)
		return err
	}
	var recommendations []analysisv1alpha1.Recommendation
	for _, recommendation := range recommendResult.Items {
		selected := true
		for field, value := range query.Filters {
			if !utils.ObjectMetaFilter(recommendation.ObjectMeta, utils.Filter{Field: field, Value: value}) {
				selected = false
				break
			}
		}

		if selected {
			recommendations = append(recommendations, recommendation)
		}
	}

	RenderTable(recommendations, o.CommonOptions.Out)

	return nil
}

func RenderTable(recommendations []analysisv1alpha1.Recommendation, out io.Writer) {
	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.SetOutputMirror(out)
	header := table.Row{}
	header = append(header, table.Row{"NAME", "NAMESPACE", "TYPE", "TARGET NAME", "TARGET NAMESPACE", "TARGET KIND", "CURRENT RESOURCE", "RECOMMEND RESOURCE", "ACTION", "CREATED TIME", "UPDATED TIME"}...)
	t.AppendHeader(header)
	t.SetColumnConfigs([]table.ColumnConfig{
		{
			Name:        "NAME",
			Align:       text.AlignLeft,
			AlignFooter: text.AlignLeft,
			AlignHeader: text.AlignLeft,
			VAlign:      text.VAlignMiddle,
			WidthMin:    6,
			WidthMax:    24,
		},
	})

	for _, recommendation := range recommendations {
		row := table.Row{}

		row = append(row, recommendation.Name)
		row = append(row, recommendation.Namespace)
		row = append(row, recommendation.Spec.Type)

		row = append(row, recommendation.Spec.TargetRef.Name)
		row = append(row, recommendation.Namespace)
		row = append(row, recommendation.Spec.TargetRef.Kind)

		currentResource := ""
		recommendResource := ""
		switch recommendation.Spec.Type {
		case "Resource":
			var currentInfo analysisv1alpha1.PatchResource
			if err := json.Unmarshal([]byte(recommendation.Status.RecommendationContent.CurrentInfo), &currentInfo); err != nil {
				row = append(row, "")
			} else {
				for _, container := range currentInfo.Spec.Template.Spec.Containers {
					currentResource += container.Name + "/" + container.Resources.Requests.Cpu().String() + "/" + container.Resources.Requests.Memory().String() + "\n"
				}
			}

			var recommendInfo analysisv1alpha1.PatchResource
			if err := json.Unmarshal([]byte(recommendation.Status.RecommendationContent.RecommendedInfo), &recommendInfo); err != nil {
				row = append(row, "")
			} else {
				for _, container := range recommendInfo.Spec.Template.Spec.Containers {
					recommendResource += container.Name + "/" + container.Resources.Requests.Cpu().String() + "/" + container.Resources.Requests.Memory().String() + "\n"
				}
			}
		case "Replicas":
			var currentInfo analysisv1alpha1.PatchReplicas
			if err := json.Unmarshal([]byte(recommendation.Status.RecommendationContent.CurrentInfo), &currentInfo); err != nil {
				row = append(row, "")
			} else {
				currentResource += strconv.Itoa(int(*currentInfo.Spec.Replicas))
			}

			var recommendInfo analysisv1alpha1.PatchReplicas
			if err := json.Unmarshal([]byte(recommendation.Status.RecommendationContent.RecommendedInfo), &recommendInfo); err != nil {
				row = append(row, "")
			} else {
				recommendResource += strconv.Itoa(int(*recommendInfo.Spec.Replicas))
			}
		default:
			recommendResource = recommendation.Status.RecommendedInfo
			currentResource = recommendation.Status.CurrentInfo
		}

		row = append(row, currentResource)
		row = append(row, recommendResource)
		row = append(row, recommendation.Status.Action)

		row = append(row, recommendation.CreationTimestamp)
		row = append(row, recommendation.Status.LastUpdateTime)

		t.AppendRows([]table.Row{
			row,
		})

		t.AppendSeparator()
	}

	t.Render()
}

func (o *RecommendListOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Type, "type", "", "", "List recommendation with specify recommend type[Resource, Replicas, IdleNode]")
	cmd.Flags().StringVarP(&o.Name, "name", "", "", "Specify the name for recommendation")
	cmd.Flags().StringVarP(&o.TargetKind, "targetKind", "", "", "List recommendation with specify recommendation target kind")
	cmd.Flags().StringVarP(&o.TargetName, "targetName", "", "", "List recommendation with specify recommendation target name")
	cmd.Flags().StringVarP(&o.RuleName, "ruleName", "", "", "List recommendation with specify recommendationRule name")
}
