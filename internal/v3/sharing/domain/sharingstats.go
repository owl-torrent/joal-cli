package domain

import "fmt"

type SharingStats struct {
	Download DownloadStat
	Upload   UploadStat
	Left     LeftStat
}

func (s *SharingStats) addDownloaded(bytes int64) error {
	amount, err := NewByteAmount(bytes)
	if err != nil {
		return fmt.Errorf("can not add downloaded stat: %w", err)
	}

	err = s.Download.add(amount)
	if err != nil {
		return fmt.Errorf("can not add downloaded stat: %w", err)
	}
	s.Left.subtract(amount)
	return nil
}

func (s *SharingStats) addUploaded(bytes int64) error {
	amount, err := NewByteAmount(bytes)
	if err != nil {
		return fmt.Errorf("can not add uploaded stat: %w", err)
	}

	err = s.Upload.add(amount)
	if err != nil {
		return fmt.Errorf("can not add uploaded stat: %w", err)
	}
	return nil
}
