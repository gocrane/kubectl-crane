package cmd

import (
	"context"
	"github.com/gocrane/kubectl-crane/pkg/cmd/options"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	analysisv1alph1 "github.com/gocrane/api/analysis/v1alpha1"
)

type WorkloadMeta struct {
	ApiVersion string
	Kind       string
}

type CraneWorkloadOptions struct {
	CommonOptions *options.CommonOptions
	AllNamespaces bool
}

func NewCraneWorkloadOptions() *CraneWorkloadOptions {
	return &CraneWorkloadOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

func NewCmdCraneWorkload() *cobra.Command {
	o := NewCraneWorkloadOptions()

	cmd := &cobra.Command{
		Use:   "workload",
		Short: "view workload resource/replicas/hpa recommendations",
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}

			if err := o.Run(); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&o.AllNamespaces, "all-namespaces", "A", o.AllNamespaces, "If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.")
	o.CommonOptions.AddCommonFlag(cmd)

	return cmd
}

func (o *CraneWorkloadOptions) Validate() error {
	if err := o.CommonOptions.Validate(); err != nil {
		return err
	}

	return nil
}

func (o *CraneWorkloadOptions) Complete(cmd *cobra.Command, args []string) error {
	if err := o.CommonOptions.Complete(cmd, args); err != nil {
		return err
	}

	return nil
}

func (o *CraneWorkloadOptions) Run() error {
	namespace := "default"
	if len(*o.CommonOptions.ConfigFlags.Namespace) > 0 {
		namespace = *o.CommonOptions.ConfigFlags.Namespace
	}

	if o.AllNamespaces {
		namespace = ""
	}

	deploymentList, err := o.CommonOptions.KubeClient.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to get deployments, %v.", err)
		return err
	}

	/*	statefulsetList, err := kubeClient.AppsV1().StatefulSets(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Failed to get statefulsets, %v.", err)
			return err
		}*/

	recommendList, err := o.CommonOptions.CraneClient.AnalysisV1alpha1().Recommendations("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to get recommends, %v.", err)
		return err
	}

	recommendMap := map[string]analysisv1alph1.Recommendation{}
	for _, recommend := range recommendList.Items {
		recommendMap[GetObjectRefKey(string(recommend.Spec.Type), recommend.Spec.TargetRef)] = recommend
	}

	analyticsList, err := o.CommonOptions.CraneClient.AnalysisV1alpha1().Analytics("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to get analytics, %v.", err)
		return err
	}

	var workloadMetas []WorkloadMeta
	for _, analytics := range analyticsList.Items {
		for _, selector := range analytics.Spec.ResourceSelectors {
			if strings.ToLower(selector.Kind) != "deployment" &&
				strings.ToLower(selector.Kind) != "statefulSet" &&
				strings.ToLower(selector.Kind) != "ReplicaSet" {
				workloadMetas = append(workloadMetas, WorkloadMeta{
					Kind:       selector.Kind,
					ApiVersion: selector.APIVersion,
				})
			}
		}
	}

	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.SetOutputMirror(o.CommonOptions.Out)
	header := table.Row{}
	if o.AllNamespaces {
		header = append(header, "NAMESPACE")
	}
	header = append(header, table.Row{"NAME", "CONTAINER", "TYPE", "CPU", "MEMORY", "RECOMMEND CPU", "RECOMMEND MEMORY", "CPU DIFF", "MEMORY DIFF", "REPLICAS", "RECOMMEND REPLICAS", "REPLICAS DIFF"}...)
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

	for _, deployment := range deploymentList.Items {
		proposedRecommendation := GetProposedRecommendationsByMeta("Deployment", "apps/v1", deployment.Namespace, deployment.Name, recommendMap)

		for _, container := range deployment.Spec.Template.Spec.Containers {
			row := table.Row{}
			if o.AllNamespaces {
				row = append(row, deployment.Namespace)
			}
			row = append(row, deployment.Name)
			row = append(row, container.Name)
			row = append(row, "Deployment")

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
			if proposedRecommendation != nil && proposedRecommendation.ResourceRequest != nil {
				for _, recContainer := range proposedRecommendation.ResourceRequest.Containers {
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

			row = append(row, *deployment.Spec.Replicas)
			var recReplicas, replicasDiff int32
			if proposedRecommendation.ReplicasRecommendation != nil {
				recReplicas = *proposedRecommendation.ReplicasRecommendation.Replicas
				replicasDiff = *deployment.Spec.Replicas - recReplicas
			}
			row = append(row, PrintReplicas(recReplicas))
			row = append(row, PrintReplicas(replicasDiff))

			t.AppendRows([]table.Row{
				row,
			})
		}

		t.AppendSeparator()
	}

	t.AppendFooter(table.Row{"Total", "", "", PrintQuantity(cpuTotal), PrintQuantity(memoryTotal), PrintQuantity(recCpuTotal), PrintQuantity(recMemoryTotal), PrintQuantity(diffCpuTotal), PrintQuantity(diffMemoryTotal)})
	t.Render()

	return nil
}
