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
