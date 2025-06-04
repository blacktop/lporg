![logo](https://github.com/blacktop/lporg/raw/master/.github/imgs/porg.jpeg)

# lporg

[![Go](https://github.com/blacktop/lporg/workflows/Go/badge.svg?branch=master)](https://github.com/blacktop/lporg/actions)  [![Downloads](https://img.shields.io/github/downloads/blacktop/lporg/total.svg)](https://github.com/blacktop/lporg/releases)  [![Github All Releases](https://img.shields.io/github/release/blacktop/lporg.svg)](https://github.com/blacktop/lporg/releases) [![License](http://img.shields.io/:license-mit-blue.svg)](http://doge.mit-license.org)

> Organize Your macOS Launchpad Apps

---

## Why

This project is meant to help people setting up a brand new Mac **or** to keep all of their `Launchpad Folders` in sync across devices.

## Features

- Load/Save Launchpad app and folder settings
- Load/Save Dock app ordering settings
- Set desktop background image from URL/path in config

## Tested On

- `macOS 10.12` _(Sierra)_
- `macOS 10.13.2` _(High Sierra)_
- `macOS 10.13.3` _(High Sierra)_
- `macOS 14.0` _(Sonoma)_

## Install

```sh
brew install blacktop/tap/lporg
```

## Getting Started

```sh
❯ lporg

Organize Your Launchpad

Usage:
  lporg [command]

Available Commands:
  default     Organize by default Apple app categories
  help        Help about any command
  load        Load launchpad settings config from `FILE`
  revert      Revert to launchpad settings backup
  save        Save current launchpad settings
  version     Print the version number of lporg

Flags:
  -c, --config string   config file (default is $CONFIG/lporg/config.yaml)
  -h, --help            help for lporg
      --icloud          use iCloud for config
  -V, --verbose         verbose output

Use "lporg [command] --help" for more information about a command.
```

## Commands

### Default

```sh
lporg default
```

Organize your launchpad apps using the default Apple app categories as folders

### Save

```sh
lporg save
```

Save your current launchpad app layout to a `lporg.yml` file

### Load

```sh
lporg load -c lporg.yml
```

Load a launchpad app layout from a YAML config file

### Revert

```sh
lporg revert
```

Revert a launchpad app layout to the backed up version stored at `$CONFIG/lporg/config.yml`

### Example Configs

- [lporg.yml](https://github.com/blacktop/dotfiles/blob/master/init/lporg.yml)

## TODO

- [ ] create Brewfile from unfound apps IF they are installable via brew?
- [ ] add ability to save/load private gist configs
- [ ] add ability to have desktop image be a URL and it will download and check sha256, save in `.lporg` folder and add to desktop

## Issues

Find a bug? Want more features? Find something missing in the documentation? Let me know! Please don't hesitate to [file an issue](https://github.com/blacktop/lporg/issues/new)

## License

MIT Copyright (c) 2017-2024 **blacktop**
