package manager2

import (
	"github.com/anthonyraymond/watcher"
	"os"
	"regexp"
)

var nameFilter = watcher.RegexFilterHook(regexp.MustCompile(`.+\.(?i)torrent$`), false)

func torrentFileFilter(info os.FileInfo, fullPath string) error {
	if info.IsDir() {
		return watcher.ErrSkip
	}
	return nameFilter(info, fullPath)
}
