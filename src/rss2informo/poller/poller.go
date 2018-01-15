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
	"net/http"
	"time"

	"rss2informo/config"
	"rss2informo/database"

	"github.com/matrix-org/gomatrix"
	"github.com/mmcdole/gofeed"
	"github.com/sirupsen/logrus"
)

type Poller struct {
	db       *database.Database
	mxClient *gomatrix.Client
	parser   *gofeed.Parser
}

func NewPoller(db *database.Database, mxClient *gomatrix.Client) *Poller {
	return &Poller{
		db:       db,
		mxClient: mxClient,
		parser:   gofeed.NewParser(),
	}
}

func (p *Poller) StartPolling(feed config.Feed) {
	for {
		timeToSleep, err := p.getDurationBeforePoll(feed)
		if err != nil {
			logrus.Panic(err)
		}

		time.Sleep(timeToSleep * time.Second)

		logrus.WithField("feedURL", feed.URL).Info("Polling")

		_, latestItemTime, err := p.getLatestPosition(feed.URL)
		if err != nil {
			logrus.Panic(err)
		}

		resp, err := http.Get(feed.URL)
		if err != nil {
			logrus.Panic(err)
		}

		if resp.StatusCode != http.StatusOK {
			continue
		}

		f, err := p.parser.Parse(resp.Body)
		if err != nil {
			logrus.Panic(err)
		}

		for i := len(f.Items) - 1; i >= 0; i-- {
			item := f.Items[i]
			if len(item.Content) > 0 && item.PublishedParsed.After(latestItemTime) {
				logrus.WithFields(logrus.Fields{
					"title":         item.Title,
					"publishedDate": item.PublishedParsed.String(),
				}).Info("Got a new item")

				if err = p.replaceMedias(&(item.Content)); err != nil {
					logrus.Panic(err)
				}

				if err = p.sendMatrixEventFromItem(feed, item); err != nil {
					logrus.Panic(err)
				}
			}
			p.updateLatestPosition(feed.URL, item)

			time.Sleep(500 * time.Millisecond)
		}
	}
}

func (p *Poller) getDurationBeforePoll(feed config.Feed) (timeToPoll time.Duration, err error) {
	latestPoll, _, err := p.getLatestPosition(feed.URL)
	if err != nil {
		return
	}

	delta := time.Now().Unix() - latestPoll.Unix()

	timeToPoll = time.Duration(feed.PollInterval - delta)

	if timeToPoll < 0 {
		timeToPoll = 0
	}

	return
}

func (p *Poller) getLatestPosition(
	feedURL string,
) (latestPoll time.Time, latestItem time.Time, err error) {
	lp, li, err := p.db.GetPollerStatusForFeed(feedURL)
	if err != nil {
		return
	}

	latestPoll = time.Unix(lp, 0)
	latestItem = time.Unix(li, 0)
	return
}

func (p *Poller) updateLatestPosition(
	feedURL string, lastItem *gofeed.Item,
) (err error) {
	lastItemTs := lastItem.PublishedParsed.Unix()
	currentTs := time.Now().Unix()

	return p.db.UpdatePollerStatusForFeed(feedURL, currentTs, lastItemTs)
}
