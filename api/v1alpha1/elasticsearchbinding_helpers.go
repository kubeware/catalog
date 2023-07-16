package v1alpha1

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kmapi "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/client-go/meta"
)

const (
	ElasticSuffix string = "elastic"
)

func (in *ElasticsearchBinding) GetSecretEngineName() string {
	return meta.NameWithSuffix(in.GetName(), ElasticSuffix+"-secret-engine")
}

func (in *ElasticsearchBinding) GetUserRoleName() string {
	return meta.NameWithSuffix(in.GetName(), ElasticSuffix+"-role")
}

func (in *ElasticsearchBinding) GetUserSecretAccessRequestName() string {
	return meta.NameWithSuffix(in.GetName(), ElasticSuffix+"-req")
}

func (o ElasticsearchBinding) NewObject() (*unstructured.Unstructured, error) {
	var obj unstructured.Unstructured
	obj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "kubedb.com",
		Version: "v1alpha2",
		Kind:    "Elasticsearch",
	})
	obj.SetNamespace(o.Spec.SourceRef.Namespace)
	obj.SetName(o.Spec.SourceRef.Name)
	return &obj, nil
}

func (obj *ElasticsearchBinding) GetPhase() AppPhase {
	conditions := obj.Status.Conditions
	if !obj.GetDeletionTimestamp().IsZero() {
		return AppPhaseTerminating
	}

	if !kmapi.IsConditionTrue(conditions, string(AppConditionTypeVaultReady)) ||
		!kmapi.IsConditionTrue(conditions, string(AppConditionTypeServiceAccountReady)) {
		return AppPhasePending
	}

	if IsAccessRequestExpired(conditions) {
		return AppPhaseExpired
	}

	if !IsEngineAPIResourcesConditionDetermined(conditions) {
		return AppPhaseInProgress
	}

	if !IsEngineAPIResourcesReady(conditions) {
		return AppPhaseFailed
	}

	return AppPhaseCurrent
}
