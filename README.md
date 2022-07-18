# Proxy

Fork from: <https://github.com/cyfdecyf/cow>

## Install

```sh
go install github.com/Laisky/cow@master
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


## New Features

`tunnelAllowedPort` support `*` to allow all ports.
