package seedmanager

import (
	"github.com/radovskyb/watcher"
	"os"
	"regexp"
)

func torrentFileFilter() watcher.FilterFileHookFunc {
	nameFilter := watcher.RegexFilterHook(regexp.MustCompile(`.+\.(?i)torrent$`), false)
	fileFilter := func(info os.FileInfo, fullPath string) error {
		if info.IsDir() {
			return watcher.ErrSkip
		}
		return nil
	}

	return func(info os.FileInfo, fullPath string) error {
		err := fileFilter(info, fullPath)
		if err != nil {
			return err
		}
		return nameFilter(info, fullPath)
	}
}
