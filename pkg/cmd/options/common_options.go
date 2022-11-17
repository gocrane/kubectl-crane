package options

import (
	crane "github.com/gocrane/api/pkg/generated/clientset/versioned"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"os"
)

// CommonOptions provides information required to update
// the current context on a user's KUBECONFIG
type CommonOptions struct {
	ConfigFlags *genericclioptions.ConfigFlags
	genericclioptions.IOStreams

	RestConfig *rest.Config
	RestMapper meta.RESTMapper

	KubeClient  *kubernetes.Clientset
	CraneClient *crane.Clientset
}

var defaultConfigFlags = genericclioptions.NewConfigFlags(true)

func NewCommonOptions() *CommonOptions {
	return &CommonOptions{
		ConfigFlags: defaultConfigFlags,
		IOStreams:   genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr},
	}
}

func (o *CommonOptions) Complete(cmd *cobra.Command, args []string) error {
	var err error
	o.RestConfig, err = o.ConfigFlags.ToRESTConfig()
	if err != nil {
		return err
	}

	o.RestMapper, err = o.ConfigFlags.ToRESTMapper()
	if err != nil {
		return err
	}

	o.KubeClient, err = kubernetes.NewForConfig(o.RestConfig)
	if err != nil {
		klog.Errorf("Failed to new kubernetes client, %v.", err)
		return err
	}

	o.CraneClient, err = crane.NewForConfig(o.RestConfig)
	if err != nil {
		klog.Errorf("Failed to new crane client, %v.", err)
		return err
	}

	return nil
}

// Validate ensures that all required arguments and flag values are provided
func (o *CommonOptions) Validate() error {
	return nil
}

func (o *CommonOptions) AddCommonFlag(cmd *cobra.Command) {
	o.ConfigFlags.AddFlags(cmd.Flags())
}
