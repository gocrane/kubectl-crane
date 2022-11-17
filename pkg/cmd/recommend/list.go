package recommend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	analysisv1alph1 "github.com/gocrane/api/analysis/v1alpha1"
	"github.com/gocrane/kubectl-crane/pkg/cmd/options"
	"github.com/gocrane/kubectl-crane/pkg/util"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"strconv"
	"strings"
)

var (
	recommendListExample = `
# view all recommend result with kube-system namespace
%[1]s recommend list --namespace kube-system

# view Resource type recommend result with kube-system namespace
%[1]s recommend list --namespace kube-system --type Resource
`

	recommenderMap = map[string]int{"Replicas": 1, "Resource": 2, "IdleNode": 3}
)

type RecommendListOptions struct {
	CommonOptions *options.CommonOptions

	Name       string
	Type       string
	TargetKind string
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
		if _, ok := recommenderMap[o.Type]; !ok {
			return errors.New("the recommender only support Replicas,Resource and IdleNode")
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
	query := util.NewQuery()
	if len(o.Name) > 0 {
		query.Filters[util.FieldName] = util.Value(o.Name)
	}

	namespace := ""
	if len(*o.CommonOptions.ConfigFlags.Namespace) > 0 {
		namespace = *o.CommonOptions.ConfigFlags.Namespace
		//query.Filters[util.FieldNamespace] = util.Value(*o.CommonOptions.ConfigFlags.Namespace)
	}

	recommendResult, err := o.CommonOptions.CraneClient.AnalysisV1alpha1().Recommendations(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to get recommend result, %v.", err)
		return err
	}
	var recommendations []analysisv1alph1.Recommendation
	for _, recommendation := range recommendResult.Items {
		selected := true
		for field, value := range query.Filters {
			if !util.ObjectMetaFilter(recommendation.ObjectMeta, util.Filter{Field: field, Value: value}) {
				selected = false
				break
			}
		}
		if selected {
			if len(o.Type) > 0 && !strings.EqualFold(string(recommendation.Spec.Type), o.Type) {
				selected = false
			}
			if len(o.TargetKind) > 0 && !strings.EqualFold(recommendation.Spec.TargetRef.Kind, o.TargetKind) {
				selected = false
			}
		}

		if selected {
			recommendations = append(recommendations, recommendation)
		}
	}

	o.renderTable(recommendations)

	return nil
}

func (o *RecommendListOptions) renderTable(recommendations []analysisv1alph1.Recommendation) {
	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.SetOutputMirror(o.CommonOptions.Out)
	header := table.Row{}
	header = append(header, table.Row{"NAME", "RECOMMEND SOURCE", "NAMESPACE", "TARGET", "CURRENT RESOURCE", "RECOMMEND RESOURCE", "CREATED TIME", "UPDATED TIME"}...)
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
		row = append(row, recommendation.Spec.TargetRef.Name)
		row = append(row, recommendation.Namespace)
		row = append(row, recommendation.Spec.TargetRef.Kind)

		currentInfo := v1.Deployment{}
		if err := json.Unmarshal([]byte(recommendation.Status.RecommendationContent.CurrentInfo), &currentInfo); err != nil {
			row = append(row, "")
		} else {
			currentResource := ""
			if recommendation.Spec.Type == "Resource" {
				for _, container := range currentInfo.Spec.Template.Spec.Containers {
					currentResource += container.Name + "/" + container.Resources.Requests.Cpu().String() + "m/" + container.Resources.Requests.Memory().String() + "\n"
				}
			} else if recommendation.Spec.Type == "Replicas" {
				currentResource += strconv.Itoa(int(*currentInfo.Spec.Replicas))
			}

			row = append(row, currentResource)
		}

		recommendInfo := v1.Deployment{}
		if err := json.Unmarshal([]byte(recommendation.Status.RecommendationContent.RecommendedInfo), &recommendInfo); err != nil {
			row = append(row, "")
		} else {
			recommendResource := ""
			if recommendation.Spec.Type == "Resource" {
				for _, container := range recommendInfo.Spec.Template.Spec.Containers {
					recommendResource += container.Name + "/" + container.Resources.Requests.Cpu().String() + "/" + container.Resources.Requests.Memory().String() + "\n"
				}
			} else if recommendation.Spec.Type == "Replicas" {
				recommendResource += strconv.Itoa(int(*recommendInfo.Spec.Replicas))
			}

			row = append(row, recommendResource)
		}

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
	cmd.Flags().StringVarP(&o.Type, "type", "", "", "Specify the type for recommend[Resource, Replicas, IdleNode]")
	cmd.Flags().StringVarP(&o.Name, "name", "", "", "Specify the name for recommend")
	cmd.Flags().StringVarP(&o.TargetKind, "targetKind", "", "", "Specify the target type for recommendationrules")
}
