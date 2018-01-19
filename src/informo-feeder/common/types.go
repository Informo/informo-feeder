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

package common

type NewsContent struct {
	Headline    string `json:"headline"`
	Content     string `json:"content"`
	Description string `json:"description"`
	Date        int64  `json:"date"` // Timestamp in seconds
	Author      string `json:"author"`
	Link        string `json:"link"`
	Signature   string `json:"signature,omitempty"`
}
