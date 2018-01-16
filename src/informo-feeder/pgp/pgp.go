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

package pgp

import (
	"bytes"
	"errors"

	"informo-feeder/common"
	"informo-feeder/config"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

type Entity openpgp.Entity

var (
	ErrWrongTypeForKey = errors.New("Private key is in the wrong type")
)

func NewEntity(cfg *config.PGPConfig) (e *Entity, err error) {
	r := bytes.NewBuffer(cfg.PrivateKey)

	block, err := armor.Decode(r)
	if err != nil {
		return
	}

	if block.Type != openpgp.PrivateKeyType {
		err = ErrWrongTypeForKey
		return
	}

	reader := packet.NewReader(block.Body)

	entity, err := openpgp.ReadEntity(reader)
	if err != nil {
		return
	}

	err = entity.PrivateKey.Decrypt([]byte(cfg.Passphrase))
	e = (*Entity)(entity)
	return
}

func (e *Entity) SignNews(content *common.NewsContent) (err error) {
	strToSign := content.Headline + content.Description + content.Content

	w := new(bytes.Buffer)
	r := bytes.NewBufferString(strToSign)
	entity := (*openpgp.Entity)(e)

	if err = openpgp.ArmoredDetachSign(w, entity, r, nil); err != nil {
		return
	}

	content.Signature = string(w.Bytes())
	return
}
