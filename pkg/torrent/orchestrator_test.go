package torrent

import (
	"context"
	"errors"
	"fmt"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/internal/testutils"
	"github.com/go-playground/assert/v2"
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

	t1 := NewMockITierAnnouncer(ctrl)
	c1 := make(chan tierState)
	t1.EXPECT().States().Return(c1).AnyTimes()
	t2 := NewMockITierAnnouncer(ctrl)
	c2 := make(chan tierState)
	t2.EXPECT().States().Return(c2).AnyTimes()
	t3 := NewMockITierAnnouncer(ctrl)
	c3 := make(chan tierState)
	t3.EXPECT().States().Return(c3).AnyTimes()
	t4 := NewMockITierAnnouncer(ctrl)
	c4 := make(chan tierState)
	t4.EXPECT().States().Return(c4).AnyTimes()

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
		t3.EXPECT().LastKnownInterval().Return(1800 * time.Second, nil).Times(1),
	)

	o, _ := NewFallBackOrchestrator(t1, t2, t3, t4)
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

	t1 := NewMockITierAnnouncer(ctrl)
	c1 := make(chan tierState)
	t1.EXPECT().States().Return(c1).AnyTimes()
	t2 := NewMockITierAnnouncer(ctrl)
	c2 := make(chan tierState)
	t2.EXPECT().States().Return(c2).AnyTimes()
	t3 := NewMockITierAnnouncer(ctrl)
	c3 := make(chan tierState)
	t3.EXPECT().States().Return(c3).AnyTimes()
	t4 := NewMockITierAnnouncer(ctrl)
	c4 := make(chan tierState)
	t4.EXPECT().States().Return(c4).AnyTimes()

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
		t4.EXPECT().LastKnownInterval().Return(1800 * time.Second, nil).Times(1),
	)

	o, _ := NewFallBackOrchestrator(t1, t2, t3, t4)
	o.Start(nil)
	defer o.Stop(context.Background())

	if !latch.WaitTimeout(500 * time.Millisecond) {
		t.Fatal("latch has not been released")
	}
	t1.EXPECT().stopAnnounceLoop().Times(1)

	runtime.Gosched()
	time.Sleep(50 * time.Millisecond) // leave some time to ensure nothing more is called
}

func Test_FallbackOrchestrator_ShouldPauseForDelayIfAllTiersFailedToAnnounce(t *testing.T) {
	tiers := make([]ITierAnnouncer, 5)

	for i := 0; i < len(tiers); i++ {
		tr := newTracker(*testutils.MustParseUrl(fmt.Sprintf("http://localhost/%d", i)))
		tier, _ := newAllTrackersTierAnnouncer(tr)

		tiers[i] = tier
	}

	var order []string
	latch := congo.NewCountDownLatch(uint(len(tiers)))
	var annFunc AnnouncingFunction = func(u url.URL, event tracker.AnnounceEvent, ctx context.Context) trackerAnnounceResult {
		defer latch.CountDown()
		order = append(order, u.String())
		return trackerAnnounceResult{Err: errors.New("failed :)")}
	}

	o, _ := NewFallBackOrchestrator(tiers...)
	o.Start(annFunc)
	defer o.Stop(context.Background())
	if !latch.WaitTimeout(500 * time.Millisecond) {
		t.Fatal("latch has not been released")
	}

	runtime.Gosched()
	time.Sleep(50 * time.Millisecond) // small delay to allow the goroutine to announce if it wants to

	expected := []string{
		"http://localhost/0", "http://localhost/1", "http://localhost/2", "http://localhost/3", "http://localhost/4",
	} // should not have another /0 at the end otherwise the pause has not been respected

	assert.Equal(t, order, expected)
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
		t2.EXPECT().LastKnownInterval().Return(1 * time.Millisecond, nil).Times(1),
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
