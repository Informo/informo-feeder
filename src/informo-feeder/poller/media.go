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

package poller

import (
	"mime"
	"regexp"
	"strings"

	"github.com/matrix-org/gomatrix"
	"github.com/sirupsen/logrus"
)

var imgRegexp = `<img[^>]+["|']((//|http(s?)://)[^"'>]+)["|']`

func (p *Poller) replaceMedias(content *string) error {
	urls := p.getMediaLinks(*content)
	return p.replaceWithMatrixLink(content, urls)
}

func (p *Poller) replaceWithMatrixLink(content *string, urls []string) (err error) {
	if len(urls) > 0 {
		// map[originalURL]matrixURL
		replacements := make(map[string]string)

		var resp *gomatrix.RespMediaUpload
		for _, url := range urls {
			firstIter := true
			for err != nil || firstIter {
				if !firstIter {
					is429 := isTooManyRequestsError(err)
					if !is429 {
						return
					}
				}

				resp, err = p.mxClient.UploadLink(url)

				firstIter = false
			}

			logrus.WithFields(logrus.Fields{
				"originalURL": url,
				"mxURL":       resp.ContentURI,
			}).Debug("Replacing media link in content")

			replacements[url] = resp.ContentURI
		}

		for origURL, mxURL := range replacements {
			*content = strings.Replace(*content, origURL, mxURL, -1)
		}
	}

	return
}

func (p *Poller) getMediaLinks(content string) (medias []string) {
	medias = make([]string, 0)

	r := regexp.MustCompile(imgRegexp)
	urls := r.FindAllStringSubmatch(content, -1)

	for _, url := range urls {
		if isImage(url[1]) {
			medias = append(medias, url[1])
		}
	}

	return
}

func isImage(fileName string) bool {
	splitName := strings.Split(fileName, ".")
	ext := "." + splitName[len(splitName)-1]
	mimetype := mime.TypeByExtension(ext)
	return strings.HasPrefix(mimetype, "image")
}
