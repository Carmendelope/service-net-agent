/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package test

// We cannot use the ping plugin for testing in the main plugin package
// because that would lead to a cyclic import. It also makes no sense to
// write a second testing plugin that lives in the plugin package, so we
// just move the tests to a separate package.

import (
	"testing"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

func TestHandlerPackage(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "internal/pkg/plugin package suite")
}
