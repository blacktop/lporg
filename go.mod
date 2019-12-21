module github.com/blacktop/lporg

go 1.13

replace gopkg.in/AlecAivazis/survey.v1 v1.4.1 => github.com/AlecAivazis/survey v1.4.1

require (
	github.com/AlecAivazis/survey v1.4.1
	github.com/DHowett/go-plist v0.0.0-20171105004507-233df3c4f07b
	github.com/apex/log v1.0.0
	github.com/go-yaml/yaml v0.0.0-20171116090243-287cf08546ab
	github.com/jinzhu/gorm v0.0.0-20160404144928-5174cc5c242a
	github.com/jinzhu/inflection v0.0.0-20170102125226-1c35d901db3d
	github.com/mattn/go-colorable v0.0.9
	github.com/mattn/go-isatty v0.0.3
	github.com/mattn/go-sqlite3 v1.4.0
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b
	github.com/mitchellh/mapstructure v0.0.0-20171017171808-06020f85339e
	github.com/pkg/errors v0.8.0
	github.com/urfave/cli v1.20.0
	golang.org/x/net v0.0.0-20171212005608-d866cfc389ce
	golang.org/x/sys v0.0.0-20171222143536-83801418e1b5
	gopkg.in/AlecAivazis/survey.v1 v1.4.1 // indirect
	gopkg.in/yaml.v2 v2.0.0-20171116090243-287cf08546ab
)

replace gopkg.in/yaml.v2 v2.0.0-20171116090243-287cf08546ab => github.com/go-yaml/yaml v0.0.0-20171116090243-287cf08546ab
