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
	"time"

	"informo-feeder/common"
	"informo-feeder/config"

	"github.com/matrix-org/gomatrix"
	"github.com/mmcdole/gofeed"
	"github.com/sirupsen/logrus"
)

func (p *Poller) sendMatrixEventFromItem(
	feed config.Feed, feedItem *gofeed.Item,
) (err error) {
	content, err := p.getEventContent(feedItem)
	if err != nil {
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

			time.Sleep(5 * time.Second)
		}

		r, err = p.mxClient.SendMessageEvent(
			common.InformoRoomID,
			common.InformoNewsEventTypePrefix+feed.EventType,
			content,
		)

		firstIter = false
	}

	logrus.WithFields(logrus.Fields{
		"feedURL":   feed.URL,
		"eventType": feed.EventType,
		"eventID":   r.EventID,
	}).Info("Event published")

	return nil
}

func (p *Poller) getEventContent(
	item *gofeed.Item,
) (content common.NewsContent, err error) {
	var authorName string
	if item.Author == nil {
		authorName = ""
	} else {
		authorName = item.Author.Name
	}

	content = common.NewsContent{
		Headline:    item.Title,
		Content:     item.Content,
		Description: item.Description,
		Date:        item.PublishedParsed.Unix(),
		Author:      authorName,
		Link:        item.Link,
	}

	err = p.pgpEntity.SignNews(&content)
	return
}
