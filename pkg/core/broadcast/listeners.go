package broadcast

import "sync"

type ICoreEventListener interface {
	OnSeedStart(event SeedStartedEvent)
	OnSeedStop(event SeedStoppedEvent)
	OnConfigChanged(event ConfigChangedEvent)
	OnTorrentAdded(event TorrentAddedEvent)
	OnTorrentAnnouncing(event TorrentAnnouncingEvent)
	OnTorrentAnnounceSuccess(event TorrentAnnounceSuccessEvent)
	OnTorrentAnnounceFailed(event TorrentAnnounceFailedEvent)
	OnTorrentSwarmChanged(event TorrentSwarmChangedEvent)
	OnTorrentRemoved(event TorrentRemovedEvent)
	OnNoticeableError(event NoticeableErrorEvent)
	OnGlobalBandwidthChanged(event GlobalBandwidthChangedEvent)
	OnBandwidthWeightHasChanged(event BandwidthWeightHasChangedEvent)
}

var listeners ICoreEventListener = &compositeListener{
	listeners: []ICoreEventListener{},
	lock:      &sync.RWMutex{},
}

func RegisterListener(listener ICoreEventListener) (unregisterCallback func()) {
	if listener == nil {
		return
	}
	var listenersAggregator = listeners.(*compositeListener)
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

func (cl *compositeListener) OnSeedStart(event SeedStartedEvent) {
	for _, l := range cl.listeners {
		go l.OnSeedStart(event)
	}
}

func (cl *compositeListener) OnSeedStop(event SeedStoppedEvent) {
	for _, l := range cl.listeners {
		go l.OnSeedStop(event)
	}
}

func (cl *compositeListener) OnConfigChanged(event ConfigChangedEvent) {
	for _, l := range cl.listeners {
		go l.OnConfigChanged(event)
	}
}

func (cl *compositeListener) OnTorrentAdded(event TorrentAddedEvent) {
	for _, l := range cl.listeners {
		go l.OnTorrentAdded(event)
	}
}

func (cl *compositeListener) OnTorrentAnnouncing(event TorrentAnnouncingEvent) {
	for _, l := range cl.listeners {
		go l.OnTorrentAnnouncing(event)
	}
}

func (cl *compositeListener) OnTorrentAnnounceSuccess(event TorrentAnnounceSuccessEvent) {
	for _, l := range cl.listeners {
		go l.OnTorrentAnnounceSuccess(event)
	}
}

func (cl *compositeListener) OnTorrentAnnounceFailed(event TorrentAnnounceFailedEvent) {
	for _, l := range cl.listeners {
		go l.OnTorrentAnnounceFailed(event)
	}
}

func (cl *compositeListener) OnTorrentSwarmChanged(event TorrentSwarmChangedEvent) {
	for _, l := range cl.listeners {
		go l.OnTorrentSwarmChanged(event)
	}
}

func (cl *compositeListener) OnTorrentRemoved(event TorrentRemovedEvent) {
	for _, l := range cl.listeners {
		go l.OnTorrentRemoved(event)
	}
}

func (cl *compositeListener) OnNoticeableError(event NoticeableErrorEvent) {
	for _, l := range cl.listeners {
		go l.OnNoticeableError(event)
	}
}

func (cl *compositeListener) OnGlobalBandwidthChanged(event GlobalBandwidthChangedEvent) {
	for _, l := range cl.listeners {
		go l.OnGlobalBandwidthChanged(event)
	}
}

func (cl *compositeListener) OnBandwidthWeightHasChanged(event BandwidthWeightHasChangedEvent) {
	for _, l := range cl.listeners {
		go l.OnBandwidthWeightHasChanged(event)
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

func (l *BaseCoreEventListener) OnSeedStart(event SeedStartedEvent) {
	if l.OnSeedStartFunc != nil {
		l.OnSeedStartFunc(event)
	}
}

func (l *BaseCoreEventListener) OnSeedStop(event SeedStoppedEvent) {
	if l.OnSeedStopFunc != nil {
		l.OnSeedStopFunc(event)
	}
}

func (l *BaseCoreEventListener) OnConfigChanged(event ConfigChangedEvent) {
	if l.OnConfigChangedFunc != nil {
		l.OnConfigChangedFunc(event)
	}
}

func (l *BaseCoreEventListener) OnTorrentAdded(event TorrentAddedEvent) {
	if l.OnTorrentAddedFunc != nil {
		l.OnTorrentAddedFunc(event)
	}
}

func (l *BaseCoreEventListener) OnTorrentAnnouncing(event TorrentAnnouncingEvent) {
	if l.OnTorrentAnnouncingFunc != nil {
		l.OnTorrentAnnouncingFunc(event)
	}
}

func (l *BaseCoreEventListener) OnTorrentAnnounceSuccess(event TorrentAnnounceSuccessEvent) {
	if l.OnTorrentAnnounceSuccessFunc != nil {
		l.OnTorrentAnnounceSuccessFunc(event)
	}
}

func (l *BaseCoreEventListener) OnTorrentAnnounceFailed(event TorrentAnnounceFailedEvent) {
	if l.OnTorrentAnnounceFailedFunc != nil {
		l.OnTorrentAnnounceFailedFunc(event)
	}
}

func (l *BaseCoreEventListener) OnTorrentSwarmChanged(event TorrentSwarmChangedEvent) {
	if l.OnTorrentSwarmChangedFunc != nil {
		l.OnTorrentSwarmChangedFunc(event)
	}
}

func (l *BaseCoreEventListener) OnTorrentRemoved(event TorrentRemovedEvent) {
	if l.OnTorrentRemovedFunc != nil {
		l.OnTorrentRemovedFunc(event)
	}
}

func (l *BaseCoreEventListener) OnNoticeableError(event NoticeableErrorEvent) {
	if l.OnNoticeableErrorFunc != nil {
		l.OnNoticeableErrorFunc(event)
	}
}

func (l *BaseCoreEventListener) OnGlobalBandwidthChanged(event GlobalBandwidthChangedEvent) {
	if l.OnGlobalBandwidthChangedFunc != nil {
		l.OnGlobalBandwidthChangedFunc(event)
	}
}

func (l *BaseCoreEventListener) OnBandwidthWeightHasChanged(event BandwidthWeightHasChangedEvent) {
	if l.OnBandwidthWeightHasChangedFunc != nil {
		l.OnBandwidthWeightHasChangedFunc(event)
	}
}
