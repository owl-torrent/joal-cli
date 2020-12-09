package fileutils

import (
	"github.com/pkg/errors"
	"os"
)

func FileExistsStrict(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	if info.IsDir() {
		return false
	}
	return true
}

func DirExistsStrict(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	if !info.IsDir() {
		return false
	}
	return true
}

func FileExists(path string) (bool, error) {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, errors.Wrap(err, "failed to check if file exists")
	}
	if info.IsDir() {
		return false, nil
	}
	return true, nil
}

func DirExists(path string) (bool, error) {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, errors.Wrap(err, "failed to check if directory exists")
	}
	if !info.IsDir() {
		return false, nil
	}
	return true, nil
}
