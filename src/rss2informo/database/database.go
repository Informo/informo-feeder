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
