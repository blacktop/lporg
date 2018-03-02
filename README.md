![logo](https://github.com/blacktop/lporg/raw/master/porg.jpeg)

# lporg

[![Circle CI](https://circleci.com/gh/blacktop/lporg.png?style=shield)](https://circleci.com/gh/blacktop/lporg) [![GitHub release](https://img.shields.io/github/release/blacktop/lporg.svg)](https://github.com/https://github.com/blacktop/lporg/releases/releases) [![License](http://img.shields.io/:license-mit-blue.svg)](http://doge.mit-license.org)

> Organize Your macOS Launchpad Apps

---

## Why

This project is meant to help people setting up a brand new Mac **or** to keep all of their `Launchpad Folders` in sync across devices.

## Features

* Load/Save Launchpad app and folder settings
* Load/Save Dock app ordering settings
* Set desktop background image from URL/path in config

## Tested On

* `macOS 10.12` _(Sierra)_
* `macOS 10.13.2` _(High Sierra)_
* `macOS 10.13.3` _(High Sierra)_

## Install

```sh
$ brew install blacktop/tap/lporg
```

## Getting Started

```sh
Usage: lporg [OPTIONS] COMMAND [arg...]

Organize Your Launchpad

Version: 18.02.04, BuildTime: 20180204
Author: blacktop - <https://github.com/blacktop>

Options:
  --verbose, -V  verbose output
  --icloud, -I   save config to iCloud Drive
  --help, -h     show help
  --version, -v  print the version

Commands:
  default  organize by default app categories
  save     save current launchpad settings
  load     load launchpad settings config from `FILE`
  revert   revert to launchpad settings backup
  help     Shows a list of commands or help for one command

Run 'lporg COMMAND --help' for more information on a command.
```

## Commands

### Default

```sh
$ lporg default
```

Organize your launchpad apps using the default Apple app categories as folders

### Save

```sh
$ lporg save
```

Save your current launchpad app layout to a `launchpad.yaml` file

### Load

```sh
$ lporg load launchpad.yaml
```

Load a launchpad app layout from a YAML config file

### Revert

```sh
$ lporg revert
```

Revert a launchpad app layout to the backed up version stored at `$HOME/.launchpad.yml`

### Example Configs

* [YAML](https://github.com/blacktop/lporg/blob/master/test/launchpad-test.yaml)

## TODO

* [ ] create Brewfile from unfound apps IF they are installable via brew?
* [ ] add ability to save/load JSON as well as YAML
* [ ] add ability to save/load private gist configs
* [ ] add ability to org dock as well `dorg` ? (in progress)
* [ ] add ability to have desktop image be a URL and it will download and check sha256, save in `.lporg` folder and add to desktop
* [ ] add ability to set multiple desktop images
* [x] add ability to save/load to/from iCloud Drive `~/Library/Mobile\ Documents/com~apple~CloudDocs`
* [x] backup current launchpad layout before changing
* [x] write backup config to `$HOME/.launchpad.yml`

## Issues

Find a bug? Want more features? Find something missing in the documentation? Let me know! Please don't hesitate to [file an issue](https://github.com/blacktop/lporg/issues/new)

## License

MIT Copyright (c) 2017-2018 **blacktop**
