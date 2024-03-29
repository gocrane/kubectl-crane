package cmd

import (
	"fmt"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	analysisv1alph1 "github.com/gocrane/api/analysis/v1alpha1"
)

func GetResourceRequestRecommendationsByPod(pod corev1.Pod, recommendMap map[string]analysisv1alph1.Recommendation) *analysisv1alph1.ResourceRequestRecommendation {
	for _, ref := range pod.OwnerReferences {
		key := GetOwnerKey(analysisv1alph1.ResourceRecommender, ref, pod.Namespace)
		if recommend, exist := recommendMap[key]; exist {
			if recommend.Status.RecommendedValue == "" {
				continue
			}
			var proposedRecommendation analysisv1alph1.ProposedRecommendation
			err := yaml.Unmarshal([]byte(recommend.Status.RecommendedValue), &proposedRecommendation)
			if err != nil {
				return nil
			}

			return proposedRecommendation.ResourceRequest
		}
	}

	return nil
}

func GetProposedRecommendationsByMeta(kind, apiVersion, namespace, name string, recommendMap map[string]analysisv1alph1.Recommendation) *analysisv1alph1.ProposedRecommendation {
	var recommendation analysisv1alph1.ProposedRecommendation

	resourceKey := GetObjectKey(analysisv1alph1.ResourceRecommender, kind, apiVersion, namespace, name)

	if recommend, exist := recommendMap[resourceKey]; exist {
		var proposedRecommendation analysisv1alph1.ProposedRecommendation
		yaml.Unmarshal([]byte(recommend.Status.RecommendedValue), &proposedRecommendation)

		recommendation.ResourceRequest = proposedRecommendation.ResourceRequest
	}

	replicasKey := GetObjectKey(analysisv1alph1.ReplicasRecommender, kind, apiVersion, namespace, name)

	if recommend, exist := recommendMap[replicasKey]; exist {
		var proposedRecommendation analysisv1alph1.ProposedRecommendation
		yaml.Unmarshal([]byte(recommend.Status.RecommendedValue), &proposedRecommendation)

		recommendation.ReplicasRecommendation = proposedRecommendation.ReplicasRecommendation
		recommendation.EffectiveHPA = proposedRecommendation.EffectiveHPA
	}

	return &recommendation
}

func GetOwnerKey(recType string, ownRef metav1.OwnerReference, namespace string) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s", recType, ownRef.Kind, ownRef.APIVersion, namespace, ownRef.Name)
}

func GetObjectKey(recType, kind, apiVersion, namespace, name string) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s", recType, kind, apiVersion, namespace, name)
}

func GetObjectRefKey(recType string, objectRef corev1.ObjectReference) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s", recType, objectRef.Kind, objectRef.APIVersion, objectRef.Namespace, objectRef.Name)
}

func PrintQuantity(quantity *resource.Quantity) string {
	if quantity != nil && !quantity.IsZero() {
		return quantity.String()
	}

	return ""
}

func PrintReplicas(replicas int32) string {
	if replicas == 0 {
		return ""
	}

	return strconv.Itoa(int(replicas))
}
