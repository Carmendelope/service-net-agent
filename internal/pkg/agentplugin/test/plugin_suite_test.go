/*
 * Copyright 2019 Nalej
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
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
	ginkgo.RunSpecs(t, "internal/pkg/agentplugin package suite")
}
