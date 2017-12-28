![logo](https://github.com/blacktop/lporg/raw/master/porg.jpeg)

# lporg :construction: [WIP]

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

### Example Configs

YAML

```yaml
---
apps:
  pages:
    - number: 1
      flat_items:
        - 1Password
        - App Store
        - Mail
        - Calendar
        - Maps
        - Calculator
        - Contacts
        - System Preferences
      folders:
        - name: Other
          pages:
            - number: 1        
              items:
                - Automator
                - Chess
                - DVD Player
                - Font Book
                - Image Capture
                - QuickTime Player
                - Stickies
                - TextEdit
                - Time Machine
                - Activity Monitor
                - AirPort Utility
                - Audio MIDI Setup
                - Bluetooth File Exchange
                - Boot Camp Assistant
                - ColorSync Utility
                - Console
                - Digital Color Meter
                - Disk Utility
                - Grab
                - Grapher
                - Keychain Access
                - LCC Connection Utility
                - Logitech Unifying Software
                - Migration Assistant
                - Script Editor
                - System Information
                - Terminal
                - VoiceOver Utility
                - XQuartz     
    - number: 2
      folders:
        - name: Porg
          pages:
            - number: 1        
              items:
                - Atom
                - Brave
                - iTerm
                - Porgy Port
    
widgets:
  pages:
    - number: 1
      flat_items:
        - Web Clip
        - Weather
        - Unit Converter
        - World Clock
        - Calendar
        - Calculator
        - Dictionary
        - Contacts
        - Movies
        - Stickies
        - Stocks
        - Tile Game
```

JSON

```json
{
  "apps": {
    "pages": [
      {
        "number": 1,
        "flat_items": [
          "1Password",
          "App Store",
          "Mail",
          "Calendar",
          "Maps",
          "Calculator",
          "Contacts",
          "System Preferences"
        ],
        "folders": [
          {
            "name": "Other",
            "pages": [
              {
                "number": 1,
                "items": [
                  "Automator",
                  "Chess",
                  "DVD Player",
                  "Font Book",
                  "Image Capture",
                  "QuickTime Player",
                  "Stickies",
                  "TextEdit",
                  "Time Machine",
                  "Activity Monitor",
                  "AirPort Utility",
                  "Audio MIDI Setup",
                  "Bluetooth File Exchange",
                  "Boot Camp Assistant",
                  "ColorSync Utility",
                  "Console",
                  "Digital Color Meter",
                  "Disk Utility",
                  "Grab",
                  "Grapher",
                  "Keychain Access",
                  "LCC Connection Utility",
                  "Logitech Unifying Software",
                  "Migration Assistant",
                  "Script Editor",
                  "System Information",
                  "Terminal",
                  "VoiceOver Utility",
                  "XQuartz"
                ]
              }
            ]
          }
        ]
      },
      {
        "number": 2,
        "folders": [
          {
            "name": "Porg",
            "pages": [
              {
                "number": 1,
                "items": ["Atom", "Brave", "iTerm", "Porgy Port"]
              }
            ]
          }
        ]
      }
    ]
  },
  "widgets": {
    "pages": [
      {
        "number": 1,
        "flat_items": [
          "Web Clip",
          "Weather",
          "Unit Converter",
          "World Clock",
          "Calendar",
          "Calculator",
          "Dictionary",
          "Contacts",
          "Movies",
          "Stickies",
          "Stocks",
          "Tile Game"
        ]
      }
    ]
  }
}
```

## TODO

- [x] swith to Apex log
- [x] figure out how to write to DB and not just read :disappointed:
- [x] figure out why creating new groups is failing :confused:
- [x] add Apps/Widgets not included in config to last page by default
- [x] add saving current config out to yaml/json
- [ ] add ability to have flat apps, then folder, then more flat apps on a page
- [ ] add ability to org dock as well `dorg` ? :wink:

## Issues

Find a bug? Want more features? Find something missing in the documentation? Let me know! Please don't hesitate to [file an issue](https://github.com/blacktop/lporg/issues/new)

## License

MIT Copyright (c) 2017-2018 **blacktop**
