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

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/meta"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	craneExample = `
	# view view pod resource recommendation
	%[1]s crane pod 

	# view pod resource recommendations 
	%[1]s crane workload
`

	errNoContext = fmt.Errorf("no context is currently set, use %q to select a new one", "kubectl config use-context <context>")
)

// CraneOptions provides information required to update
// the current context on a user's KUBECONFIG
type CraneOptions struct {
	configFlags *genericclioptions.ConfigFlags
	genericclioptions.IOStreams

	restConfig *rest.Config
	restMapper meta.RESTMapper
}

// NewCraneOptions provides an instance of CraneOptions with default values
func NewCraneOptions(streams genericclioptions.IOStreams) *CraneOptions {
	return &CraneOptions{
		configFlags: genericclioptions.NewConfigFlags(true),
		IOStreams:   streams,
	}
}

// NewCmdCrane provides a cobra command wrapping CraneOptions
func NewCmdCrane(streams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "crane",
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			return fmt.Errorf("Please refer to usage:%s ", fmt.Sprintf(craneExample, "kubectl"))
		},
	}

	cmd.AddCommand(newCmdCranePod(streams))
	cmd.AddCommand(newCmdCraneWorkload(streams))

	return cmd
}

func (o *CraneOptions) Complete(cmd *cobra.Command, args []string) error {
	var err error
	o.restConfig, err = o.configFlags.ToRESTConfig()
	if err != nil {
		return err
	}

	o.restMapper, err = o.configFlags.ToRESTMapper()
	if err != nil {
		return err
	}

	return nil
}

// Validate ensures that all required arguments and flag values are provided
func (o *CraneOptions) Validate() error {
	return nil
}
