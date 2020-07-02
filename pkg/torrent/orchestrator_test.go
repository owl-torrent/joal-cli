package torrent

import (
	"context"
	"fmt"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/internal/testutils"
	"github.com/golang/mock/gomock"
	"github.com/nvn1729/congo"
	"net/url"
	"runtime"
	"testing"
	"time"
)

func Test_FallbackOrchestrator_ShouldNotBuildWithEmptyTierList(t *testing.T) {
	_, err := NewFallBackOrchestrator()
	if err == nil {
		t.Fatal("should have failed to build")
	}
}

func Test_FallbackOrchestrator_ShouldAnnounceOnlyOnFirstTierIfItSucceed(t *testing.T) {
	var tiers []ITierAnnouncer

	latch := congo.NewCountDownLatch(5) // wait 5 announce, all of them should be done with first tier

	for i := 0; i < 30; i++ {
		tr := newTracker(*testutils.MustParseUrl(fmt.Sprintf("http://localhost/%d", i)))
		tier, _ := newAllTrackersTierAnnouncer(tr)

		tiers = append(tiers, tier)
	}

	annFunc := buildAnnouncingFunc(1*time.Millisecond, func(u url.URL) {
		if u.Path != "/0" {
			t.Fatalf("tracker %s should not have been called", u.Path)
		}
		latch.CountDown()
	})

	o, _ := NewFallBackOrchestrator(tiers...)
	o.Start(annFunc)
	defer o.Stop(context.Background())
	if !latch.WaitTimeout(500 * time.Millisecond) {
		t.Fatal("latch has not been released")
	}
}

func Test_FallbackOrchestrator_ShouldTryTiersOneByOneUntilOneSucceed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var tiers []ITierAnnouncer

	t1 := NewMockITierAnnouncer(ctrl)
	c1 := make(chan tierState)
	t1.EXPECT().States().Return(c1).AnyTimes()
	tiers = append(tiers, t1)

	t2 := NewMockITierAnnouncer(ctrl)
	c2 := make(chan tierState)
	t2.EXPECT().States().Return(c2).AnyTimes()
	tiers = append(tiers, t2)

	t3 := NewMockITierAnnouncer(ctrl)
	c3 := make(chan tierState)
	t3.EXPECT().States().Return(c3).AnyTimes()
	tiers = append(tiers, t3)

	t4 := NewMockITierAnnouncer(ctrl)
	c4 := make(chan tierState)
	t4.EXPECT().States().Return(c4).AnyTimes()
	tiers = append(tiers, t4)

	latch := congo.NewCountDownLatch(1)
	gomock.InOrder(
		t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) { c1 <- tierState(DEAD) }).Times(1),
		t1.EXPECT().stopAnnounceLoop().Times(1),
		t2.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) { c2 <- tierState(DEAD) }).Times(1),
		t2.EXPECT().stopAnnounceLoop().Times(1),
		t3.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) {
			c3 <- tierState(ALIVE)
			latch.CountDown()
		}).Times(1),
		t3.EXPECT().stopAnnounceLoop().Times(1),
		t3.EXPECT().LastKnownInterval().Return(1800*time.Second, nil).Times(1),
	)

	o, _ := NewFallBackOrchestrator(tiers...)
	o.Start(nil)
	defer o.Stop(context.Background())

	if !latch.WaitTimeout(500 * time.Millisecond) {
		t.Fatal("latch has not been released")
	}
	t1.EXPECT().stopAnnounceLoop().Times(1)

	runtime.Gosched()
	time.Sleep(50 * time.Millisecond) // leave some time to ensure nothing more is called
}

func Test_FallbackOrchestrator_ShouldTryTiersOneByOneUntilOneSucceedUpToLast(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var tiers []ITierAnnouncer

	t1 := NewMockITierAnnouncer(ctrl)
	c1 := make(chan tierState)
	t1.EXPECT().States().Return(c1).AnyTimes()
	tiers = append(tiers, t1)

	t2 := NewMockITierAnnouncer(ctrl)
	c2 := make(chan tierState)
	t2.EXPECT().States().Return(c2).AnyTimes()
	tiers = append(tiers, t2)

	t3 := NewMockITierAnnouncer(ctrl)
	c3 := make(chan tierState)
	t3.EXPECT().States().Return(c3).AnyTimes()
	tiers = append(tiers, t3)

	t4 := NewMockITierAnnouncer(ctrl)
	c4 := make(chan tierState)
	t4.EXPECT().States().Return(c4).AnyTimes()
	tiers = append(tiers, t4)

	latch := congo.NewCountDownLatch(1)
	gomock.InOrder(
		t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) { c1 <- tierState(DEAD) }).Times(1),
		t1.EXPECT().stopAnnounceLoop().Times(1),
		t2.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) { c2 <- tierState(DEAD) }).Times(1),
		t2.EXPECT().stopAnnounceLoop().Times(1),
		t3.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) { c3 <- tierState(DEAD) }).Times(1),
		t3.EXPECT().stopAnnounceLoop().Times(1),
		t4.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) {
			c4 <- tierState(ALIVE)
			latch.CountDown()
		}).Times(1),
		t4.EXPECT().stopAnnounceLoop().Times(1),
		t4.EXPECT().LastKnownInterval().Return(1800*time.Second, nil).Times(1),
	)

	o, _ := NewFallBackOrchestrator(tiers...)
	o.Start(nil)
	defer o.Stop(context.Background())

	if !latch.WaitTimeout(500 * time.Millisecond) {
		t.Fatal("latch has not been released")
	}
	t1.EXPECT().stopAnnounceLoop().Times(1)

	runtime.Gosched()
	time.Sleep(50 * time.Millisecond) // leave some time to ensure nothing more is called
}

func Test_FallbackOrchestrator_ShouldPauseBeforeReAnnouncingIfAllTiersFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tiers := make([]ITierAnnouncer, 0)
	tiersChannels := make([]chan tierState, 0)

	for i := 0; i < 3; i++ {
		tier := NewMockITierAnnouncer(ctrl)
		c := make(chan tierState)

		tiers = append(tiers, tier)
		tiersChannels = append(tiersChannels, c)

		tier.EXPECT().States().Return(c).AnyTimes()
		tier.EXPECT().stopAnnounceLoop().AnyTimes()
	}

	// After all tiers has failed, primary tier will be asked for the last known interval, this test will verify that the tier does wait for the interval and does not re-announce immediatly
	tiers[0].(*MockITierAnnouncer).EXPECT().LastKnownInterval().Return(1800*time.Second, nil).AnyTimes()

	shouldNotRelease := congo.NewCountDownLatch(1)
	gomock.InOrder(
		tiers[0].(*MockITierAnnouncer).EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) {
			tiersChannels[0] <- tierState(DEAD)
		}).Times(1),
		tiers[1].(*MockITierAnnouncer).EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) {
			tiersChannels[1] <- tierState(DEAD)
		}).Times(1),
		tiers[2].(*MockITierAnnouncer).EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) {
			tiersChannels[2] <- tierState(DEAD)
		}).Times(1),
		tiers[0].(*MockITierAnnouncer).EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) {
			// this should not be called because the tier should be in pause
			shouldNotRelease.CountDown()
		}).Times(0),
	)

	o, _ := NewFallBackOrchestrator(tiers...)
	o.Start(nil)
	defer o.Stop(context.Background())

	runtime.Gosched()
	if shouldNotRelease.WaitTimeout(100 * time.Millisecond) {
		t.Fatal("shouldNotRelease shouldn't have been release, startAnnounceLoop has been called immediatly after all tiers failed")
	}
}

func Test_FallbackOrchestrator_ShouldReAnnounceOnFirstTrackerAfterABackupTierHasSucceed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t1 := NewMockITierAnnouncer(ctrl)
	c1 := make(chan tierState)
	t1.EXPECT().States().Return(c1).AnyTimes()
	t2 := NewMockITierAnnouncer(ctrl)
	c2 := make(chan tierState)
	t2.EXPECT().States().Return(c2).AnyTimes()
	t3 := NewMockITierAnnouncer(ctrl)
	c3 := make(chan tierState)
	t3.EXPECT().States().Return(c3).AnyTimes()

	latch := congo.NewCountDownLatch(1)
	gomock.InOrder(
		t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) { c1 <- tierState(DEAD) }).Times(1),
		t1.EXPECT().stopAnnounceLoop().Times(1),
		t2.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) { c2 <- tierState(ALIVE) }).Times(1),
		t2.EXPECT().stopAnnounceLoop().Times(1),
		t2.EXPECT().LastKnownInterval().Return(1*time.Millisecond, nil).Times(1),
		t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.None)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) {
			c1 <- tierState(ALIVE)
			latch.CountDown()
		}).Times(1),
	)

	o, _ := NewFallBackOrchestrator(t1, t2, t3)
	o.Start(nil)
	defer o.Stop(context.Background())

	if !latch.WaitTimeout(500 * time.Millisecond) {
		t.Fatal("latch has not been released")
	}
	t1.EXPECT().stopAnnounceLoop().Times(1)
}

func Test_FallbackOrchestrator_ShouldKeepAnnouncingToFirstTrackerIfItSucceed(t *testing.T) {
	t.Fatal("not implemented")
}

func Test_FallbackOrchestrator_ShouldStartAndStopLoop(t *testing.T) {
	t.Fatal("not implemented")
}

func Test_FallbackOrchestrator_ShouldNotBlockIfStopIsCalledWhenNotStarted(t *testing.T) {
	t.Fatal("not implemented")
}

func Test_FallbackOrchestrator_ShouldBeSafeToRunWithTremendousAmountOfTiers(t *testing.T) {
	t.Fatal("not implemented")
}

func Test_FallbackOrchestrator_ShouldBeReusableAfterStop(t *testing.T) {
	t.Fatal("Not implemented")
}

func Test_AllOrchestrator_ShouldNotBuildWithEmptyTierList(t *testing.T) {
	t.Fatal("not implemented")
}

func Test_AllOrchestrator_ShouldAnnounceOnlyOnAllTiers(t *testing.T) {
	t.Fatal("not implemented")
}

func Test_AllOrchestrator_ShouldContinueAnnouncingEvenIfOneTierFails(t *testing.T) {
	t.Fatal("not implemented")
}

func Test_AllOrchestrator_ShouldContinueAnnouncingEvenIfAllTierFails(t *testing.T) {
	t.Fatal("not implemented")
}

func Test_AllOrchestrator_ShouldStartAndStopLoop(t *testing.T) {
	t.Fatal("not implemented")
}

func Test_AllOrchestrator_ShouldNotBlockIfStopIsCalledWhenNotStarted(t *testing.T) {
	t.Fatal("not implemented")
}

func Test_AllOrchestrator_ShouldBeSafeToRunWithTremendousAmountOfTiers(t *testing.T) {
	t.Fatal("not implemented")
}

func Test_AllOrchestrator_ShouldBeReusableAfterStop(t *testing.T) {
	t.Fatal("Not implemented")
}
