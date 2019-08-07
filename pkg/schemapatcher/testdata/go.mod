module testdata.kubebuilder.io/inplaceschema

go 1.12

require (
	k8s.io/apimachinery v0.0.0-20190806215851-162a2dabc72f
	sigs.k8s.io/controller-tools v0.0.0-00010101000000-000000000000 // indirect
)

// use the current copy of controller-tools
replace sigs.k8s.io/controller-tools => ../../..
