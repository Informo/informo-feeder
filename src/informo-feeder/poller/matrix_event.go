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
	"encoding/base64"
	"encoding/json"
	"time"

	"informo-feeder/common"
	"informo-feeder/config"

	"github.com/matrix-org/gomatrix"
	"github.com/matrix-org/gomatrixserverlib"
	"github.com/mmcdole/gofeed"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ed25519"
)

func (p *Poller) sendMatrixEventFromItem(
	feed config.Feed, itemContent string, feedItem *gofeed.Item,
) (err error) {
	var extract string
	var extractMaxLength = 80

	content, err := p.getEventContent(feedItem, itemContent)
	if err != nil {
		return
	}

	if err = p.signEvent(&content, feed.Identifier); err != nil {
		return
	}

	firstIter := true

	var r *gomatrix.RespSendEvent
	for err != nil || firstIter {
		if !firstIter {
			is429 := isTooManyRequestsError(err)
			if !is429 {
				return
			}
			// Wait if the error was "429 Too Many Requests"
			time.Sleep(500 * time.Millisecond)
		}

		if !p.testMode {
			r, err = p.mxClient.SendMessageEvent(
				common.InformoRoomID,
				common.InformoNewsEventTypePrefix+feed.Identifier,
				content,
			)
		} else {
			if len(content.Content) > extractMaxLength {
				extract = content.Content[:extractMaxLength]
			} else {
				extract = content.Content
			}

			logrus.WithFields(logrus.Fields{
				"feedURL":    feed.URL,
				"identifier": feed.Identifier,
				"content":    extract,
			}).Debug("Feed test mode enabled, not sending any actual event")
		}

		firstIter = false
	}

	if !p.testMode {
		logrus.WithFields(logrus.Fields{
			"feedURL":    feed.URL,
			"identifier": feed.Identifier,
			"eventID":    r.EventID,
		}).Info("Event published")
	}

	return nil
}

func (p *Poller) getEventContent(
	item *gofeed.Item, itemContent string,
) (content common.NewsContent, err error) {
	var authorName string
	if item.Author == nil {
		authorName = ""
	} else {
		authorName = item.Author.Name
	}

	content = common.NewsContent{
		Headline:    item.Title,
		Content:     itemContent,
		Description: item.Description,
		Date:        item.PublishedParsed.Unix(),
		Author:      authorName,
		Link:        item.Link,
	}

	return
}

func (p *Poller) signEvent(content *common.NewsContent, eventType string) (err error) {
	jsonBytes, err := json.Marshal(content)
	if err != nil {
		return
	}

	canonical, err := gomatrixserverlib.CanonicalJSON(jsonBytes)
	if err != nil {
		return
	}

	priv := p.cfg.Keys.PrivateKeys[eventType]
	content.Signature = base64.StdEncoding.EncodeToString(ed25519.Sign(priv, canonical))
	return
}
