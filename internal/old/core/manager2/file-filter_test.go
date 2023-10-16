package manager2

import (
	"github.com/anthonyraymond/watcher"
	"os"
	"testing"
	"time"
)

type dumbFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	sys     interface{}
	dir     bool
}

func (fs *dumbFileInfo) IsDir() bool {
	return fs.dir
}
func (fs *dumbFileInfo) ModTime() time.Time {
	return fs.modTime
}
func (fs *dumbFileInfo) Mode() os.FileMode {
	return fs.mode
}
func (fs *dumbFileInfo) Name() string {
	return fs.name
}
func (fs *dumbFileInfo) Size() int64 {
	return fs.size
}
func (fs *dumbFileInfo) Sys() interface{} {
	return fs.sys
}

func TestFileFilter(t *testing.T) {
	type args struct {
		info dumbFileInfo
	}
	tests := []struct {
		name     string
		args     args
		wantSkip bool
	}{
		{name: "shouldFilterDirectory", args: args{info: dumbFileInfo{dir: true, name: "mydir"}}, wantSkip: true},
		{name: "shouldFilterDirectory", args: args{info: dumbFileInfo{dir: true, name: "mydir.torrent"}}, wantSkip: true},
		{name: "shouldFilterNonTorrentFile", args: args{info: dumbFileInfo{name: "file.txt"}}, wantSkip: true},
		{name: "shouldFilterNonTorrentFile", args: args{info: dumbFileInfo{name: "torrent.txt"}}, wantSkip: true},
		{name: "shouldFilterNonTorrentFile", args: args{info: dumbFileInfo{name: "file.txttorrent"}}, wantSkip: true},
		{name: "shouldFilterNonTorrentFile", args: args{info: dumbFileInfo{name: "file.torrent.txt"}}, wantSkip: true},
		{name: "shouldFilterDotFile", args: args{info: dumbFileInfo{name: ".torrent"}}, wantSkip: true},
		{name: "shouldBeCaseAgnostic", args: args{info: dumbFileInfo{name: "file.ToRrent"}}, wantSkip: false},
		{name: "shouldAcceptFileWithMultipleDots", args: args{info: dumbFileInfo{name: "my.file.ToRrent"}}, wantSkip: false},
		{name: "shouldAcceptFileWithSpace", args: args{info: dumbFileInfo{name: "my file.torrent"}}, wantSkip: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := torrentFileFilter(&tt.args.info, "-")
			if tt.wantSkip && err != watcher.ErrSkip {
				t.Errorf("expected %v to be skipped", tt.args.info)
			}
			if !tt.wantSkip && err != nil {
				t.Errorf("%v should not have been skipped", tt.args.info)
			}
		})
	}
}
