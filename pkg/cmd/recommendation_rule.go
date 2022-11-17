package cmd

import (
	"github.com/gocrane/kubectl-crane/pkg/cmd/options"
	"github.com/gocrane/kubectl-crane/pkg/cmd/recommendationRule"
	"github.com/spf13/cobra"
)

type RecommendationRuleOptions struct {
	CommonOptions *options.CommonOptions
}

func NewRecommendationRuleOptions() *RecommendationRuleOptions {
	return &RecommendationRuleOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

func NewCmdRecommendationRule() *cobra.Command {
	recommendationRuleOptions := NewRecommendationRuleOptions()

	cmd := &cobra.Command{
		Use:     "recommendationrule",
		Aliases: []string{"rr"},
		Short:   "manage recommendation rules",
	}
	recommendationRuleOptions.CommonOptions.AddCommonFlag(cmd)

	cmd.AddCommand(recommendationRule.NewCmdRecommendationRuleList())
	cmd.AddCommand(recommendationRule.NewCmdRecommendationRuleCreate())

	return cmd
}
