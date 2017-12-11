package database

import (
	"database/sql"
)

const pollerSchema = `
-- Store poller status
CREATE TABLE IF NOT EXISTS poller (
	feed_url TEXT NOT NULL PRIMARY KEY,
	-- The latest poll time (as a timestamp in seconds)
	latest_poll_ts INTEGER NOT NULL DEFAULT 0,
	-- The timestamp (in seconds) of the latest item retrieved
	latest_item_ts INTEGER NOT NULL DEFAULT 0
);
`

const insertOrUpdatePollerStatusSQL = `
	INSERT OR REPLACE INTO poller (feed_url, latest_poll_ts, latest_item_ts)
	VALUES ($1, $2, $3)
`

const selectLatestPollerStatusForFeedSQL = "" +
	"SELECT latest_poll_ts, latest_item_ts FROM poller WHERE feed_url = $1"

type pollerStatements struct {
	insertOrUpdatePollerStatusStmt      *sql.Stmt
	selectLatestPollerStatusForFeedStmt *sql.Stmt
}

func (p *pollerStatements) prepare(db *sql.DB) (err error) {
	_, err = db.Exec(pollerSchema)
	if err != nil {
		return
	}
	if p.insertOrUpdatePollerStatusStmt, err = db.Prepare(insertOrUpdatePollerStatusSQL); err != nil {
		return
	}
	if p.selectLatestPollerStatusForFeedStmt, err = db.Prepare(selectLatestPollerStatusForFeedSQL); err != nil {
		return
	}
	return
}

func (p *pollerStatements) insertOrUpdatePollerStatus(
	feedURL string, latestPoll int64, latestItemTs int64,
) (err error) {
	_, err = p.insertOrUpdatePollerStatusStmt.Exec(feedURL, latestPoll, latestItemTs)
	return
}

func (p *pollerStatements) selectLatestPollerStatusForFeed(
	feedURL string,
) (latestPoll int64, latestItemTs int64, err error) {
	err = p.selectLatestPollerStatusForFeedStmt.QueryRow(feedURL).Scan(&latestPoll, &latestItemTs)

	if err == sql.ErrNoRows {
		return 0, 0, nil
	}

	return
}
