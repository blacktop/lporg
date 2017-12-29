![logo](https://github.com/blacktop/lporg/raw/master/porg.jpeg)

# lporg

[![Circle CI](https://circleci.com/gh/blacktop/lporg.png?style=shield)](https://circleci.com/gh/blacktop/lporg) [![GitHub release](https://img.shields.io/github/release/blacktop/lporg.svg)](https://github.com/https://github.com/blacktop/lporg/releases/releases) [![License](http://img.shields.io/:license-mit-blue.svg)](http://doge.mit-license.org)

> Organize Your macOS Launchpad Apps

--------------------------------------------------------------------------------

## Why

This project is meant to help people setting up a brand new Mac **or** to keep all of their `Launchpad Folders` in sync across devices.

## Install

```sh
$ brew install blacktop/tap/lporg
```

## Getting Started

```sh
Usage: lporg [OPTIONS] COMMAND [arg...]

Organize Your Launchpad

Version: , BuildTime:
Author: blacktop - <https://github.com/blacktop>

Options:
  --verbose, -V  verbose output
  --help, -h     show help
  --version, -v  print the version

Commands:
  default  Organize by Categories
  save     Save Current Launchpad App Config
  load     Load Launchpad App Config From File
  help     Shows a list of commands or help for one command

Run 'lporg COMMAND --help' for more information on a command.
```

> **NOTE:** Tested on High Sierra

## Tested On

- macOS 10.13.2 *(High Sierra)*

### Example Configs

- [YAML](https://github.com/blacktop/lporg/blob/master/examples/launchpad.yaml)
- [JSON](https://github.com/blacktop/lporg/blob/master/examples/launchpad.json)

## TODO

- [x] swith to Apex log
- [x] figure out how to write to DB and not just read :disappointed:
- [x] figure out why creating new groups is failing :confused:
- [x] add Apps/Widgets not included in config to last page by default
- [x] add saving current config out to yaml/json
- [x] save is still broken (maybe because of empty folder)
- [ ] add ability to have flat apps, then folder, then more flat apps on a page
- [ ] create Brewfile from unfound apps IF they are installable via brew?
- [x] fix issue where if no files for a folder are installed don't create that folder (might cause eventual DB coruption)
- [ ] add ability to org dock as well `dorg` ? :wink:

## Issues

Find a bug? Want more features? Find something missing in the documentation? Let me know! Please don't hesitate to [file an issue](https://github.com/blacktop/lporg/issues/new)

## License

MIT Copyright (c) 2017-2018 **blacktop**
