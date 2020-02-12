package main

import (
	"fmt"
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"
	"testing"
)

func TestQuantity(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	q, err := resource.ParseQuantity("1Ki")
	g.Expect(err).Should(gomega.BeNil())
	fmt.Println(q.Value())
}
