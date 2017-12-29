![logo](https://github.com/blacktop/lporg/raw/master/porg.jpeg)

# lporg

[![Circle CI](https://circleci.com/gh/blacktop/lporg.png?style=shield)](https://circleci.com/gh/blacktop/lporg) [![GitHub release](https://img.shields.io/github/release/blacktop/lporg.svg)](https://github.com/https://github.com/blacktop/lporg/releases/releases) [![License](http://img.shields.io/:license-mit-blue.svg)](http://doge.mit-license.org)

> Organize Your macOS Launchpad Apps

--------------------------------------------------------------------------------

## Why

This project is meant to help people setting up a brand new Mac **or** to keep all of their `Launchpad Folders` in sync across devices.

## Tested On

- `macOS 10.13.2` *(High Sierra)*

## Install

```sh
$ brew install blacktop/tap/lporg
```

## Getting Started

```sh
Usage: lporg [OPTIONS] COMMAND [arg...]

Organize Your Launchpad

Version: 17.12.4, BuildTime: 20171229
Author: blacktop - <https://github.com/blacktop>

Options:
  --verbose, -V  verbose output
  --help, -h     show help
  --version, -v  print the version

Commands:
  default  organize by default app categories
  save     save current launchpad settings
  load     load launchpad settings config from `FILE`
  help     Shows a list of commands or help for one command

Run 'lporg COMMAND --help' for more information on a command.
```

## Features

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

### Example Configs

- [YAML](https://github.com/blacktop/lporg/blob/master/examples/launchpad.yaml)
- [JSON](https://github.com/blacktop/lporg/blob/master/examples/launchpad.json)

## TODO

- [ ] create Brewfile from unfound apps IF they are installable via brew?
- [ ] add ability to save/load JSON as well as YAML
- [ ] add ability to org dock as well `dorg` ? :wink:
- [ ] backup current launchpad layout before changing
- [ ] write backup config to `$HOME/.launchpad.yml`

## Issues

Find a bug? Want more features? Find something missing in the documentation? Let me know! Please don't hesitate to [file an issue](https://github.com/blacktop/lporg/issues/new)

## License

MIT Copyright (c) 2017-2018 **blacktop**
