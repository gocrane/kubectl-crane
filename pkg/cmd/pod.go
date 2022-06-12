package cmd

import (
	"context"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	analysisv1alph1 "github.com/gocrane/api/analysis/v1alpha1"
	craneclientset "github.com/gocrane/api/pkg/generated/clientset/versioned"
)

type CraneOptionsPod struct {
	CraneOptions
	AllNamespaces bool
}

func newCmdCranePod(streams genericclioptions.IOStreams) *cobra.Command {
	craneOptions := NewCraneOptions(streams)
	podOptions := &CraneOptionsPod{
		CraneOptions: *craneOptions,
	}

	cmd := &cobra.Command{
		Use:   "pod",
		Short: "view pod resource recommendations",
		RunE: func(c *cobra.Command, args []string) error {
			if err := podOptions.Complete(c, args); err != nil {
				return err
			}
			if err := podOptions.Validate(); err != nil {
				return err
			}

			if err := podOptions.Run(); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&podOptions.AllNamespaces, "all-namespaces", "A", podOptions.AllNamespaces, "If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.")
	podOptions.configFlags.AddFlags(cmd.Flags())

	return cmd
}

func (o *CraneOptionsPod) Validate() error {
	if err := o.CraneOptions.Validate(); err != nil {
		return err
	}

	return nil
}

func (o *CraneOptionsPod) Complete(cmd *cobra.Command, args []string) error {
	if err := o.CraneOptions.Complete(cmd, args); err != nil {
		return err
	}

	return nil
}

func (o *CraneOptionsPod) Run() error {
	kubeClient, err := kubernetes.NewForConfig(o.restConfig)
	if err != nil {
		klog.Errorf("Failed to new kubernetes client, %v.", err)
		return err
	}

	craneClient, err := craneclientset.NewForConfig(o.restConfig)
	if err != nil {
		klog.Errorf("Failed to new crane client, %v.", err)
		return err
	}

	namespace := "default"
	if len(*o.configFlags.Namespace) > 0 {
		namespace = *o.configFlags.Namespace
	}

	if o.AllNamespaces {
		namespace = ""
	}

	podList, err := kubeClient.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to get pods, %v.", err)
		return err
	}

	recommendList, err := craneClient.AnalysisV1alpha1().Recommendations("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to get recommends, %v.", err)
		return err
	}

	recommendMap := map[string]analysisv1alph1.Recommendation{}
	for _, recommend := range recommendList.Items {
		recommendMap[GetObjectRefKey(string(recommend.Spec.Type), recommend.Spec.TargetRef)] = recommend
	}

	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.SetOutputMirror(o.Out)
	header := table.Row{}
	if o.AllNamespaces {
		header = append(header, "NAMESPACE")
	}
	header = append(header, table.Row{"NAME", "CONTAINER", "CPU", "MEMORY", "RECOMMEND CPU", "RECOMMEND MEMORY", "CPU DIFF", "MEMORY DIFF"}...)
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
		{
			Name:        "CONTAINER",
			Align:       text.AlignLeft,
			AlignFooter: text.AlignLeft,
			AlignHeader: text.AlignLeft,
			VAlign:      text.VAlignMiddle,
			WidthMin:    6,
			WidthMax:    24,
		},
	})
	cpuTotal := resource.NewQuantity(0, resource.DecimalSI)
	memoryTotal := resource.NewQuantity(0, resource.BinarySI)
	recCpuTotal := resource.NewQuantity(0, resource.DecimalSI)
	recMemoryTotal := resource.NewQuantity(0, resource.BinarySI)
	diffCpuTotal := resource.NewQuantity(0, resource.DecimalSI)
	diffMemoryTotal := resource.NewQuantity(0, resource.BinarySI)

	for _, pod := range podList.Items {
		resourceRecommendation := GetResourceRequestRecommendationsByPod(pod, recommendMap)

		for _, container := range pod.Spec.Containers {
			row := table.Row{}
			if o.AllNamespaces {
				row = append(row, pod.Namespace)
			}
			row = append(row, pod.Name)
			row = append(row, container.Name)
			containerCpu := resource.NewQuantity(0, resource.DecimalSI)
			containerMemory := resource.NewQuantity(0, resource.BinarySI)
			if container.Resources.Requests != nil {
				requestCpu := container.Resources.Requests[corev1.ResourceCPU]
				if !requestCpu.IsZero() {
					containerCpu = &requestCpu
				}
				requestMemory := container.Resources.Requests[corev1.ResourceMemory]
				if !requestMemory.IsZero() {
					containerMemory = &requestMemory
				}
			}
			row = append(row, PrintQuantity(containerCpu))
			row = append(row, PrintQuantity(containerMemory))
			cpuTotal.Add(*containerCpu)
			memoryTotal.Add(*containerMemory)

			recCpu := resource.NewQuantity(0, resource.DecimalSI)
			recMemory := resource.NewQuantity(0, resource.BinarySI)
			if resourceRecommendation != nil && resourceRecommendation != nil {
				for _, recContainer := range resourceRecommendation.Containers {
					if recContainer.ContainerName == container.Name {
						recCpuResource, err := resource.ParseQuantity(recContainer.Target[corev1.ResourceCPU])
						if err == nil {
							recCpu = &recCpuResource
						}
						recMemoryResource, err := resource.ParseQuantity(recContainer.Target[corev1.ResourceMemory])
						if err == nil {
							recMemory = &recMemoryResource
						}
					}
				}
			}
			row = append(row, PrintQuantity(recCpu))
			row = append(row, PrintQuantity(recMemory))
			recCpuTotal.Add(*recCpu)
			recMemoryTotal.Add(*recMemory)

			containerCpuDiff := resource.NewQuantity(0, resource.DecimalSI)
			if !recCpu.IsZero() {
				containerCpuDiff = containerCpu
				containerCpuDiff.Sub(*recCpu)
			}

			containerMemoryDiff := resource.NewQuantity(0, resource.BinarySI)
			if !recMemory.IsZero() {
				containerMemoryDiff = containerMemory
				containerMemoryDiff.Sub(*recMemory)
			}

			row = append(row, PrintQuantity(containerCpuDiff))
			row = append(row, PrintQuantity(containerMemoryDiff))
			diffCpuTotal.Add(*containerCpuDiff)
			diffMemoryTotal.Add(*containerMemoryDiff)

			t.AppendRows([]table.Row{
				row,
			})
		}

		t.AppendSeparator()
	}

	t.AppendFooter(table.Row{"Total", "", PrintQuantity(cpuTotal), PrintQuantity(memoryTotal), PrintQuantity(recCpuTotal), PrintQuantity(recMemoryTotal), PrintQuantity(diffCpuTotal), PrintQuantity(diffMemoryTotal)})
	t.Render()

	return nil
}
