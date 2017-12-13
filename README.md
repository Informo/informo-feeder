# RSS 2 Informo

RSS 2 (to) Informo is a tool to publish news over the Informo network from a RSS or Atom feed.

## Build

You can either install RSS 2 Informo by using a release on one of the [repository's releases](https://github.com/Informo/rss2informo/releases), or by building it by yourself.

The projet is built using [gb](https://getgb.io/), which you can install by running:

```
go get github.com/constabulary/gb/...
```

Then all you need to do is:

```bash
git clone https://github.com/Informo/rss2informo
cd rss2informo
gb build
```

You can the run RSS 2 Informo by calling:

```
./bin/rss2informo
```

## Configure

RSS 2 Informo needs a configuration file to run. If not provided in the command line call, the file `config.yaml` (located in the current directory) is used by default. You can override this default by running rss2informo as:

```bash
rss2informo --config /path/to/config.yaml
```

The configuration file should look as follow:

```yaml
matrix:
  homeserver: https://matrix.org
  access_token: ACCESS_TOKEN
  mxid: @acmenews:matrix.org

feeds:
  - url: "http://www.acmenews.org/feed/"
    event_type: "network.informo.news.acmenews"
    poll_interval: 3600

database:
  path: ./rss2informo.db
```

The configuration file is made of three parts:

* The `matrix` part contains all the necessary parameters to connect to the Informo node (which is a Matrix homeserver) used to send news articles to the network:
    + The `homeserver` is the address at which the node can be reached (with the protocol part).
    + The `access_token` is the Matrix access token to use when sending events.
    + The `mxid` is the Matrix ID to use to publish events (using the form `@localpart:homeserver.tld`).
* The `feeds` part is an array of which each element represents a feed to poll and parse:
    + The `url` is the URL at which the feed can be retrieved in RSS or Atom format.
    + The `event_type` is the event class used when sending the event. It must have been previously declared by an Informo administrator to be accessible by a client. It must begin with `network.informo.news`.
    + The `poll_interval` is the interval in seconds between two retrievals of the feed.
* The `database` part contains all the necessary parameters to connect to the database:
    + The `path` is the path to the SQLite3 database.

The database is used by RSS 2 Informo to store data about the latest poll. It is mainly used to keep track of the latest poll for each feed, so an item doesn't get published more than once on the network, and so a restart doesn't reset the interval between two polls.

## Getting your content on Informo

So as to avoid spam or impersonation, new sources can only be added by manual action from an Informo administrator. This may change later along Matrix's efforts towards decentralised reputation.

If you wish to add a news source to the Informo network, please send an email to an Informo administrator at <admin@informo.network>. Your email should include the Matrix ID of the account RSS 2 Informo will use to publish news (if you have one set up), along with the URL of your source's website and feed. Your feed must contain the following elements for each RSS item/Atom entry:

Content | RSS item sub-tag | Atom entry sub-tag
--- | --- | ---
Article's headline | `title` | `title`
Article's summary | `description` | `summary`
Article's content | `content:encoded` | `content`
Article's publishing date | `pubDate` | `published`
Article's author's name | `author` | `author`
Article's link | `link` | `link`

If your source isn't considered as scam and is exposing a feed matching these criteria, the administrator in touch will send an event in the network to append your source to the list of the authorised sources, and set your Matrix ID as an authorised publisher for this source.
