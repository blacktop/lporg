package desktop

import "testing"

func TestSetDesktopImage(t *testing.T) {
	type args struct {
		image string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "set desktop image",
			args:    args{image: "/Users/blacktop/Library/Mobile Documents/com~apple~CloudDocs/Pictures/downloads/cara-delevingne-mirror.jpg"},
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SetDesktopImage(tt.args.image)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetDesktopImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SetDesktopImage() = %v, want %v", got, tt.want)
			}
		})
	}
}
