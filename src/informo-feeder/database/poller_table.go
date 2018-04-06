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
-- Store the result from the latest poll for a given feed. One row equals to one
-- item.
CREATE TABLE IF NOT EXISTS poller (
	-- The identifier of the feed the item comes from.
	feed TEXT NOT NULL,
	-- The URL of the item.
	item_url TEXT NOT NULL
);
`

const selectItemsURLsForFeedSQL = `
	SELECT item_url FROM poller WHERE feed = $1
`

const insertItemForFeedSQL = `
	INSERT INTO poller (feed, item_url) VALUES ($1, $2)
`

const deleteItemsForFeedSQL = `
	DELETE FROM poller WHERE feed = $1
`

type pollerStatements struct {
	selectItemsURLsForFeedStmt *sql.Stmt
	insertItemForFeedStmt      *sql.Stmt
	deleteItemsForFeedStmt     *sql.Stmt
}

func (p *pollerStatements) prepare(db *sql.DB) (err error) {
	_, err = db.Exec(pollerSchema)
	if err != nil {
		return
	}
	if p.selectItemsURLsForFeedStmt, err = db.Prepare(selectItemsURLsForFeedSQL); err != nil {
		return
	}
	if p.insertItemForFeedStmt, err = db.Prepare(insertItemForFeedSQL); err != nil {
		return
	}
	if p.deleteItemsForFeedStmt, err = db.Prepare(deleteItemsForFeedSQL); err != nil {
		return
	}
	return
}

func (p *pollerStatements) selectItemsURLsForFeed(feed string) (urls []string, err error) {
	urls = make([]string, 0)

	rows, err := p.selectItemsURLsForFeedStmt.Query(feed)
	if err != nil {
		return
	}

	var u string
	for rows.Next() {
		if err = rows.Scan(&u); err != nil {
			return
		}

		urls = append(urls, u)
	}

	return
}

func (p *pollerStatements) insertItemForFeed(feed string, itemURL string) (err error) {
	_, err = p.insertItemForFeedStmt.Exec(feed, itemURL)

	return
}

func (p *pollerStatements) deleteItemsForFeed(feed string) (err error) {
	_, err = p.deleteItemsForFeedStmt.Exec(feed)

	return
}
