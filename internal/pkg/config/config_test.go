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

package config

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// We will just test added functionality, not everything Viper can do

var _ = ginkgo.Describe("config", func() {

	var path string
	var file string

	var c *Config

	ginkgo.BeforeSuite(func() {
		var err error
		path, err = ioutil.TempDir("", "testdata")
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(path).To(gomega.BeADirectory())

		// Add a suffix so we test creation of a path
		path = filepath.Join(path, "test")

		file = filepath.Join(path, "testconfig.yml")
	})

	ginkgo.AfterSuite(func() {
		err := os.RemoveAll(path)
		gomega.Expect(err).To(gomega.Succeed())
	})

	ginkgo.BeforeEach(func() {
		c = NewConfig()
		c.Path = path
		c.ConfigFile = file
		fillTestConfig(c)
	})

	ginkgo.AfterEach(func() {
		os.Remove(file) // Ignore error in case file didn't exist
		c = nil
	})

	ginkgo.It("should write and read a config", func() {
		// Write
		gomega.Expect(c.Write()).To(gomega.Succeed())
		gomega.Expect(file).To(gomega.BeARegularFile())

		// Check permission
		info, err := os.Stat(file)
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(info.Mode().Perm()).To(gomega.BeEquivalentTo(0600))

		// Read data back
		c2 := NewConfig()
		c2.ConfigFile = file
		gomega.Expect(c2.Read()).To(gomega.Succeed())

		// Verify
		gomega.Expect(c2.AllSettings()).To(gomega.Equal(c.AllSettings()))
	})

	ginkgo.It("should create and merge a new subconfig", func() {
		sub := c.GetSubConfig("subconfig")
		gomega.Expect(sub).ToNot(gomega.BeNil())

		sub.Set("subkey", "subvalue")
		sub.MergeToParent()

		gomega.Expect(c.GetString("subconfig.subkey")).To(gomega.Equal("subvalue"))
	})

	ginkgo.It("should create and merge an existing subconfig", func() {
		sub := c.GetSubConfig("sub")
		gomega.Expect(sub).ToNot(gomega.BeNil())

		// Overwrite existing value
		sub.Set("entry", "subvalue")
		sub.MergeToParent()

		gomega.Expect(c.GetString("sub.entry")).To(gomega.Equal("subvalue"))
	})

	ginkgo.It("should merge subconfigs on write", func() {
		sub := c.GetSubConfig("sub")
		gomega.Expect(sub).ToNot(gomega.BeNil())

		sub.Set("subkey", "subvalue")
		gomega.Expect(sub.Write()).To(gomega.Succeed())

		// Read and check
		c2 := NewConfig()
		c2.ConfigFile = file
		gomega.Expect(c2.Read()).To(gomega.Succeed())
		gomega.Expect(c2.GetSubConfig("sub").AllSettings()).To(gomega.Equal(sub.AllSettings()))
	})

	ginkgo.It("should succesfully unset values", func() {
		// First we write and read, so that we have values that aren't overridden by set
		gomega.Expect(c.Write()).To(gomega.Succeed())
		c2 := NewConfig()
		c2.ConfigFile = file
		gomega.Expect(c2.Read()).To(gomega.Succeed())

		gomega.Expect(c2.IsSet("main")).To(gomega.BeTrue())
		c2.Unset("main")
		gomega.Expect(c2.IsSet("main")).To(gomega.BeFalse())
	})

	ginkgo.It("should succesfully unset values in subconfig", func() {
		sub := c.GetSubConfig("sub")
		gomega.Expect(sub).ToNot(gomega.BeNil())

		sub.Unset("entry")
		sub.MergeToParent()
		gomega.Expect(c.IsSet("sub.entry")).To(gomega.BeFalse())
	})

	ginkgo.It("should not print tokens", func() {
		// We're checking log output
		buffer := new(bytes.Buffer)
		tmpLogger := log.Logger
		tmpLevel := zerolog.GlobalLevel()
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.Output(buffer)
		defer func() {
			log.Logger = tmpLogger
			zerolog.SetGlobalLevel(tmpLevel)
		}()

		c.Set("test.token", "secrettoken")
		c.Print()

		gomega.Expect(buffer.String()).ToNot(gomega.ContainSubstring("secrettoken"))
	})
})

func fillTestConfig(c *Config) {
	c.Set("main", true)
	c.Set("sub.entry", "string")
	c.Set("sub.entry2", 12345)
	c.Set("deeply.nested.entry", []interface{}{"line1", "line2", "line3"})
}
