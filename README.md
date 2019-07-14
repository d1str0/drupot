# Drupot
Drupal Honeypot

## Installation
Drupot supports go modules. 

`go get github.com/d1str0/drupot`

`go build`

## Running Drupot
`./drupot -c config.toml`

## Configuration
`config.toml.example` contains an example of *all* currently available
configuration options.

### Drupal
    [drupal]
    port = 80
    changelog_filepath = "changelogs/CHANGELOG-7.63.txt"

`port` allows you to set the http port to listen on. Currently, this is only ever
served over http. Future versions will support https.

`changelog_filepath` allows you to set what exactly is returned in the
/CHANGELOG.txt file. This allows you to save multiple versions of the CHANGELOG
and serve them at different times. This allows you to mimic different versions
of Drupal.

### hpfeeds
    [hpfeeds]
    enabled = true
    host = "hpfeeds.threatstream.com"
    port = 10000
    ident = "agave"
    auth = "somesecret"
    channel = "agave.events"
    meta = "Drupal scan event detected"

hpfeeds can be enabled for logging if wanted. Supply host, port, ident, auth,
and channel information relevant to an hpfeeds broker you want to report to. 

`meta` provides a static string to send in every hpfeeds request. Could be use
to differentiate Drupal versions hosted by honeypot or used to differentiate
Drupot data in busy hpfeeds channels.

### Fetch Public IP
    [fetch_public_ip]
    enabled = true
    urls = ["http://icanhazip.com/", "http://ifconfig.me/ip"]
    

If enabled, Drupot will attempt to fetch the public IP of itself from the listed
URLs. If enabled and no public IP can be fetched, Drupot will quit.

## Sister Projects
* [Magenpot](https://github.com/trevorleake/magenpot), a Magento honeypot
* [bbpot](https://github.com/d1str0/bbpot), a phpBB honeypot (WIP)
* [Presspot](https://github.com/brooks32/presspot), a WordPress honeypot (WIP)
