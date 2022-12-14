package cmd

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	analysisv1alph1 "github.com/gocrane/api/analysis/v1alpha1"
	"github.com/gocrane/kubectl-crane/pkg/cmd/options"
	"github.com/gocrane/kubectl-crane/pkg/cmd/recommend"
)

var (
	viewRecommendExample = `
# view all recommend result with kube-system namespace
%[1]s view-recommend --api-version apps/v1 --kind Deployment -n {namespace} {name}
`
)

type ViewRecommendOptions struct {
	CommonOptions *options.CommonOptions

	APIVersion string
	Kind       string

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
			if err := o.Validate(args); err != nil {
				klog.Infof(fmt.Sprintf("\nExample:\n"+viewRecommendExample, "kubectl-crane"))
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

func (o *ViewRecommendOptions) Validate(args []string) error {
	if err := o.CommonOptions.Validate(); err != nil {
		return err
	}

	if o.APIVersion == "" || o.Kind == "" || o.CommonOptions.ConfigFlags.Namespace == nil || len(args) == 0 {
		return errors.New("the recommender target is valid, please follow the guide `kubectl-crane view-recommend --api-version apps/v1 --kind Deployment -n {namespace} {name}`")
	}

	return nil
}

func (o *ViewRecommendOptions) Complete(cmd *cobra.Command, args []string) error {
	if err := o.CommonOptions.Complete(cmd, args); err != nil {
		return err
	}

	o.ResourceSelector = corev1.ObjectReference{
		APIVersion: o.APIVersion,
		Kind:       o.Kind,
		Namespace:  *o.CommonOptions.ConfigFlags.Namespace,
		Name:       args[0],
	}

	return nil
}

func (o *ViewRecommendOptions) Run() error {
	recommendResult, err := o.CommonOptions.CraneClient.AnalysisV1alpha1().Recommendations(*o.CommonOptions.ConfigFlags.Namespace).List(context.TODO(), metav1.ListOptions{})
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

	recommend.RenderTable(recommendations, o.CommonOptions.Out)

	return nil
}

func (o *ViewRecommendOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.APIVersion, "api-version", "", "", "Specify target api-version")
	cmd.Flags().StringVarP(&o.Kind, "kind", "", "", "Specify target kind")
}
