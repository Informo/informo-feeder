package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db     *sql.DB
	poller pollerStatements
}

func NewDatabase(dbPath string) (*Database, error) {
	var db *sql.DB
	var err error
	if db, err = sql.Open("sqlite3", dbPath); err != nil {
		return nil, err
	}
	poller := pollerStatements{}
	if err = poller.prepare(db); err != nil {
		return nil, err
	}

	return &Database{db, poller}, nil
}

func (d *Database) UpdatePollerStatusForFeed(
	feedURL string, latestPoll int64, latestItemTs int64,
) error {
	return d.poller.insertOrUpdatePollerStatus(feedURL, latestPoll, latestItemTs)
}

func (d *Database) GetPollerStatusForFeed(
	feedURL string,
) (latestPoll int64, latestItemTs int64, err error) {
	return d.poller.selectLatestPollerStatusForFeed(feedURL)
}
