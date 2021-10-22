package emulatedclient

type AnnounceCapabilities struct {
	SupportAnnounceList         bool `yaml:"supportAnnounceList" validate:"required"`
	AnnounceToAllTiers          bool `yaml:"announceToAllTiers" validate:"required"`
	AnnounceToAllTrackersInTier bool `yaml:"announceToAllTrackersInTier" validate:"required"`
}
