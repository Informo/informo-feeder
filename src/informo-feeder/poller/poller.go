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
	"errors"
	"net/http"
	"regexp"
	"time"

	"informo-feeder/config"
	"informo-feeder/database"

	"github.com/matrix-org/gomatrix"
	"github.com/mmcdole/gofeed"
	"github.com/sirupsen/logrus"
)

var (
	errNoHTML  = errors.New("Could not find any HTML content")
	htmlRegexp = regexp.MustCompile("</[^ ]+>")
)

type Poller struct {
	db              *database.Database
	mxClient        *gomatrix.Client
	parser          *gofeed.Parser
	cfg             *config.Config
	testMode        bool
	lastPollResults map[string][]string
}

func NewPoller(
	db *database.Database,
	mxClient *gomatrix.Client,
	cfg *config.Config,
	testMode bool,
) *Poller {
	return &Poller{
		db:              db,
		mxClient:        mxClient,
		parser:          gofeed.NewParser(),
		cfg:             cfg,
		testMode:        testMode,
		lastPollResults: make(map[string][]string),
	}
}

func (p *Poller) StartPolling(feed config.Feed) {
	var err error

	for {
		p.lastPollResults[feed.Identifier], err = p.db.GetItemsURLsForFeed(feed.Identifier)
		if err != nil {
			logrus.Panic(err)
		}

		logrus.WithFields(logrus.Fields{
			"feed":  feed.Identifier,
			"items": len(p.lastPollResults[feed.Identifier]),
		}).Debug("Loaded last poll's results")

		logrus.WithField("feedURL", feed.URL).Info("Polling")

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

		logrus.WithFields(logrus.Fields{
			"feed":  feed.Identifier,
			"items": len(f.Items),
		}).Debug("Fetched feed")

		for i := len(f.Items) - 1; i >= 0; i-- {
			item := f.Items[i]
			if !p.itemKnown(feed.Identifier, item.Link) {
				if err = p.prepareThenSend(feed, item); err == errNoHTML {
					logrus.WithFields(logrus.Fields{
						"feed":          feed.Identifier,
						"title":         item.Title,
						"publishedDate": item.PublishedParsed.String(),
					}).Warn("Could not find any HTML content")

					continue
				} else if err != nil {
					logrus.Panic(err)
				}
			}

			time.Sleep(500 * time.Millisecond)
		}

		// Perform the clear + refill here, this way we don't risk the feeder being
		// interupted while sending events and restarting with an empty DB for the
		// feed.
		if err = p.db.ClearItemsForFeed(feed.Identifier); err != nil {
			logrus.Panic(err)
		}

		for _, item := range f.Items {
			if err = p.db.SaveItem(feed.Identifier, item.Link); err != nil {
				panic(err)
			}
		}

		time.Sleep(time.Duration(feed.PollInterval) * time.Second)
	}
}

func (p *Poller) prepareThenSend(feed config.Feed, item *gofeed.Item) error {
	var content string

	if len(item.Content) > 0 {
		content = item.Content
	} else if len(item.Description) > 0 {
		if htmlRegexp.MatchString(item.Description) {
			content = item.Description
		} else {
			return errNoHTML
		}
	}

	logrus.WithFields(logrus.Fields{
		"title":         item.Title,
		"publishedDate": item.PublishedParsed.String(),
	}).Info("Got a new item")

	if err := p.replaceMedias(&content); err != nil {
		return err
	}

	return p.sendMatrixEventFromItem(feed, content, item)
}

func (p *Poller) itemKnown(feedIdentifier string, itemURL string) (known bool) {
	for _, url := range p.lastPollResults[feedIdentifier] {
		if url == itemURL {
			known = true
		}
	}

	return
}

func isTooManyRequestsError(err error) bool {
	httpErr, ok := err.(gomatrix.HTTPError)
	if !ok || httpErr.Code != http.StatusTooManyRequests {
		if ok {
			logrus.WithFields(logrus.Fields{
				"code":    httpErr.Code,
				"message": httpErr.Message,
			}).Debug("HTTP error isn't 429 Too Many Requests")
		}
		return false
	}

	logrus.Debug("Got 429 Too Many Requests error")
	return true
}
