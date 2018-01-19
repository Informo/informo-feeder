// Copyright 2017 Informo core team <core@informo.network>
//
// Licensed under the GNU Affero General Public License, Version 3.0
// (the "License"); you may not use this file except in compliance with the
// License.
// You may obtain a copy of the License at
//
//     https://www.gnu.org/licenses/agpl-3.0.html
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// MatrixConfig represents the Matrix settings as specified in the configuration
// file.
type MatrixConfig struct {
	Homeserver  string `yaml:"homeserver"`
	AccessToken string `yaml:"access_token"`
	MXID        string `yaml:"mxid"`
}

// DatabaseConfig represents the database settings as specified in the
// configuration file.
type DatabaseConfig struct {
	Path string `yaml:"path,omitempty"`
}

// Feed represents a feed that the Informo feeder will poll at a given frequency.
type Feed struct {
	URL          string `yaml:"url"`
	Identifier   string `yaml:"identifier"`
	PollInterval int64  `yaml:"poll_interval"`
}

// Config represents the top-level configuration structure for the Informo feeder.
type Config struct {
	Keys     KeysConfig     `yaml:"keys"`
	Matrix   MatrixConfig   `yaml:"matrix"`
	Feeds    []Feed         `yaml:"feeds"`
	Database DatabaseConfig `yaml:"database"`
}

// Load creates a new instance of the Config structure, marshal the content from
// the configuration file into it, and loads the pair of signing keys into it.
// It then returns a reference to the Config instance.
// Returns an error if there was an issue opening the configuration file, parsing
// it or loading the keys.
func Load(filePath string) (cfg *Config, err error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return
	}

	cfg = new(Config)
	if err = yaml.Unmarshal(content, cfg); err != nil {
		return
	}

	if err = cfg.loadKeys(); err != nil {
		return
	}

	logrus.Info("Config loaded")

	return
}
