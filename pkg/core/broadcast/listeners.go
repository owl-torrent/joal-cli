package broadcast

import "sync"

type ICoreEventListener interface {
	onSeedStart(event SeedStartedEvent)
	onSeedStop(event SeedStoppedEvent)
	onConfigChanged(event ConfigChangedEvent)
	onTorrentAdded(event TorrentAddedEvent)
	onTorrentAnnouncing(event TorrentAnnouncingEvent)
	onTorrentAnnounceSuccess(event TorrentAnnounceSuccessEvent)
	onTorrentAnnounceFailed(event TorrentAnnounceFailedEvent)
	onTorrentSwarmChanged(event TorrentSwarmChangedEvent)
	onTorrentRemoved(event TorrentRemovedEvent)
	onNoticeableError(event NoticeableErrorEvent)
	onGlobalBandwidthChanged(event GlobalBandwidthChangedEvent)
	onBandwidthWeightHasChanged(event BandwidthWeightHasChangedEvent)
}

var listenersAggregator ICoreEventListener = &compositeListener{
	listeners: []ICoreEventListener{},
	lock:      &sync.RWMutex{},
}

func RegisterListener(listener ICoreEventListener) (unregisterCallback func()) {
	if listener == nil {
		return
	}
	var listenersAggregator = listenersAggregator.(*compositeListener)
	listenersAggregator.lock.Lock()
	defer listenersAggregator.lock.Unlock()

	listenersAggregator.listeners = append(listenersAggregator.listeners, listener)
	return func() {
		listenersAggregator.lock.Lock()
		defer listenersAggregator.lock.Unlock()
		var index = -1
		for i, l := range listenersAggregator.listeners {
			if l == listener {
				index = i
				break
			}
		}
		if index == -1 {
			// listener not found in array (may happen if Unregister is called twice)
			return
		}

		listenersAggregator.listeners = append(listenersAggregator.listeners[:index], listenersAggregator.listeners[index+1:]...)
	}
}

// this is a thread safe composite pattern implementation of ICoreEventListener
type compositeListener struct {
	listeners []ICoreEventListener
	lock      *sync.RWMutex
}

func (cl *compositeListener) onSeedStart(event SeedStartedEvent) {
	for _, l := range cl.listeners {
		go l.onSeedStart(event)
	}
}

func (cl *compositeListener) onSeedStop(event SeedStoppedEvent) {
	for _, l := range cl.listeners {
		go l.onSeedStop(event)
	}
}

func (cl *compositeListener) onConfigChanged(event ConfigChangedEvent) {
	for _, l := range cl.listeners {
		go l.onConfigChanged(event)
	}
}

func (cl *compositeListener) onTorrentAdded(event TorrentAddedEvent) {
	for _, l := range cl.listeners {
		go l.onTorrentAdded(event)
	}
}

func (cl *compositeListener) onTorrentAnnouncing(event TorrentAnnouncingEvent) {
	for _, l := range cl.listeners {
		go l.onTorrentAnnouncing(event)
	}
}

func (cl *compositeListener) onTorrentAnnounceSuccess(event TorrentAnnounceSuccessEvent) {
	for _, l := range cl.listeners {
		go l.onTorrentAnnounceSuccess(event)
	}
}

func (cl *compositeListener) onTorrentAnnounceFailed(event TorrentAnnounceFailedEvent) {
	for _, l := range cl.listeners {
		go l.onTorrentAnnounceFailed(event)
	}
}

func (cl *compositeListener) onTorrentSwarmChanged(event TorrentSwarmChangedEvent) {
	for _, l := range cl.listeners {
		go l.onTorrentSwarmChanged(event)
	}
}

func (cl *compositeListener) onTorrentRemoved(event TorrentRemovedEvent) {
	for _, l := range cl.listeners {
		go l.onTorrentRemoved(event)
	}
}

func (cl *compositeListener) onNoticeableError(event NoticeableErrorEvent) {
	for _, l := range cl.listeners {
		go l.onNoticeableError(event)
	}
}

func (cl *compositeListener) onGlobalBandwidthChanged(event GlobalBandwidthChangedEvent) {
	for _, l := range cl.listeners {
		go l.onGlobalBandwidthChanged(event)
	}
}

func (cl *compositeListener) onBandwidthWeightHasChanged(event BandwidthWeightHasChangedEvent) {
	for _, l := range cl.listeners {
		go l.onBandwidthWeightHasChanged(event)
	}
}

// This is a base struct provided as a default implementation of ICoreEventListener, it can be used as a no-code opt-in listener
type BaseCoreEventListener struct {
	unregisterCallback              func()
	OnSeedStartFunc                 func(event SeedStartedEvent)
	OnSeedStopFunc                  func(event SeedStoppedEvent)
	OnConfigChangedFunc             func(event ConfigChangedEvent)
	OnTorrentAddedFunc              func(event TorrentAddedEvent)
	OnTorrentAnnouncingFunc         func(event TorrentAnnouncingEvent)
	OnTorrentAnnounceSuccessFunc    func(event TorrentAnnounceSuccessEvent)
	OnTorrentAnnounceFailedFunc     func(event TorrentAnnounceFailedEvent)
	OnTorrentSwarmChangedFunc       func(event TorrentSwarmChangedEvent)
	OnTorrentRemovedFunc            func(event TorrentRemovedEvent)
	OnNoticeableErrorFunc           func(event NoticeableErrorEvent)
	OnGlobalBandwidthChangedFunc    func(event GlobalBandwidthChangedEvent)
	OnBandwidthWeightHasChangedFunc func(event BandwidthWeightHasChangedEvent)
}

func (l *BaseCoreEventListener) Register() {
	l.unregisterCallback = RegisterListener(l)
}

func (l *BaseCoreEventListener) Unregister() {
	if l.unregisterCallback != nil {
		l.unregisterCallback()
		l.unregisterCallback = nil
	}
}

func (l *BaseCoreEventListener) onSeedStart(event SeedStartedEvent) {
	if l.OnSeedStartFunc != nil {
		l.OnSeedStartFunc(event)
	}
}

func (l *BaseCoreEventListener) onSeedStop(event SeedStoppedEvent) {
	if l.OnSeedStopFunc != nil {
		l.OnSeedStopFunc(event)
	}
}

func (l *BaseCoreEventListener) onConfigChanged(event ConfigChangedEvent) {
	if l.OnConfigChangedFunc != nil {
		l.OnConfigChangedFunc(event)
	}
}

func (l *BaseCoreEventListener) onTorrentAdded(event TorrentAddedEvent) {
	if l.OnTorrentAddedFunc != nil {
		l.OnTorrentAddedFunc(event)
	}
}

func (l *BaseCoreEventListener) onTorrentAnnouncing(event TorrentAnnouncingEvent) {
	if l.OnTorrentAnnouncingFunc != nil {
		l.OnTorrentAnnouncingFunc(event)
	}
}

func (l *BaseCoreEventListener) onTorrentAnnounceSuccess(event TorrentAnnounceSuccessEvent) {
	if l.OnTorrentAnnounceSuccessFunc != nil {
		l.OnTorrentAnnounceSuccessFunc(event)
	}
}

func (l *BaseCoreEventListener) onTorrentAnnounceFailed(event TorrentAnnounceFailedEvent) {
	if l.OnTorrentAnnounceFailedFunc != nil {
		l.OnTorrentAnnounceFailedFunc(event)
	}
}

func (l *BaseCoreEventListener) onTorrentSwarmChanged(event TorrentSwarmChangedEvent) {
	if l.OnTorrentSwarmChangedFunc != nil {
		l.OnTorrentSwarmChangedFunc(event)
	}
}

func (l *BaseCoreEventListener) onTorrentRemoved(event TorrentRemovedEvent) {
	if l.OnTorrentRemovedFunc != nil {
		l.OnTorrentRemovedFunc(event)
	}
}

func (l *BaseCoreEventListener) onNoticeableError(event NoticeableErrorEvent) {
	if l.OnNoticeableErrorFunc != nil {
		l.OnNoticeableErrorFunc(event)
	}
}

func (l *BaseCoreEventListener) onGlobalBandwidthChanged(event GlobalBandwidthChangedEvent) {
	if l.OnGlobalBandwidthChangedFunc != nil {
		l.OnGlobalBandwidthChangedFunc(event)
	}
}

func (l *BaseCoreEventListener) onBandwidthWeightHasChanged(event BandwidthWeightHasChangedEvent) {
	if l.OnBandwidthWeightHasChangedFunc != nil {
		l.OnBandwidthWeightHasChangedFunc(event)
	}
}
