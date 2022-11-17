/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"github.com/gocrane/kubectl-crane/pkg/cmd/options"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var (
	craneExample = `
# view pod resource recommendation
%[1]s pod

# view pod resource recommendations
%[1]s workload
`

	errNoContext = fmt.Errorf("no context is currently set, use %q to select a new one", "kubectl config use-context <context>")
)

// CraneOptions provides information required to update
// the current context on a user's KUBECONFIG
type CraneOptions struct {
	commonOptions *options.CommonOptions
}

var defaultConfigFlags = genericclioptions.NewConfigFlags(true)

func NewCraneOptions() *CraneOptions {
	return &CraneOptions{&options.CommonOptions{
		ConfigFlags: defaultConfigFlags,
		IOStreams:   genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr},
	}}
}

// NewCraneCommand creates the `kubectl-crane` command
func NewCraneCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "kubectl-crane",
		SilenceUsage: true,
		Short:        "Kubectl plugin for crane, including recommendation and cost estimate.",
		Example:      fmt.Sprintf(craneExample, "kubectl-crane"),
	}

	cmd.AddCommand(NewCmdCranePod())
	cmd.AddCommand(NewCmdCraneWorkload())
	cmd.AddCommand(NewCmdRecommendationRule())
	cmd.AddCommand(NewCmdRecommend())
	cmd.AddCommand(NewCmdViewRecommend())
	cmd.AddCommand(NewCmdVersion())

	return cmd
}
