# Proxy

Fork from: <https://github.com/cyfdecyf/cow>

## New Features

* default allow all ports
* `tunnelAllowedPort` support `*` to allow all ports.

## Install

### Docker

```sh
docker run -p 7777:7777 ppcelery/cow:latest
```

you can mount your own config rc to `/etc/cow/rc`

### Manual

```sh
go install github.com/Laisky/cow@latest
```

Show Sample config:

(save your config file at `$HOME/.cow/rc`)

```sh
# write default config
cow -sample > $HOME/.cow/rc
```

Add systemd service:

```sh
# check
cow -systemd

# install systemd service
cow -systemd > /etc/systemd/system/cow.service

sudo systemctl enable --now cow
```
