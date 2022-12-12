package recommend

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"

	"github.com/gocrane/kubectl-crane/pkg/cmd/options"
	"github.com/gocrane/kubectl-crane/pkg/utils"
)

var (
	recommendAdoptExample = `
# view all recommend result with kube-system namespace
%[1]s recommend adopt --name workloads-rule-resource-ntzns

# pre-commit
%[1]s recommend adopt --name workloads-rule-resource-ntzns --dry-run=All
`
)

type RecommendAdoptOptions struct {
	CommonOptions *options.CommonOptions

	DryRun string
	Name   string
}

func NewRecommendAdoptOptions() *RecommendAdoptOptions {
	return &RecommendAdoptOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

func NewCmdRecommendAdopt() *cobra.Command {
	o := NewRecommendAdoptOptions()

	command := &cobra.Command{
		Use:     "adopt",
		Short:   "Adopt a recommend to resource",
		Example: fmt.Sprintf(recommendListExample, "kubectl-crane"),
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				klog.Infof(fmt.Sprintf("\nExample:\n"+recommendAdoptExample, "kubectl-crane"))
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

func (o *RecommendAdoptOptions) Validate() error {
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

func (o *RecommendAdoptOptions) Complete(cmd *cobra.Command, args []string) error {
	if err := o.CommonOptions.Complete(cmd, args); err != nil {
		return err
	}

	return nil
}

func (o *RecommendAdoptOptions) Run() error {
	recommend, err := o.CommonOptions.CraneClient.AnalysisV1alpha1().Recommendations(*o.CommonOptions.ConfigFlags.Namespace).Get(context.TODO(), o.Name, metav1.GetOptions{})
	if err != nil {
		return errors.New("the recommend doesn't exist, please specify a existed recommend name with --name")
	}

	if string(recommend.Spec.Type) == "Replicas" ||
		string(recommend.Spec.Type) == "Resource" {
		gvr, err := utils.GetGroupVersionResource(o.CommonOptions.DiscoveryClient, recommend.Spec.TargetRef.APIVersion, recommend.Spec.TargetRef.Kind)
		if err != nil {
			return errors.New(fmt.Sprintf("Recommendation type %s is not supported for adoption ", string(recommend.Spec.Type)))
		}

		patchOptions := metav1.PatchOptions{}
		if len(o.DryRun) != 0 {
			patchOptions.DryRun = []string{"All"}
		}

		patched, err := o.CommonOptions.DynamicClient.Resource(*gvr).Namespace(recommend.Spec.TargetRef.Namespace).Patch(context.TODO(), recommend.Spec.TargetRef.Name, types.StrategicMergePatchType, []byte(recommend.Status.RecommendedInfo), patchOptions)
		if err != nil {
			return errors.New(fmt.Sprintf("adopt the recommend failed because %v", err))
		}

		// when dry-run set, print the object
		if len(o.DryRun) != 0 {
			printer := printers.NewTypeSetter(scheme.Scheme).ToPrinter(&printers.YAMLPrinter{})
			if err = printer.PrintObj(patched, o.CommonOptions.Out); err != nil {
				return err
			}
		}

	} else {
		return errors.New(fmt.Sprintf("Recommendation type %s is not supported for adoption ", string(recommend.Spec.Type)))
	}

	return nil
}

func (o *RecommendAdoptOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Name, "name", "", "", "Specify the name for recommend")
	cmd.Flags().StringVarP(&o.DryRun, "dry-run", "", "", "Pre-commit")
}
