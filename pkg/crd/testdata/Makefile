GEN?=../../../.run-controller-gen.sh output:stdout

all:
	# Generate separate CRDs for cronjob v1 and v2
	# - using crd v1
	$(GEN) crd:crdVersions=v1 paths=./cronjob/v1 > cronjob/v1/cronjob_crdv1.yaml
	$(GEN) crd:crdVersions=v1 paths=./cronjob/v2 > cronjob/v2/cronjob_crdv1.yaml
	# - using crd v1beta1
	$(GEN) crd:crdVersions=v1beta1 paths=./cronjob/v1 > cronjob/v1/cronjob_crdv1beta1.yaml
	$(GEN) crd:crdVersions=v1beta1 paths=./cronjob/v2 > cronjob/v2/cronjob_crdv1beta1.yaml
	# Generate a combined CRD of Cronjob v1 & v2
	# - using crd v1
	$(GEN) crd:crdVersions=v1 paths=./cronjob/... > cronjob/cronjob_crdv1.yaml
	# -- with maxDescLen=10
	$(GEN) crd:crdVersions=v1,maxDescLen=10 paths=./cronjob/... > cronjob/cronjob_crdv1_maxdesclen.yaml
	# - using crd v1beta1
	$(GEN) crd:crdVersions=v1beta1 paths=./cronjob/... > cronjob/cronjob_crdv1beta1.yaml
	# -- with maxDescLen=10
	$(GEN) crd:crdVersions=v1beta1,maxDescLen=10 paths=./cronjob/... > cronjob/cronjob_crdv1beta1_maxdesclen.yaml
	# -- with trivialVersions=true
	$(GEN) crd:crdVersions=v1beta1,trivialVersions=true paths=./cronjob/... > cronjob/cronjob_crdv1beta1_trivial.yaml
	# -- with preserveUnknownFields=false
	$(GEN) crd:crdVersions=v1beta1,preserveUnknownFields=false paths=./cronjob/... > cronjob/cronjob_crdv1beta1_preserve_false.yaml
	# - using crd v1 and crd v1beta1
	$(GEN) "crd:crdVersions=v1;v1beta1" paths=./cronjob/... > cronjob/cronjob_crdv1v1beta1.yaml

.PHONY: all