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

		for _, i := range f.Items {
			if len(i.Content) > 0 && i.PublishedParsed.After(latestItemTime) {
				logrus.WithFields(logrus.Fields{
					"title":         i.Title,
					"publishedDate": i.PublishedParsed.String(),
				}).Info("Got a new item")

				if err = p.replaceMedias(&(i.Content)); err != nil {
					logrus.Panic(err)
				}

				if err = p.sendMatrixEventFromItem(feed, i); err != nil {
					logrus.Panic(err)
				}
			}
			p.updateLatestPosition(feed.URL, f.Items[0])

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
