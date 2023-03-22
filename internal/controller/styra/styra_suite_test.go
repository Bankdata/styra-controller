package styra_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestStyra(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "internal/controller/styra")
}
