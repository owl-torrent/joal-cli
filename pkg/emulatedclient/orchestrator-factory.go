package emulatedclient

import (
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anthonyraymond/joal-cli/pkg/orchestrator"
)

// The factory implements orchestrator.Iconfig so it can be passed along as an IConfig to the orchestrator.NewOrchestrator
type orchestratorFactory struct {
	SupportAnnounceList         bool `yaml:"supportAnnounceList" validate:"required"`
	AnnounceToAllTiers          bool `yaml:"announceToAllTiers" validate:"required"`
	AnnounceToAllTrackersInTier bool `yaml:"announceToAllTrackersInTier" validate:"required"`
}

func (f orchestratorFactory) DoesSupportAnnounceList() bool {
	return f.SupportAnnounceList
}

func (f orchestratorFactory) ShouldAnnounceToAllTiers() bool {
	return f.AnnounceToAllTiers
}

func (f orchestratorFactory) ShouldAnnounceToAllTrackersInTier() bool {
	return f.AnnounceToAllTrackersInTier
}

func (f orchestratorFactory) createOrchestrator(meta metainfo.MetaInfo) (orchestrator.IOrchestrator, error) {
	return orchestrator.NewOrchestrator(meta, f)
}
