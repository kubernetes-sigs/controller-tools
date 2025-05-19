package markers_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCRDMarkers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CRD Markers Suite")
}
