package cmd

import (
	autoscalingv2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"

	autoscalingapi "github.com/gocrane/api/autoscaling/v1alpha1"
)

// ProposedRecommendation is the result for one recommendation
type ProposedRecommendation struct {
	// EffectiveHPA is the proposed recommendation for type Replicas
	EffectiveHPA *EffectiveHorizontalPodAutoscalerRecommendation `json:"effectiveHPA,omitempty"`

	// ReplicasRecommendation is the proposed replicas for type Replicas
	ReplicasRecommendation *ReplicasRecommendation `json:"replicasRecommendation,omitempty"`

	// ResourceRequest is the proposed recommendation for type Resource
	ResourceRequest *ResourceRequestRecommendation `json:"resourceRequest,omitempty"`
}

type ReplicasRecommendation struct {
	Replicas *int32 `json:"replicas,omitempty"`
}

type EffectiveHorizontalPodAutoscalerRecommendation struct {
	MinReplicas *int32                     `json:"minReplicas,omitempty"`
	MaxReplicas *int32                     `json:"maxReplicas,omitempty"`
	Metrics     []autoscalingv2.MetricSpec `json:"metrics,omitempty"`
	Prediction  *autoscalingapi.Prediction `json:"prediction,omitempty"`
}

type ResourceRequestRecommendation struct {
	Containers []ContainerRecommendation `json:"containers,omitempty"`
}

type ContainerRecommendation struct {
	ContainerName string       `json:"containerName,omitempty"`
	Target        ResourceList `json:"target,omitempty"`
}

type ResourceList map[corev1.ResourceName]string
