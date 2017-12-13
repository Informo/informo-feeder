package poller

import (
	"rss2informo/common"
	"rss2informo/config"

	"github.com/mmcdole/gofeed"

	"github.com/sirupsen/logrus"
)

type NewsContent struct {
	Headline    string `json:"headline"`
	Content     string `json:"content"`
	Description string `json:"description"`
	Date        int64  `json:"date"` // Timestamp in seconds
	Author      string `json:"author"`
	Link        string `json:"link"`
}

func (p *Poller) sendMatrixEventFromItem(
	feed config.Feed, feedItem *gofeed.Item,
) (err error) {
	content, err := p.getEventContent(feedItem)
	if err != nil {
		return
	}

	r, err := p.mxClient.SendMessageEvent(
		common.InformoRoomID,
		feed.EventType,
		content,
	)
	if err != nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"feedURL":   feed.URL,
		"eventType": feed.EventType,
		"eventID":   r.EventID,
	}).Info("Event published")

	return nil
}

func (p *Poller) getEventContent(item *gofeed.Item) (content NewsContent, err error) {
	var authorName string
	if item.Author == nil {
		authorName = ""
	} else {
		authorName = item.Author.Name
	}

	content = NewsContent{
		Headline:    item.Title,
		Content:     item.Content,
		Description: item.Description,
		Date:        item.PublishedParsed.Unix(),
		Author:      authorName,
		Link:        item.Link,
	}

	return
}
