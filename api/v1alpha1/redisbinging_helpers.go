/*
Copyright 2023.

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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kmapi "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/client-go/meta"
)

const (
	RedisSuffix string = "redis"
)

func (in *RedisBinding) GetSecretEngineName() string {
	return meta.NameWithSuffix(in.GetName(), RedisSuffix+"-secret-engine")
}

func (in *RedisBinding) GetUserRoleName() string {
	return meta.NameWithSuffix(in.GetName(), RedisSuffix+"-role")
}

func (in *RedisBinding) GetUserSecretAccessRequestName() string {
	return meta.NameWithSuffix(in.GetName(), RedisSuffix+"-req")
}

func (o RedisBinding) NewObject() (*unstructured.Unstructured, error) {
	var obj unstructured.Unstructured
	obj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "kubedb.com",
		Version: "v1alpha2",
		Kind:    "Redis",
	})
	obj.SetNamespace(o.Spec.SourceRef.Namespace)
	obj.SetName(o.Spec.SourceRef.Name)
	return &obj, nil
}

func GetPhaseForRedis(obj *RedisBinding) AppPhase {
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
