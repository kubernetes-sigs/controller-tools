//go:build !ignore_autogenerated
// Code generated by applyconfigurations. DO NOT EDIT.

package testapplyconfiguration

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	testing "k8s.io/client-go/testing"
	cronjob "sigs.k8s.io/controller-tools/pkg/applyconfigurations/testdata/cronjob"
	internal "sigs.k8s.io/controller-tools/pkg/applyconfigurations/testdata/cronjob/testapplyconfiguration/internal"
	testdatacronjob "sigs.k8s.io/controller-tools/pkg/applyconfigurations/testdata/cronjob/testapplyconfiguration/testdata/cronjob"
)

// ForKind returns an apply configuration type for the given GroupVersionKind, or nil if no
// apply configuration type exists for the given GroupVersionKind.
func ForKind(kind schema.GroupVersionKind) interface{} {
	switch kind {
	// Group=testdata, Version=cronjob
	case cronjob.SchemeGroupVersion.WithKind("AssociativeType"):
		return &testdatacronjob.AssociativeTypeApplyConfiguration{}
	case cronjob.SchemeGroupVersion.WithKind("CronJob"):
		return &testdatacronjob.CronJobApplyConfiguration{}
	case cronjob.SchemeGroupVersion.WithKind("CronJobSpec"):
		return &testdatacronjob.CronJobSpecApplyConfiguration{}
	case cronjob.SchemeGroupVersion.WithKind("CronJobStatus"):
		return &testdatacronjob.CronJobStatusApplyConfiguration{}
	case cronjob.SchemeGroupVersion.WithKind("ExampleStruct"):
		return &testdatacronjob.ExampleStructApplyConfiguration{}

	}
	return nil
}

func NewTypeConverter(scheme *runtime.Scheme) *testing.TypeConverter {
	return &testing.TypeConverter{Scheme: scheme, TypeResolver: internal.Parser()}
}
