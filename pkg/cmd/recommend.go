package cmd

import (
	"github.com/gocrane/kubectl-crane/pkg/cmd/options"
	"github.com/gocrane/kubectl-crane/pkg/cmd/recommend"
	"github.com/spf13/cobra"
)

type RecommendOptions struct {
	CommonOptions *options.CommonOptions
}

func NewRecommendOptions() *RecommendOptions {
	return &RecommendOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

func NewCmdRecommend() *cobra.Command {
	recommendOptions := NewRecommendOptions()

	cmd := &cobra.Command{
		Use:   "recommend",
		Short: "view or adopt recommend result",
	}
	recommendOptions.CommonOptions.AddCommonFlag(cmd)

	cmd.AddCommand(recommend.NewCmdRecommendList())
	cmd.AddCommand(recommend.NewCmdRecommendAdopt())

	return cmd
}
