package emulatedclient

type AnnounceCapabilities struct {
	SupportAnnounceList         bool `yaml:"supportAnnounceList"`
	AnnounceToAllTiers          bool `yaml:"announceToAllTiers"`
	AnnounceToAllTrackersInTier bool `yaml:"announceToAllTrackersInTier"`
}
