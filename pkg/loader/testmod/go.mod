module sigs.k8s.io/controller-tools/pkg/loader/testmod

go 1.17

replace sigs.k8s.io/controller-tools/pkg/loader/testmod/submod1 => ./submod1

require sigs.k8s.io/controller-tools/pkg/loader/testmod/submod1 v0.0.0-00010101000000-000000000000 // indirect
