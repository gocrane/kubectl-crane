package recommendationRule

import (
	"context"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	analysisv1alph1 "github.com/gocrane/api/analysis/v1alpha1"
	"github.com/gocrane/kubectl-crane/pkg/cmd/options"
	"github.com/gocrane/kubectl-crane/pkg/utils"
)

type RecommendationRuleListOptions struct {
	CommonOptions *options.CommonOptions

	Name        string
	Recommender string
}

func NewRecommendationRuleListOptions() *RecommendationRuleListOptions {
	return &RecommendationRuleListOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

func NewCmdRecommendationRuleList() *cobra.Command {
	o := NewRecommendationRuleListOptions()

	command := &cobra.Command{
		Use:   "list",
		Short: "view recommendation rules",
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
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

func (o *RecommendationRuleListOptions) Validate() error {
	if err := o.CommonOptions.Validate(); err != nil {
		return err
	}

	return nil
}

func (o *RecommendationRuleListOptions) Complete(cmd *cobra.Command, args []string) error {
	if err := o.CommonOptions.Complete(cmd, args); err != nil {
		return err
	}

	return nil
}

func (o *RecommendationRuleListOptions) Run() error {
	query := utils.NewQuery()
	if len(o.Name) > 0 {
		query.Filters[utils.FieldName] = utils.Value(o.Name)
	}

	recommendationRuleResult, err := o.CommonOptions.CraneClient.AnalysisV1alpha1().RecommendationRules().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to get recommendation rules, %v.", err)
		return err
	}
	var recommendationRules []analysisv1alph1.RecommendationRule
	for _, recommendationRule := range recommendationRuleResult.Items {
		selected := true
		for field, value := range query.Filters {
			if !utils.ObjectMetaFilter(recommendationRule.ObjectMeta, utils.Filter{Field: field, Value: value}) {
				selected = false
				break
			}
		}
		if selected && len(o.Recommender) > 0 {
			for _, recommender := range recommendationRule.Spec.Recommenders {
				if !strings.EqualFold(recommender.Name, o.Recommender) {
					selected = false
				} else {
					selected = true
					break
				}
			}
		}

		if selected {
			recommendationRules = append(recommendationRules, recommendationRule)
		}
	}

	o.renderTable(recommendationRules)

	return nil
}

func (o *RecommendationRuleListOptions) renderTable(recommendationRules []analysisv1alph1.RecommendationRule) {
	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.SetOutputMirror(o.CommonOptions.Out)
	header := table.Row{}
	header = append(header, table.Row{"NAME", "RECOMMENDER", "TARGET", "NAMESPACE", "RUN INTERVAL", "LAST UPDATE TIME", "CREATE TIME"}...)
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

	for _, recommendRule := range recommendationRules {
		row := table.Row{}

		row = append(row, recommendRule.Name)

		var recommenders []string
		for _, recommender := range recommendRule.Spec.Recommenders {
			recommenders = append(recommenders, recommender.Name)
		}
		row = append(row, strings.Join(recommenders, ","))

		var targets []string
		for _, resourceSelector := range recommendRule.Spec.ResourceSelectors {
			targets = append(targets, resourceSelector.Kind)
		}
		row = append(row, strings.Join(targets, ","))

		var namespaces []string
		if recommendRule.Spec.NamespaceSelector.Any {
			namespaces = append(namespaces, "Any")
		} else {
			namespaces = append(namespaces, recommendRule.Spec.NamespaceSelector.MatchNames...)
		}

		row = append(row, strings.Join(namespaces, ","))

		row = append(row, recommendRule.Spec.RunInterval)
		row = append(row, recommendRule.Status.LastUpdateTime)
		row = append(row, recommendRule.CreationTimestamp)

		t.AppendRows([]table.Row{
			row,
		})

		t.AppendSeparator()
	}

	t.Render()
}

func (o *RecommendationRuleListOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Name, "name", "", "", "Specify name for recommendationrules")
	cmd.Flags().StringVarP(&o.Recommender, "recommender", "", "", "Specify type for recommendationrules")
}
