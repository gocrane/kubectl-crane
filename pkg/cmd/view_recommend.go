package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	analysisv1alph1 "github.com/gocrane/api/analysis/v1alpha1"
	"github.com/gocrane/kubectl-crane/pkg/cmd/options"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"reflect"
	"strconv"
)

var (
	viewRecommendExample = `
# view all recommend result with kube-system namespace
%[1]s view-recommend --sourceSelector '{"apiVersion":"","kind": "", "name": "", "namespace":""}'
`
)

type ViewRecommendOptions struct {
	CommonOptions *options.CommonOptions

	Selector string

	ResourceSelector corev1.ObjectReference
}

func NewViewRecommendOptions() *ViewRecommendOptions {
	return &ViewRecommendOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

func NewCmdViewRecommend() *cobra.Command {
	o := NewViewRecommendOptions()

	command := &cobra.Command{
		Use:     "view-recommend",
		Short:   "View a source which recommends related.",
		Example: fmt.Sprintf(viewRecommendExample, "kubectl-crane"),
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				klog.Infof(fmt.Sprintf("\nExample:\n"+viewRecommendExample, "kubectl-crane"))
				return err
			}

			if err := o.Run(); err != nil {
				return err
			}

			return nil
		},
	}
	o.AddFlags(command)

	return command
}

func (o *ViewRecommendOptions) Validate() error {
	if err := o.CommonOptions.Validate(); err != nil {
		return err
	}

	err := json.Unmarshal([]byte(o.Selector), &o.ResourceSelector)
	if err != nil {
		return errors.New("please check the recommender target is valid")
	}

	return nil
}

func (o *ViewRecommendOptions) Complete(cmd *cobra.Command, args []string) error {
	if err := o.CommonOptions.Complete(cmd, args); err != nil {
		return err
	}

	return nil
}

func (o *ViewRecommendOptions) Run() error {
	recommendResult, err := o.CommonOptions.CraneClient.AnalysisV1alpha1().Recommendations(o.ResourceSelector.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to get recommend result, %v.", err)
		return err
	}
	var recommendations []analysisv1alph1.Recommendation
	for _, recommendation := range recommendResult.Items {
		selected := true
		if !reflect.DeepEqual(o.ResourceSelector, recommendation.Spec.TargetRef) {
			selected = false
		}

		if selected {
			recommendations = append(recommendations, recommendation)
		}
	}

	o.renderTable(recommendations)

	return nil
}

func (o *ViewRecommendOptions) renderTable(recommendations []analysisv1alph1.Recommendation) {
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

func (o *ViewRecommendOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Selector, "selector", "", "", "Specify source selector")
}
