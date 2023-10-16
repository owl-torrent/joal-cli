package fileutils

import (
	"fmt"
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
		return false, fmt.Errorf("failed to check if file exists: %w", err)
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
		return false, fmt.Errorf("failed to check if directory exists: %w", err)
	}
	if !info.IsDir() {
		return false, nil
	}
	return true, nil
}
