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
	"net/url"

	// Database driver.
	// TODO: Chose the driver from the configuration file.
	_ "github.com/mattn/go-sqlite3"
)

// Database contains a representation of the database as it is used by the feeder.
type Database struct {
	db     *sql.DB
	poller pollerStatements
}

// NewDatabase returns a new instance of the Database structure.
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

// GetItemsURLsForFeed returns a slice containing the URL of each item retrieved
// the last time the poller polled a given feed.
// Returns an error if the retrieval went wrong.
func (d *Database) GetItemsURLsForFeed(feedIdentifier string) (map[string]bool, error) {
	return d.poller.selectItemsURLsForFeed(feedIdentifier)
}

// SaveItem saves the URL of an item in the database, associated with the feed
// if was retrieved from.
// Returns an error if the insertion went wrong.
func (d *Database) SaveItem(feedIdentifier string, itemURL string) error {
	// Check if the provided URL is valid.
	if _, err := url.Parse(itemURL); err != nil {
		return err
	}

	return d.poller.insertItemForFeed(feedIdentifier, itemURL)
}

// ClearItemsForFeed removes all items from the database associated with a given
// feed.
// Returns an error if the deletion went wrong.
func (d *Database) ClearItemsForFeed(feedIdentifier string) error {
	return d.poller.deleteItemsForFeed(feedIdentifier)
}
