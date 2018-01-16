// Copyright 2018 Informo core team <core@informo.network>
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
	"errors"
	"io/ioutil"
	"os/exec"
)

type PGPConfig struct {
	PrivateKey []byte
	Passphrase string
}

var (
	ErrPGPMissingKeyID   = errors.New("Missing key ID")
	ErrPGPMissingKeyFile = errors.New("Missing key file")
)

func (c *PGPConfig) UnmarshalYAML(unmarshal func(interface{}) error) (err error) {
	var cfg struct {
		UseAgent      bool   `yaml:"use_agent"`
		KeyID         string `yaml:"key_id,omitempty"`
		KeyFile       string `yaml:"key_file,omitempty"`
		KeyPassphrase string `yaml:"key_passphrase"`
	}

	if err = unmarshal(&cfg); err != nil {
		return
	}

	c.Passphrase = cfg.KeyPassphrase

	if cfg.UseAgent {
		if cfg.KeyID == "" {
			return ErrPGPMissingKeyID
		}

		c.PrivateKey, err = getPrivateKeyFromAgent(cfg.KeyID)
	} else {
		if cfg.KeyFile == "" {
			return ErrPGPMissingKeyFile
		}

		c.PrivateKey, err = ioutil.ReadFile(cfg.KeyFile)
	}

	return
}

func getPrivateKeyFromAgent(keyID string) (privkey []byte, err error) {
	cmd := exec.Command(
		"gpg",
		"--armor",
		"--export-secret-keys",
		keyID,
	)

	privkey, err = cmd.Output()
	return
}
