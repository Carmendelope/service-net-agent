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

// Agent configuration file management

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/nalej/derrors"

	"github.com/nalej/service-net-agent/version"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Config struct {
	Path       string
	ConfigFile string
	LogFile    string

	// For plugins - calling write on a subtree will call write on parent
	parent *Config
	// Key under which this child sits at parent
	childKey string

	// Write lock
	writeLock sync.Mutex

	*viper.Viper
}

func NewConfig() *Config {
	c := &Config{
		Viper: viper.New(),
	}

	// We store a token, make it only readable for user
	// See comment in Write()
	// c.Viper.SetConfigPermissions(0600)

	return c
}

func (c *Config) Read() derrors.Error {
	log.Info().Str("file", c.ConfigFile).Msg("reading configuration file")

	// Pass filename to Viper
	c.SetConfigFile(c.ConfigFile)

	// Check if file exists. Ok if not, just don't try to read.
	if _, err := os.Stat(c.ConfigFile); os.IsNotExist(err) {
		return nil
	}

	// Read
	err := c.ReadInConfig()
	if err != nil {
		return derrors.NewInvalidArgumentError("failed reading configuration file", err)
	}

	return nil
}

func (c *Config) GetSubConfig(prefix string) *Config {
	sub := c.Sub(prefix)
	if sub == nil {
		sub = viper.New()
	}

	config := &Config{
		parent:   c,
		childKey: prefix,
		Viper:    sub,
	}

	return config
}

func (c *Config) MergeToParent() {
	if c.parent == nil {
		return
	}

	c.parent.ReplaceSubtree(c.childKey, c.Viper)
}

func (c *Config) ReplaceSubtree(prefix string, config *viper.Viper) {
	c.writeLock.Lock()
	defer c.writeLock.Unlock()

	c.unsetLocked(prefix)
	c.mergeSubtreeLocked(prefix, config)
}

func (c *Config) mergeSubtreeLocked(prefix string, config *viper.Viper) {
	for k, v := range config.AllSettings() {
		c.Set(fmt.Sprintf("%s.%s", prefix, k), v)
	}
}

func (c *Config) Unset(key string) {
	c.writeLock.Lock()
	defer c.writeLock.Unlock()

	c.unsetLocked(key)
}

func (c *Config) DeleteConfigFile() derrors.Error {
	// Don't want to delete mid-write
	c.writeLock.Lock()
	defer c.writeLock.Unlock()

	log.Debug().Str("file", c.ConfigFile).Msg("removing config file")

	// Just checking if our code doesn't do weird things
	if c.parent != nil {
		return derrors.NewInvalidArgumentError("can't delete config file for a sub-config")
	}

	// Check if file exists. Ok if not, just don't try to delete.
	if _, err := os.Stat(c.ConfigFile); os.IsNotExist(err) {
		return nil
	}

	err := os.Remove(c.ConfigFile)
	if err != nil {
		return derrors.NewInternalError("failed to delete config file", err).WithParams(c.ConfigFile)
	}

	return nil
}

func (c *Config) unsetLocked(key string) {
	// Viper is not meant for deleting keys - we deep-copy everything,
	// skipping keys that match
	newConf := viper.New()
	for _, k := range c.AllKeys() {
		if k == key || strings.HasPrefix(k, key+".") {
			continue
		}
		newConf.Set(k, c.Get(k))
	}

	c.Viper = newConf
}

func (c *Config) Write() derrors.Error {
	// Writing of sub-configs will write the parent
	if c.parent != nil {
		// Merge key and write
		c.MergeToParent()
		return c.parent.Write()
	}

	// Writing should be thread-safe
	c.writeLock.Lock()
	defer c.writeLock.Unlock()

	confDir := filepath.Dir(c.ConfigFile)
	err := os.MkdirAll(confDir, 0755)
	if err != nil {
		return derrors.NewPermissionDeniedError("failed creating config dir", err).WithParams(confDir)
	}

	// We set this here in case we re-created a config when deleting keys
	c.SetConfigFile(c.ConfigFile)

	// Unstable version of viper allows to set filemode, we need
	// to do it after writing. This does introduce a slight vulnerability
	// as there is a small window during which the file can be read.
	// This will be fixed with the next version of viper.
	err = c.WriteConfig()
	if err != nil {
		return derrors.NewInternalError("failed writing config file", err).WithParams(c.ConfigFile)
	}

	err = os.Chmod(c.ConfigFile, 0600)
	if err != nil {
		return derrors.NewInternalError("failed setting config file permissions", err).WithParams(c.ConfigFile)
	}

	return nil
}

func (c *Config) Print() {
	log.Info().Str("app", version.AppVersion).Str("commit", version.Commit).Msg("version")
	for _, key := range c.AllKeys() {
		val := c.Get(key)

		// Don't print secrets
		if strings.Contains(key, "token") {
			val = interface{}(strings.Repeat("*", len(val.(string))))
		}

		log.Info().Interface(key, val).Msg("configuration value")
	}
}
