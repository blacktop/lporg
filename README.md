![logo](https://github.com/blacktop/lporg/raw/master/porg.jpeg)

# lporg :construction: [WIP]

[![Circle CI](https://circleci.com/gh/blacktop/lporg.png?style=shield)](https://circleci.com/gh/blacktop/lporg) [![GitHub release](https://img.shields.io/github/release/blacktop/lporg.svg)](https://github.com/https://github.com/blacktop/lporg/releases/releases) [![License](http://img.shields.io/:license-mit-blue.svg)](http://doge.mit-license.org)

> Organize Your macOS Launchpad Apps

--------------------------------------------------------------------------------

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

### Example Output

YAML

```yaml
Other:
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

Productivity:
- App Store
- Pages
```

JSON

```json
{
  "RowID": 122,
  "App": {
    "ItemID": 122,
    "Title": "Spotify",
    "BundleID": "com.spotify.client",
    "StoreID": {
      "String": "",
      "Valid": false
    },
    "CategoryID": {
      "Int64": 8,
      "Valid": true
    },
    "Category": {
      "ID": 8,
      "UTI": "public.app-category.music"
    },
    "Moddate": 527776226,
    "Bookmark": "stuff"
  },
  "UUID": "A5A1AAAA-AAAA-AAAA-AAAA-A055AAAD93B",
  "Flags": {
    "Int64": 0,
    "Valid": true
  },
  "Type": {
    "Int64": 4,
    "Valid": true
  },
  "Group": {
    "ID": 173,
    "CategoryID": {
      "Int64": 0,
      "Valid": false
    },
    "Title": {
      "String": "",
      "Valid": false
    }
  },
  "ParentID": 172,
  "Ordering": {
    "Int64": 29,
    "Valid": true
  }
}
```

## TODO

- [x] swith to Apex log
- [ ] figure out how to write to DB and not just read :disappointed:

## Issues

Find a bug? Want more features? Find something missing in the documentation? Let me know! Please don't hesitate to [file an issue](https://github.com/blacktop/lporg/issues/new)

## License

MIT Copyright (c) 2017-2018 **blacktop**
