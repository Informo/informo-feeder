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
	"github.com/sirupsen/logrus"
)

type utcFormatter struct {
	logrus.Formatter
}

func (f utcFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	entry.Time = entry.Time.UTC()
	return f.Formatter.Format(entry)
}

func logConfig() error {
	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.SetFormatter(&utcFormatter{
		&logrus.TextFormatter{
			TimestampFormat:  "2006-01-02T15:04:05.000000000Z07:00",
			FullTimestamp:    true,
			DisableColors:    false,
			DisableTimestamp: false,
			DisableSorting:   false,
		},
	})

	return nil
}
