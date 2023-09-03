// Package dock provides functions for manipulating the macOS dock
package dock

import (
	"reflect"
	"testing"
)

func TestLoadDockPlist(t *testing.T) {
	type args struct {
		path []string
	}
	tests := []struct {
		name    string
		args    args
		want    *Plist
		wantErr bool
	}{
		{
			name: "Load Dock Plist",
			args: args{
				path: []string{"../../.hack/test/com.apple.dock.plist"},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadDockPlist(tt.args.path...)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadDockPlist() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadDockPlist() = %v, want %v", got, tt.want)
			}
		})
	}
}
