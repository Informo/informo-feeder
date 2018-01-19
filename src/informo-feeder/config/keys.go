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
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ed25519"
)

type KeysConfig struct {
	Directory   string                        `yaml:"directory"`
	Prefix      string                        `yaml:"prefix,omitempty"`
	PublicKeys  map[string]ed25519.PublicKey  `yaml:"-"`
	PrivateKeys map[string]ed25519.PrivateKey `yaml:"-"`
}

var (
	ErrKeysDirNotDir           = errors.New("The keys directory already exists but isn't a directory")
	ErrKeyNotInformoPrivateKey = errors.New("Key isn't an informo feeder private key")
)

func (c *Config) loadKeys() (err error) {
	info, err := os.Stat(c.Keys.Directory)

	var created bool
	if created, err = c.createKeysDirIfNotExists(err); err != nil {
		return
	}

	if !created {
		// 448 = mode 0700
		if info.Mode().Perm() != 448 {
			warnMsg := "The mode of the directory containing your private keys is %o."
			warnMsg = warnMsg + " In order to ensure maximum security, it is"
			warnMsg = warnMsg + " recommended to set this to 700 (using 'chmod 700'"
			warnMsg = warnMsg + " for example)."
			logrus.WithFields(logrus.Fields{
				"path": c.Keys.Directory,
				"mode": info.Mode(),
			}).Warnf(warnMsg, info.Mode().Perm())
		}

		if !info.Mode().IsDir() {
			return ErrKeysDirNotDir
		}
	}

	c.Keys.PublicKeys = make(map[string]ed25519.PublicKey)
	c.Keys.PrivateKeys = make(map[string]ed25519.PrivateKey)
	for _, source := range c.Feeds {
		id := source.EventType
		c.Keys.PublicKeys[id], c.Keys.PrivateKeys[id], err = c.loadOrGenerateKeys(id)
		if err != nil {
			logrus.WithField(
				"identifier", id,
			).Error(err.Error())
		}
	}

	return
}

func (c *Config) createKeysDirIfNotExists(pathErr error) (created bool, err error) {
	if !os.IsNotExist(pathErr) {
		return
	}

	logrus.WithField(
		"path", c.Keys.Directory,
	).Info("Attempting to create keys directory")

	if err = os.Mkdir(c.Keys.Directory, 0700); err != nil {
		return
	}

	created = true

	logrus.WithField(
		"path", c.Keys.Directory,
	).Info("Keys directory created")

	return
}

func (c *Config) loadOrGenerateKeys(
	identifier string,
) (pub ed25519.PublicKey, priv ed25519.PrivateKey, err error) {
	path := filepath.Join(c.Keys.Directory, identifier+".pem")
	_, err = os.Stat(path)
	pemFileExists := !os.IsNotExist(err)
	if err != nil && pemFileExists {
		return
	}

	var logMessage string
	if pemFileExists {
		logMessage = "Loaded keys"
		pub, priv, err = c.loadKeyFromFile(path, identifier)
	} else {
		logMessage = "Generated keys"
		pub, priv, err = c.generateAndSaveKey(path, identifier)
	}

	logrus.WithFields(logrus.Fields{
		"identifier": identifier,
		"public_key": base64.StdEncoding.EncodeToString(pub),
	}).Info(logMessage)

	return
}

func (c *Config) generateAndSaveKey(
	pemPath string, identifier string,
) (pub ed25519.PublicKey, priv ed25519.PrivateKey, err error) {
	var data [32]byte
	if _, err = rand.Read(data[:]); err != nil {
		return
	}

	keyOut, err := os.OpenFile(pemPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return
	}
	defer keyOut.Close()

	if err = pem.Encode(keyOut, &pem.Block{
		Type:  "INFORMO FEEDER PRIVATE KEY",
		Bytes: data[:],
	}); err != nil {
		return
	}

	pub, priv, err = ed25519.GenerateKey(bytes.NewBuffer(data[:]))

	keyGenLogMsg := "\n\n"
	keyGenLogMsg = keyGenLogMsg + "A key pair has been generated for the source %s\n"
	keyGenLogMsg = keyGenLogMsg + "The private key is located at %s\n"
	keyGenLogMsg = keyGenLogMsg + "The public key is %s\n"
	keyGenLogMsg = keyGenLogMsg + "Please let the Informo moderation team know"
	keyGenLogMsg = keyGenLogMsg + " about your new public key in order for your"
	keyGenLogMsg = keyGenLogMsg + " news to be verified by the users.\n"
	keyGenLogMsg = keyGenLogMsg + "\n\n"

	fmt.Printf(
		keyGenLogMsg,
		identifier,
		pemPath,
		base64.StdEncoding.EncodeToString(pub),
	)

	return
}

func (c *Config) loadKeyFromFile(
	pemPath string, identifier string,
) (pub ed25519.PublicKey, priv ed25519.PrivateKey, err error) {
	content, err := ioutil.ReadFile(pemPath)
	if err != nil {
		return
	}

	keyBlock, content := pem.Decode(content)
	if keyBlock.Type != "INFORMO FEEDER PRIVATE KEY" {
		logrus.WithField(
			"identifier", identifier,
		).Error(ErrKeyNotInformoPrivateKey.Error())
		err = ErrKeyNotInformoPrivateKey
		return
	}

	return ed25519.GenerateKey(bytes.NewReader(keyBlock.Bytes))
}
