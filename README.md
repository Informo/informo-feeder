# Informo feeder

The Informo feeder is a tool to publish news over the Informo network from a RSS or Atom feed.

## Build

You can either install the Informo feeder by using a release on one of the [repository's releases](https://github.com/Informo/informo-feeder/releases), or by building it by yourself.

The projet is built using [gb](https://getgb.io/), which you can install by running:

```
go get github.com/constabulary/gb/...
```

Then all you need to do is:

```bash
git clone https://github.com/Informo/informo-feeder
cd informo-feeder
gb build
```

You can the run the Informo feeder by calling:

```
./bin/informo-feeder
```

## Configure

The Informo feeder needs a configuration file to run. If not provided in the command line call, the file `config.yaml` (located in the current directory) is used by default. You can override this default by running informo-feeder as:

```bash
informo-feeder --config /path/to/config.yaml
```

The configuration file itself is documented in the [`config.sample.yaml` file](/config.sample.yaml).

## Getting your content on Informo

So as to avoid spam or impersonation, new sources can only be added by manual action from an Informo administrator. This may change later along Matrix's efforts towards decentralised reputation.

If you wish to add a news source to the Informo network, please send an email to an Informo administrator at <admin@informo.network>. Your email should include the Matrix ID of the account the Informo feeder will use to publish news (if you have one set up), along with the URL of your source's website and feed. Your feed must contain the following elements for each RSS item/Atom entry:

Content | RSS item sub-tag | Atom entry sub-tag
--- | --- | ---
Article's headline | `title` | `title`
Article's summary | `description` | `summary`
Article's content | `content:encoded` | `content`
Article's publishing date | `pubDate` | `published`
Article's author's name | `author` | `author`
Article's link | `link` | `link`

If your source isn't considered as scam and is exposing a feed matching these criteria, the administrator in touch will send an event in the network to append your source to the list of the authorised sources, and set your Matrix ID as an authorised publisher for this source.
