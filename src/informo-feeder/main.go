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

package main

import (
	"flag"

	"informo-feeder/config"
	"informo-feeder/database"
	"informo-feeder/poller"

	"github.com/matrix-org/gomatrix"

	"github.com/sirupsen/logrus"
)

var (
	configFile = flag.String("config", "config.yaml", "Configuration file")
	debug      = flag.Bool("debug", false, "Print debugging messages")
)

func main() {
	flag.Parse()

	logConfig()

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

	p := poller.NewPoller(db, client, cfg)
	for _, feed := range cfg.Feeds {
		go p.StartPolling(feed)
		logrus.WithField("feedURL", feed.URL).Info("Poller started")
	}

	select {}
}
