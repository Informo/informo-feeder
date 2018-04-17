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

// Poller describes the overall poller in charge of polling feeds, parsing them
// and sending new events to Matrix.
type Poller struct {
	db       *database.Database
	mxClient *gomatrix.Client
	parser   *gofeed.Parser
	cfg      *config.Config
	testMode bool
}

// NewPoller instantiates a new Poller.
func NewPoller(
	db *database.Database,
	mxClient *gomatrix.Client,
	cfg *config.Config,
	testMode bool,
) *Poller {
	return &Poller{
		db:       db,
		mxClient: mxClient,
		parser:   gofeed.NewParser(),
		cfg:      cfg,
		testMode: testMode,
	}
}

// StartPolling starts an infinite loop that will:
//     - load the results of the previous poll from the database
//     - poll and parse the given feed
//     - send to Matrix each item that wasn't retrieved in the previous poll
//     - erase the results of the previous poll
//     - save the result from the current iteration to the database
//     - wait for a given time (specified in the configuration file)
// If a fatal error is encountered, it panics rather than returning an error.
func (p *Poller) StartPolling(feed config.Feed) {
	var err error
	var lastPollResults map[string]bool
	var itemIsKnown bool

	for {
		// Load the last poll's results.
		lastPollResults, err = p.db.GetItemsURLsForFeed(feed.Identifier)
		if err != nil {
			logrus.Panic(err)
		}

		logrus.WithFields(logrus.Fields{
			"feed":  feed.Identifier,
			"items": len(lastPollResults),
		}).Debug("Loaded last poll's results")

		logrus.WithField("feedURL", feed.URL).Info("Polling")

		// Retrieve the feed's XML.
		resp, err := http.Get(feed.URL)
		if err != nil {
			logrus.Panic(err)
		}

		// If the server didn't reply with a 200 OK status code, wait for the
		// correct amount of time, then jump to the next iteration.
		if resp.StatusCode != http.StatusOK {
			time.Sleep(time.Duration(feed.PollInterval) * time.Second)

			continue
		}

		// Parse the XML retrieved from the remote server.
		f, err := p.parser.Parse(resp.Body)
		if err != nil {
			logrus.Panic(err)
		}

		logrus.WithFields(logrus.Fields{
			"feed":  feed.Identifier,
			"items": len(f.Items),
		}).Debug("Fetched feed")

		// Iterate over the posts in chronological order. We can't promise to
		// send all events chronologically (for example, if a new item appears
		// in the middle of the feed between two iterations, we will send it
		// after all the others, that we retrieved from the previous iteration),
		// but we try to.
		for i := len(f.Items) - 1; i >= 0; i-- {
			item := f.Items[i]
			// If the URL isn't part of the map, itemIsKnown will equal false.
			_, itemIsKnown = lastPollResults[item.Link]
			// Only send the event if it wasn't part of the previous iteration.
			if !itemIsKnown {
				// Not findind any HTML in an item isn't a fatal error, log it
				// and jump to the next iteration (after waiting enough).
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

				// Save the new item to the database.
				if err = p.db.SaveItem(feed.Identifier, item.Link); err != nil {
					logrus.Panic(err)
				}
			}
		}

		// Wait before jumping to the next iteration.
		time.Sleep(time.Duration(feed.PollInterval) * time.Second)
	}
}

// prepareThenSend checks if any HTML could be found in the item (if there is a
// content, it's always HTML, if not, checks if HTML could be found in the item's
// description), in which case it will replace media links (with mxc:// URLs)
// in the item's HTML, then send it to Matrix.
// Returns an error if no HTML could be found, or if replacing medias failed.
func (p *Poller) prepareThenSend(feed config.Feed, item *gofeed.Item) error {
	// Look for HTML content.
	var content string
	if len(item.Content) > 0 {
		// If there's a content, it's always HTML.
		content = item.Content
	} else if len(item.Description) > 0 {
		// If there's a description, check if it contains HTML.
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

	// Replace media links with mxc:// URLs.
	if err := p.replaceMedias(&content); err != nil {
		return err
	}

	// Create and send a Matrix event for this item.
	return p.sendMatrixEventFromItem(feed, content, item)
}

// isTooManyRequestsError checks if the given error is a rate limit error sent by
// the Matrix server.
// Logs (if debugging logging is enabled) if the error is an HTTP error sent by
// a Matrix server, then returns the result.
func isTooManyRequestsError(err error) bool {
	// Checks whether the error is an HTTP error retrieved from a Matrix server.
	httpErr, ok := err.(gomatrix.HTTPError)
	if !ok || httpErr.Code != http.StatusTooManyRequests {
		if ok {
			// If it is an HTTP error sent by a Matrix server but not a 429 Too
			// Many Requests error, log it.
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
