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

// KeysConfig contains the settings required to load the signing keys as loaded
// from the configuration file, along with the actual pairs of keys.
type KeysConfig struct {
	Directory   string                        `yaml:"directory"`
	Prefix      string                        `yaml:"prefix,omitempty"`
	PublicKeys  map[string]ed25519.PublicKey  `yaml:"-"`
	PrivateKeys map[string]ed25519.PrivateKey `yaml:"-"`
}

var (
	// ErrKeysDirNotDir is returned if the path specified in the configuration
	// file as the directory to load/generate the keys into is not a directory.
	ErrKeysDirNotDir = errors.New("The keys directory already exists but isn't a directory")
	// ErrKeyNotInformoPrivateKey is returned if the PEM block that was expected
	// to be an Informo feeder private key isn't.
	ErrKeyNotInformoPrivateKey = errors.New("Key isn't an informo feeder private key")
)

// loadKeys fills the KeysConfig member of a Config instance with the public
// and private key for each source. If there's no key for a source, one is
// generated and saved on disk. If the directory where the keys are supposed to
// be stored doesn't exist, creates it.
// Returns an error if there was an issue creating the keys directory or if its
// path exists but isn't a directory. If an error is encountered while loading
// or generating a key pair, only log the error and iterates to the next source.
func (c *Config) loadKeys() (err error) {
	// Check if the keys directory exists from the value of the error returned by
	// a stat on its path.
	info, err := os.Stat(c.Keys.Directory)
	var created bool
	if created, err = c.createKeysDirIfNotExists(err); err != nil {
		return
	}

	// Only perform these checks if we didn't have to create the directory.
	// At this point, we're sure the directory exists, and if we enter this block
	// it means os.Stat() returned with a FileInfo instance.
	if !created {
		// Check if the path is a directory.
		if !info.Mode().IsDir() {
			return ErrKeysDirNotDir
		}

		// Check if the directory's permissions are 0700 (448 being 0700 converted
		// from octal to decimal). Only display a warning if not.
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
	}

	// Initiate the maps for the public and private keys so we don't panic because
	// we try to write to a forbidden memory address.
	c.Keys.PublicKeys = make(map[string]ed25519.PublicKey)
	c.Keys.PrivateKeys = make(map[string]ed25519.PrivateKey)
	// Iterate over the sources.
	for _, source := range c.Feeds {
		id := source.Identifier
		// Load the pair of keys for this source. If there's no key for a source,
		// it will be generated then loaded.
		c.Keys.PublicKeys[id], c.Keys.PrivateKeys[id], err = c.loadOrGenerateKeys(id)
		if err != nil {
			// Only log any error returned here so we don't break the loop.
			logrus.WithField(
				"identifier", id,
			).Error(err.Error())
		}
	}

	return
}

// createKeysDirIfNotExists checks if an error (returned by os.Stat()) is
// os.ErrNotExist. If so, attempts to create the directory to store the keys in.
// It also returns a boolean set to true if the directory had to be created,
// false if not.
// If the error isn't os.ErrNotExist, returns it so the calling function can
// process it.
// Returns an error if the directory couldn't be created.
func (c *Config) createKeysDirIfNotExists(pathErr error) (created bool, err error) {
	// Check the error
	if !os.IsNotExist(pathErr) {
		return
	}

	logrus.WithField(
		"path", c.Keys.Directory,
	).Info("Attempting to create keys directory")

	// Create the directory
	if err = os.Mkdir(c.Keys.Directory, 0700); err != nil {
		return
	}

	created = true

	logrus.WithField(
		"path", c.Keys.Directory,
	).Info("Keys directory created")

	return
}

// loadOrGenerateKeys returns the public and private keys for a source, identified
// with a given identifier, loaded from a PEM file. If the PEM file doesn't exist,
// it is created and filled with a randomly-generated string which will be used
// as a constant seed to generate the key pair for this source. In both cases,
// both the public key and the private key are returned.
// Returns an error if there was an issue checking if the PEM file exists or
// loading or generating the keys.
func (c *Config) loadOrGenerateKeys(
	identifier string,
) (pub ed25519.PublicKey, priv ed25519.PrivateKey, err error) {
	// Check if the PEM file exists.
	path := filepath.Join(c.Keys.Directory, identifier+".pem")
	_, err = os.Stat(path)
	pemFileExists := !os.IsNotExist(err)
	if err != nil && pemFileExists {
		return
	}

	// Decide what to do based on the existence of the PEM file, and log the
	// chosen operation once it is complete.
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

// generateAndSaveKey generates a constant randomised seed used to retrieve the
// pair of keys for a source and saves it in the given PEM file. It then prints
// out an information message to let the user know about the public key and what
// to do with it.
// Returns an error if there was an issue generating the randomised seed, opening
// the PEM file, writing the seed in it or getting the pair of keys from it.
func (c *Config) generateAndSaveKey(
	pemPath string, identifier string,
) (pub ed25519.PublicKey, priv ed25519.PrivateKey, err error) {
	// Generate a 32-bytes randomised string.
	var data [32]byte
	if _, err = rand.Read(data[:]); err != nil {
		return
	}

	// Open/create the file for writing with the correct permissions.
	keyOut, err := os.OpenFile(pemPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return
	}
	defer keyOut.Close()

	// Write the seed to the file as a PEM block.
	if err = pem.Encode(keyOut, &pem.Block{
		Type:  "INFORMO FEEDER PRIVATE KEY",
		Bytes: data[:],
	}); err != nil {
		return
	}

	// Get the public and private keys from the seed.
	pub, priv, err = ed25519.GenerateKey(bytes.NewBuffer(data[:]))

	// Print out an informational message about the keys. The public key is
	// encoded as base64 (which the JS client can read) so it can be printed in
	// a terminal.
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

// loadKeyFromFile reads the content of a PEM file and extracts the seed from it.
// It then returns the public and private keys generated from the seed.
// Returns an error if there was an issue reading the file, generating the keys,
// or if the PEM block isn't an Informo feeder private key block.
func (c *Config) loadKeyFromFile(
	pemPath string, identifier string,
) (pub ed25519.PublicKey, priv ed25519.PrivateKey, err error) {
	// Read the PEM file's content.
	content, err := ioutil.ReadFile(pemPath)
	if err != nil {
		return
	}

	// Decode the PEM content.
	keyBlock, _ := pem.Decode(content)
	// If the block isn't of the right type, return with an error as we can't be
	// sure of the content (either the file was generated by another program or
	// it has been tempered with).
	if keyBlock.Type != "INFORMO FEEDER PRIVATE KEY" {
		logrus.WithField(
			"identifier", identifier,
		).Error(ErrKeyNotInformoPrivateKey.Error())
		err = ErrKeyNotInformoPrivateKey
		return
	}

	// Generate and return the keys.
	return ed25519.GenerateKey(bytes.NewReader(keyBlock.Bytes))
}
