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

type MatrixConfig struct {
	Homeserver  string `yaml:"homeserver"`
	AccessToken string `yaml:"access_token"`
	MXID        string `yaml:"mxid"`
}

type DatabaseConfig struct {
	Path string `yaml:"path,omitempty"`
}

type Feed struct {
	URL          string `yaml:"url"`
	EventType    string `yaml:"event_type"`
	PollInterval int64  `yaml:"poll_interval"`
}

type Config struct {
	Matrix   MatrixConfig   `yaml:"matrix"`
	Feeds    []Feed         `yaml:"feeds"`
	Database DatabaseConfig `yaml:"database"`
}

func Load(filePath string) (cfg Config, err error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(content, &cfg)
	if err != nil {
		return
	}

	logrus.Info("Config loaded")

	return
}
