package main

import (
	"flag"

	"rss2informo/config"
	"rss2informo/database"
	"rss2informo/poller"

	"github.com/matrix-org/gomatrix"

	"github.com/sirupsen/logrus"
)

var (
	configFile = flag.String("config", "config.yaml", "Configuration file")
)

func main() {
	logConfig()

	flag.Parse()

	cfg, err := config.Load(*configFile)
	if err != nil {
		logrus.Panic(err)
	}

	db, err := database.NewDatabase(cfg.Database.Path)
	if err != nil {
		logrus.Panic(err)
	}

	client, err := gomatrix.NewClient(
		cfg.Matrix.Homeserver, cfg.Matrix.MXID, cfg.Matrix.AccessToken,
	)

	p := poller.NewPoller(db, client)
	for _, feed := range cfg.Feeds {
		go p.StartPolling(feed)
		logrus.WithField("feedURL", feed.URL).Info("Poller started")
	}

	select {}
}
