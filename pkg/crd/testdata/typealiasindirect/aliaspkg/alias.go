package aliaspkg

import "testdata.kubebuilder.io/cronjob/typealiasindirect/targetpkg"

type MyAlias = targetpkg.TargetStruct
