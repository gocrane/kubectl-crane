package recommend

import (
	"context"
	"errors"
	"fmt"

	"github.com/gocrane/kubectl-crane/pkg/cmd/options"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
)

var (
	triggerRecommendExample = `
# manually trigger the specified recommendation rule
%[1]s recommend trigger --name workloads-ntzns -n default

# pre-commit
%[1]s recommend trigger --name workloads-ntzns -n default --dry-run
`
)

type RecommendTriggerOptions struct {
	CommonOptions *options.CommonOptions

	DryRun bool
	Name   string
}

func NewRecommendTriggerOptions() *RecommendTriggerOptions {
	return &RecommendTriggerOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

func NewCmdRecommendTrigger() *cobra.Command {
	o := NewRecommendTriggerOptions()

	command := &cobra.Command{
		Use:     "trigger",
		Short:   "Manually triggering a recommendation",
		Example: fmt.Sprintf(triggerRecommendExample, "kubectl-crane"),
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				klog.Infof(fmt.Sprintf("\nExample:\n"+triggerRecommendExample, "kubectl-crane"))
				return err
			}

			if err := o.Run(); err != nil {
				return err
			}

			return nil
		},
	}

	o.AddFlags(command)
	o.CommonOptions.AddCommonFlag(command)

	return command
}

func (o *RecommendTriggerOptions) Validate() error {
	if err := o.CommonOptions.Validate(); err != nil {
		return err
	}

	if len(o.Name) == 0 {
		return errors.New("please specify the recommend name")
	}

	if len(*o.CommonOptions.ConfigFlags.Namespace) == 0 {
		return errors.New("please specify the recommend namespace")
	}

	return nil
}

func (o *RecommendTriggerOptions) Complete(cmd *cobra.Command, args []string) error {
	if err := o.CommonOptions.Complete(cmd, args); err != nil {
		return err
	}

	return nil
}

func (o *RecommendTriggerOptions) Run() error {
	recommend, err := o.CommonOptions.CraneClient.AnalysisV1alpha1().Recommendations(*o.CommonOptions.ConfigFlags.Namespace).Get(context.TODO(), o.Name, metav1.GetOptions{})
	if err != nil {
		return errors.New("the recommend doesn't exist, please specify a existed recommend name with --name")
	}

	// Modify the run number annotation for the recommendation that should be triggered.
	// The run number should be lower than the runNumber in the recommendationRule.
	// Just set the run number annotation to zero
	if recommend.Annotations == nil {
		recommend.Annotations = make(map[string]string, 0)
	}
	recommend.Annotations[RunNumberAnnotation] = "0"
	updateOptions := metav1.UpdateOptions{}
	if o.DryRun {
		updateOptions.DryRun = []string{"All"}
	}
	if recommend, err = o.CommonOptions.CraneClient.AnalysisV1alpha1().Recommendations(*o.CommonOptions.ConfigFlags.Namespace).Update(context.TODO(), recommend, updateOptions); err != nil {
		return fmt.Errorf("failed to trigger the recommendation %s, %v", recommend.Name, err)
	}

	// when dry-run set, print the object
	if o.DryRun {
		recommend.Kind = "Recommendation"
		recommend.APIVersion = "analysis.crane.io/v1alpha1"
		printer := printers.NewTypeSetter(scheme.Scheme).ToPrinter(&printers.YAMLPrinter{})
		if err = printer.PrintObj(recommend, o.CommonOptions.Out); err != nil {
			return err
		}

		return nil
	}

	klog.Infof(fmt.Sprintf("success to trigger the recommendation %s", o.Name))
	return nil
}

func (o *RecommendTriggerOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Name, "name", "", "", "Specify the name for recommend")
	cmd.Flags().BoolVarP(&o.DryRun, "dry-run", "", false, "dry-run")
}
