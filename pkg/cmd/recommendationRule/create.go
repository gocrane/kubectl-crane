package recommendationRule

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"

	"github.com/gocrane/api/analysis/v1alpha1"

	"github.com/gocrane/kubectl-crane/pkg/cmd/options"
)

var (
	recommendationRuleCreateExample = `
# create a simple recommendation rule for kube-system namespace
%[1]s rr create --namespace kube-system --target '[{"kind": "Deployment", "apiVersion": "apps/v1"}]' --run-interval 4h

# create a simple recommendation rule for all namespace
%[1]s rr create --target '[{"kind": "Deployment", "apiVersion": "apps/v1"}]' --run-interval 4h

# pre-commit
%[1]s rr create --target '[{"kind": "Deployment", "apiVersion": "apps/v1"}]' --run-interval 4h --dry-run=All

# create a simple recommendation rule for all namespace with Any and Resource\Replicas recommender
%[1]s rr create --namespace Any --recommender Resource,Replicas --target '[{"kind": "Deployment", "apiVersion": "apps/v1"}]' --run-interval 4h
`

	recommenderMap = map[string]int{"Replicas": 1, "Resource": 2, "IdleNode": 3}
)

type RecommendationRuleCreateOptions struct {
	CommonOptions *options.CommonOptions

	Recommender string
	Target      string
	RunInterval string
	DryRun      string

	ResourceSelectors []v1alpha1.ResourceSelector
}

func NewRecommendationRuleCreateOptions() *RecommendationRuleCreateOptions {
	return &RecommendationRuleCreateOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

func NewCmdRecommendationRuleCreate() *cobra.Command {
	o := NewRecommendationRuleCreateOptions()

	command := &cobra.Command{
		Use:     "create",
		Short:   "create a simple recommendation rules",
		Example: fmt.Sprintf(recommendationRuleCreateExample, "kubectl-crane"),
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				klog.Infof(fmt.Sprintf("\nExample:\n"+recommendationRuleCreateExample, "kubectl-crane"))
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

func (o *RecommendationRuleCreateOptions) Validate() error {
	if err := o.CommonOptions.Validate(); err != nil {
		return err
	}

	err := json.Unmarshal([]byte(o.Target), &o.ResourceSelectors)
	if err != nil {
		return errors.New("please check the recommender target is valid")
	}

	recommenders := strings.Split(o.Recommender, ",")
	for _, recommender := range recommenders {
		if _, ok := recommenderMap[recommender]; !ok {
			return errors.New("the recommender only support Replicas,Resource and IdleNode")
		}
	}

	if len(o.RunInterval) == 0 {
		return errors.New("please specify the runInterval with --runInterval")
	}

	return nil
}

func (o *RecommendationRuleCreateOptions) Complete(cmd *cobra.Command, args []string) error {
	if err := o.CommonOptions.Complete(cmd, args); err != nil {
		return err
	}

	return nil
}

func (o *RecommendationRuleCreateOptions) Run() error {
	recommendationRule := &v1alpha1.RecommendationRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RecommendationRule",
			APIVersion: "analysis.crane.io/v1alpha1",
		},
	}
	recommendationRule.Namespace = ""

	name := "recommendation"

	if len(*o.CommonOptions.ConfigFlags.Namespace) == 0 || strings.EqualFold(*o.CommonOptions.ConfigFlags.Namespace, "Any") {
		recommendationRule.Spec.NamespaceSelector.Any = true
		name += "-Any"
	} else {
		recommendationRule.Spec.NamespaceSelector.MatchNames = strings.Split(*o.CommonOptions.ConfigFlags.Namespace, ",")
	}

	recommendationRule.Spec.ResourceSelectors = o.ResourceSelectors
	name += "-" + o.ResourceSelectors[0].Kind

	recommenders := strings.Split(o.Recommender, ",")
	for _, recommender := range recommenders {
		name += "-" + recommender
		recommendationRule.Spec.Recommenders = append(recommendationRule.Spec.Recommenders, v1alpha1.Recommender{
			Name: recommender,
		})
	}

	recommendationRule.Spec.RunInterval = o.RunInterval
	name += "-" + o.RunInterval
	name += "-" + time.Now().Format("2006-01-02")

	recommendationRule.Name = name

	createOptions := metav1.CreateOptions{}
	if len(o.DryRun) != 0 {
		createOptions.DryRun = []string{"All"}
	}

	created, err := o.CommonOptions.CraneClient.AnalysisV1alpha1().RecommendationRules("default").Create(context.Background(), recommendationRule, createOptions)

	// when dry-run set, print the object
	if len(o.DryRun) != 0 {
		created.Kind = "RecommendationRule"
		created.APIVersion = "analysis.crane.io/v1alpha1"
		printer := printers.NewTypeSetter(scheme.Scheme).ToPrinter(&printers.YAMLPrinter{})
		if err = printer.PrintObj(created, o.CommonOptions.Out); err != nil {
			return err
		}
	}

	klog.Infof(fmt.Sprintf("the recommendation rule %s created successfully", name))
	return nil
}

func (o *RecommendationRuleCreateOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Recommender, "recommender", "", "Resource", "specify type for recommendationrules，separated with ',' if more than one, default is Resource")
	cmd.Flags().StringVarP(&o.Target, "target", "", "", "specify recommend target for recommendationrules")
	cmd.Flags().StringVarP(&o.RunInterval, "run-interval", "", "", "Specify runInterval for recommendationrules")
	cmd.Flags().StringVarP(&o.DryRun, "dry-run", "", "", "pre-commit")
}
