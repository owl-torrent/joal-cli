package torrent

import (
	"context"
	"github.com/anacrolix/torrent/tracker"
	"github.com/golang/mock/gomock"
	"github.com/nvn1729/congo"
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var tiers []ITierAnnouncer

	t1 := NewMockITierAnnouncer(ctrl)
	c1 := make(chan tierState)
	t1.EXPECT().States().Return(c1).AnyTimes()
	t1.EXPECT().stopAnnounceLoop().AnyTimes()
	tiers = append(tiers, t1)

	t2 := NewMockITierAnnouncer(ctrl)
	tiers = append(tiers, t2)

	t3 := NewMockITierAnnouncer(ctrl)
	tiers = append(tiers, t3)

	t4 := NewMockITierAnnouncer(ctrl)
	tiers = append(tiers, t4)

	latch := congo.NewCountDownLatch(1)
	gomock.InOrder(
		t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) {
			defer latch.CountDown()
			c1 <- tierState(ALIVE)
		}).Times(1),
	)

	o, _ := NewFallBackOrchestrator(tiers...)
	o.Start(nil)
	defer o.Stop(context.Background())

	if !latch.WaitTimeout(500 * time.Millisecond) {
		t.Fatal("latch has not been released")
	}

	runtime.Gosched()
	time.Sleep(50 * time.Millisecond) // leave some time to ensure nothing more is called
}

func Test_FallbackOrchestrator_ShouldTryTiersOneByOneUntilOneSucceed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var tiers []ITierAnnouncer

	t1 := NewMockITierAnnouncer(ctrl)
	c1 := make(chan tierState)
	t1.EXPECT().States().Return(c1).AnyTimes()
	t1.EXPECT().stopAnnounceLoop().AnyTimes()
	tiers = append(tiers, t1)

	t2 := NewMockITierAnnouncer(ctrl)
	c2 := make(chan tierState)
	t2.EXPECT().States().Return(c2).AnyTimes()
	t2.EXPECT().stopAnnounceLoop().AnyTimes()
	tiers = append(tiers, t2)

	t3 := NewMockITierAnnouncer(ctrl)
	c3 := make(chan tierState)
	t3.EXPECT().States().Return(c3).AnyTimes()
	t3.EXPECT().stopAnnounceLoop().AnyTimes()
	tiers = append(tiers, t3)

	t4 := NewMockITierAnnouncer(ctrl)
	c4 := make(chan tierState)
	t4.EXPECT().States().Return(c4).AnyTimes()
	t4.EXPECT().stopAnnounceLoop().AnyTimes()
	tiers = append(tiers, t4)

	latch := congo.NewCountDownLatch(1)
	gomock.InOrder(
		t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) { c1 <- tierState(DEAD) }).Times(1),
		t2.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) { c2 <- tierState(DEAD) }).Times(1),
		t3.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) {
			c3 <- tierState(ALIVE)
			latch.CountDown()
		}).Times(1),
		t3.EXPECT().LastKnownInterval().Return(1800*time.Second, nil).Times(1),
	)

	o, _ := NewFallBackOrchestrator(tiers...)
	o.Start(nil)
	defer o.Stop(context.Background())

	if !latch.WaitTimeout(500 * time.Millisecond) {
		t.Fatal("latch has not been released")
	}

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
	t1.EXPECT().stopAnnounceLoop().AnyTimes()
	tiers = append(tiers, t1)

	t2 := NewMockITierAnnouncer(ctrl)
	c2 := make(chan tierState)
	t2.EXPECT().States().Return(c2).AnyTimes()
	t2.EXPECT().stopAnnounceLoop().AnyTimes()
	tiers = append(tiers, t2)

	t3 := NewMockITierAnnouncer(ctrl)
	c3 := make(chan tierState)
	t3.EXPECT().States().Return(c3).AnyTimes()
	t3.EXPECT().stopAnnounceLoop().AnyTimes()
	tiers = append(tiers, t3)

	t4 := NewMockITierAnnouncer(ctrl)
	c4 := make(chan tierState)
	t4.EXPECT().States().Return(c4).AnyTimes()
	t4.EXPECT().stopAnnounceLoop().AnyTimes()
	tiers = append(tiers, t4)

	latch := congo.NewCountDownLatch(1)
	gomock.InOrder(
		t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) { c1 <- tierState(DEAD) }).Times(1),
		t2.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) { c2 <- tierState(DEAD) }).Times(1),
		t3.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) { c3 <- tierState(DEAD) }).Times(1),
		t4.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) {
			c4 <- tierState(ALIVE)
			latch.CountDown()
		}).Times(1),
		t4.EXPECT().LastKnownInterval().Return(1800*time.Second, nil).Times(1),
	)

	o, _ := NewFallBackOrchestrator(tiers...)
	o.Start(nil)
	defer o.Stop(context.Background())

	if !latch.WaitTimeout(500 * time.Millisecond) {
		t.Fatal("latch has not been released")
	}

	runtime.Gosched()
	time.Sleep(50 * time.Millisecond) // leave some time to ensure nothing more is called
}

func Test_FallbackOrchestrator_ShouldPauseBeforeReAnnouncingIfAllTiersFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tiers := make([]ITierAnnouncer, 0)

	t1 := NewMockITierAnnouncer(ctrl)
	c1 := make(chan tierState)
	t1.EXPECT().States().Return(c1).AnyTimes()
	t1.EXPECT().stopAnnounceLoop().AnyTimes()
	tiers = append(tiers, t1)

	t2 := NewMockITierAnnouncer(ctrl)
	c2 := make(chan tierState)
	t2.EXPECT().States().Return(c2).AnyTimes()
	t2.EXPECT().stopAnnounceLoop().AnyTimes()
	tiers = append(tiers, t2)

	t3 := NewMockITierAnnouncer(ctrl)
	c3 := make(chan tierState)
	t3.EXPECT().States().Return(c3).AnyTimes()
	t3.EXPECT().stopAnnounceLoop().AnyTimes()
	tiers = append(tiers, t3)

	// After all tiers has failed, primary tier will be asked for the last known interval, this test will verify that the tier does wait for the interval and does not re-announce immediatly
	t1.EXPECT().LastKnownInterval().Return(1800*time.Second, nil).AnyTimes()

	shouldNotRelease := congo.NewCountDownLatch(1)
	gomock.InOrder(
		t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) {
			c1 <- tierState(DEAD)
		}).Times(1),
		t2.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) {
			c2 <- tierState(DEAD)
		}).Times(1),
		t3.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) {
			c3 <- tierState(DEAD)
		}).Times(1),
	)
	t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) {
		// this should not be called because the tier should be in pause
		shouldNotRelease.CountDown()
	}).Times(0)

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

	tiers := make([]ITierAnnouncer, 0)

	t1 := NewMockITierAnnouncer(ctrl)
	c1 := make(chan tierState)
	t1.EXPECT().States().Return(c1).AnyTimes()
	t1.EXPECT().stopAnnounceLoop().AnyTimes()
	tiers = append(tiers, t1)

	t2 := NewMockITierAnnouncer(ctrl)
	c2 := make(chan tierState)
	t2.EXPECT().States().Return(c2).AnyTimes()
	t2.EXPECT().stopAnnounceLoop().AnyTimes()
	tiers = append(tiers, t2)

	t3 := NewMockITierAnnouncer(ctrl)
	c3 := make(chan tierState)
	t3.EXPECT().States().Return(c3).AnyTimes()
	t3.EXPECT().stopAnnounceLoop().AnyTimes()
	tiers = append(tiers, t3)

	latch := congo.NewCountDownLatch(1)
	gomock.InOrder(
		t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) { c1 <- tierState(DEAD) }).Times(1),
		t2.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) { c2 <- tierState(ALIVE) }).Times(1),
		t2.EXPECT().LastKnownInterval().Return(1*time.Millisecond, nil).Times(1),
		t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.None)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) {
			c1 <- tierState(ALIVE)
			latch.CountDown()
		}).Times(1),
	)

	o, _ := NewFallBackOrchestrator(tiers...)
	o.Start(nil)
	defer o.Stop(context.Background())

	if !latch.WaitTimeout(500 * time.Millisecond) {
		t.Fatal("latch has not been released")
	}
}

func Test_FallbackOrchestrator_ShouldKeepAnnouncingToFirstTrackerIfItSucceed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tiers := make([]ITierAnnouncer, 0)

	t1 := NewMockITierAnnouncer(ctrl)
	c1 := make(chan tierState)
	t1.EXPECT().States().Return(c1).AnyTimes()
	t1.EXPECT().stopAnnounceLoop().AnyTimes()
	tiers = append(tiers, t1)

	t2 := NewMockITierAnnouncer(ctrl)
	c2 := make(chan tierState)
	t2.EXPECT().States().Return(c2).AnyTimes()
	t2.EXPECT().stopAnnounceLoop().AnyTimes()
	tiers = append(tiers, t2)

	t3 := NewMockITierAnnouncer(ctrl)
	c3 := make(chan tierState)
	t3.EXPECT().States().Return(c3).AnyTimes()
	t3.EXPECT().stopAnnounceLoop().AnyTimes()
	tiers = append(tiers, t3)

	gomock.InOrder(
		t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) {
			c1 <- tierState(ALIVE)
		}).Times(1),
	)

	shouldNotRelease := congo.NewCountDownLatch(1)
	t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) {
		// this should not be called because the tier should be in pause
		shouldNotRelease.CountDown()
	}).Times(0)

	o, _ := NewFallBackOrchestrator(tiers...)
	o.Start(nil)
	defer o.Stop(context.Background())

	runtime.Gosched()
	if shouldNotRelease.WaitTimeout(100 * time.Millisecond) {
		t.Fatal("shouldNotRelease shouldn't have been release, startAnnounceLoop has been called immediatly after all tiers failed")
	}
}

func Test_FallbackOrchestrator_ShouldStopPreviousTierWhenMovingToNext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tiers := make([]ITierAnnouncer, 0)

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

	t1.EXPECT().LastKnownInterval().Return(1800*time.Millisecond, nil).Times(1)

	latch := congo.NewCountDownLatch(1)
	gomock.InOrder(
		t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) { c1 <- tierState(DEAD) }).Times(1),
		t1.EXPECT().stopAnnounceLoop().Times(1),
		t2.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) { c2 <- tierState(DEAD) }).Times(1),
		t2.EXPECT().stopAnnounceLoop().Times(1),
		t3.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) { c3 <- tierState(DEAD) }).Times(1),
		t3.EXPECT().stopAnnounceLoop().Do(func() {
			latch.CountDown()
		}).Times(1),
	)

	t1.EXPECT().stopAnnounceLoop().AnyTimes()

	o, _ := NewFallBackOrchestrator(tiers...)
	o.Start(nil)
	defer o.Stop(context.Background())

	if !latch.WaitTimeout(500 * time.Millisecond) {
		t.Fatal("latch has not been released")
	}
}

func Test_FallbackOrchestrator_ShouldStopPreviousTierWhenMovingBackToPrimaryAfterBackupSucceed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tiers := make([]ITierAnnouncer, 0)

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

	t2.EXPECT().LastKnownInterval().Return(1800*time.Millisecond, nil).Times(1)

	latch := congo.NewCountDownLatch(1)
	gomock.InOrder(
		t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) { c1 <- tierState(DEAD) }).Times(1),
		t1.EXPECT().stopAnnounceLoop().Times(1),
		t2.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) { c2 <- tierState(ALIVE) }).Times(1),
		t2.EXPECT().stopAnnounceLoop().Do(func() {
			latch.CountDown()
		}).Times(1),
	)

	t1.EXPECT().stopAnnounceLoop().AnyTimes()

	o, _ := NewFallBackOrchestrator(tiers...)
	o.Start(nil)
	defer o.Stop(context.Background())

	if !latch.WaitTimeout(500 * time.Millisecond) {
		t.Fatal("latch has not been released")
	}
}

func Test_FallbackOrchestrator_ShouldStartAndStopLoop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tiers := make([]ITierAnnouncer, 0)

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

	latch := congo.NewCountDownLatch(1)
	gomock.InOrder(
		t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) {
			defer latch.CountDown()
			c1 <- tierState(ALIVE)
		}).Times(1),
	)

	o, _ := NewFallBackOrchestrator(tiers...)
	o.Start(nil)
	defer o.Stop(context.Background())

	t1.EXPECT().stopAnnounceLoop().Times(1)

	if !latch.WaitTimeout(500 * time.Millisecond) {
		t.Fatal("latch has not been released")
	}
}

func Test_FallbackOrchestrator_ShouldNotBlockIfStopIsCalledWhenNotStarted(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	o, _ := NewFallBackOrchestrator(NewMockITierAnnouncer(ctrl))

	latch := congo.NewCountDownLatch(1)
	go func() {
		o.Stop(context.Background())
		o.Stop(context.Background())
		o.Stop(context.Background())
		o.Stop(context.Background())
		o.Stop(context.Background())
		o.Stop(context.Background())
		latch.CountDown()
	}()

	if !latch.WaitTimeout(500 * time.Millisecond) {
		t.Fatal("Should not have blocked")
	}
}

func Test_FallbackOrchestrator_ShouldBeSafeToRunWithTremendousAmountOfTiers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tiers := make([]ITierAnnouncer, 0)

	latch := congo.NewCountDownLatch(10000)

	for i := 0; i < 3000; i++ {
		tier := NewMockITierAnnouncer(ctrl)
		c := make(chan tierState)
		tier.EXPECT().States().Return(c).AnyTimes()
		tier.EXPECT().stopAnnounceLoop().AnyTimes()
		tier.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) {
			c <- tierState(DEAD)
			latch.CountDown()
		}).AnyTimes()
		tier.EXPECT().LastKnownInterval().Return(1*time.Millisecond, nil).AnyTimes()

		tiers = append(tiers, tier)
	}

	o, _ := NewFallBackOrchestrator(tiers...)
	o.Start(nil)
	defer o.Stop(context.Background())

	if !latch.WaitTimeout(500 * time.Millisecond) {
		t.Fatal("latch has not been released")
	}
}

func Test_FallbackOrchestrator_ShouldBeReusableAfterStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tiers := make([]ITierAnnouncer, 0)

	t1 := NewMockITierAnnouncer(ctrl)
	c1 := make(chan tierState)
	t1.EXPECT().States().Return(c1).AnyTimes()
	t1.EXPECT().stopAnnounceLoop().AnyTimes()
	tiers = append(tiers, t1)

	latch := congo.NewCountDownLatch(1)
	t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) {
		c1 <- tierState(ALIVE)
		latch.CountDown()
	}).Times(1)

	o, _ := NewFallBackOrchestrator(tiers...)

	o.Start(nil)
	if !latch.WaitTimeout(500 * time.Millisecond) {
		t.Fatal("latch has not been released")
	}
	o.Stop(context.Background())

	latch = congo.NewCountDownLatch(1)
	t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) {
		c1 <- tierState(ALIVE)
		latch.CountDown()
	}).Times(1)

	o.Start(nil)
	if !latch.WaitTimeout(500 * time.Millisecond) {
		t.Fatal("latch has not been released")
	}
	o.Stop(context.Background())
}

func Test_FallbackOrchestrator_ShouldAnnounceStopOnStop(t *testing.T) {
	t.Fatal("Not implemented")
}

func Test_AllOrchestrator_ShouldNotBuildWithEmptyTierList(t *testing.T) {
	/*_, err := NewAllOrchestrator()
	if err == nil {
		t.Fatal("should have failed to build")
	}*/
}

func Test_AllOrchestrator_ShouldAnnounceOnAllTiers(t *testing.T) {
	/*ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tiers := make([]ITierAnnouncer, 0)
	latchs := make([]*congo.CountDownLatch, 0)

	for i := 0; i < 5; i++ {
		latch := congo.NewCountDownLatch(1)
		tier := NewMockITierAnnouncer(ctrl)
		c := make(chan tierState)
		tier.EXPECT().States().Return(c).AnyTimes()
		tier.EXPECT().stopAnnounceLoop().AnyTimes()
		tier.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) {
			c <- tierState(ALIVE)
			latch.CountDown()
		}).Times(1)

		tiers = append(tiers, tier)
		latchs = append(latchs, latch)
	}

	o, _ := NewAllOrchestrator(tiers...)
	o.Start(nil)
	defer o.Stop(context.Background())

	for i, latch := range latchs {
		if !latch.WaitTimeout(500 * time.Millisecond) {
			t.Fatalf("latch has not been released at index %d", i)
		}
	}*/
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
	/*ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	o, _ := NewAllOrchestrator(NewMockITierAnnouncer(ctrl))

	latch := congo.NewCountDownLatch(1)
	go func () {
		o.Stop(context.Background())
		o.Stop(context.Background())
		o.Stop(context.Background())
		o.Stop(context.Background())
		o.Stop(context.Background())
		o.Stop(context.Background())
		latch.CountDown()
	}()

	if !latch.WaitTimeout(500 * time.Millisecond) {
		t.Fatal("Should not have blocked")
	}*/
}

func Test_AllOrchestrator_ShouldBeSafeToRunWithTremendousAmountOfTiers(t *testing.T) {
	/*ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tiers := make([]ITierAnnouncer, 0)

	latch := congo.NewCountDownLatch(10000)

	for i := 0; i < 3000; i++ {
		tier := NewMockITierAnnouncer(ctrl)
		c := make(chan tierState)
		tier.EXPECT().States().Return(c).AnyTimes()
		tier.EXPECT().stopAnnounceLoop().AnyTimes()
		tier.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) {
			c <- tierState(ALIVE)
			latch.CountDown()
		}).AnyTimes()
		tier.EXPECT().LastKnownInterval().Return(1 * time.Millisecond, nil).AnyTimes()

		tiers = append(tiers, tier)
	}

	o, _ := NewAllOrchestrator(tiers...)
	o.Start(nil)
	defer o.Stop(context.Background())

	if !latch.WaitTimeout(500 * time.Millisecond) {
		t.Fatal("latch has not been released")
	}*/
}

func Test_AllOrchestrator_ShouldBeReusableAfterStop(t *testing.T) {
	/*ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tiers := make([]ITierAnnouncer, 0)

	t1 := NewMockITierAnnouncer(ctrl)
	c1 := make(chan tierState)
	t1.EXPECT().States().Return(c1).AnyTimes()
	t1.EXPECT().stopAnnounceLoop().AnyTimes()
	tiers = append(tiers, t1)

	latch := congo.NewCountDownLatch(1)
	t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) {
		c1 <- tierState(ALIVE)
		latch.CountDown()
	}).Times(1)

	o, _ := NewAllOrchestrator(tiers...)
	o.Start(nil)

	if !latch.WaitTimeout(500 * time.Millisecond) {
		t.Fatal("latch has not been released")
	}

	o.Stop(context.Background())

	latch = congo.NewCountDownLatch(1)
	t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).Do(func(annFunc AnnouncingFunction, e tracker.AnnounceEvent) {
		c1 <- tierState(ALIVE)
		latch.CountDown()
	}).Times(1)

	o.Start(nil)
	o.Stop(context.Background())*/
}

func Test_AllOrchestrator_ShouldAnnounceStopOnStop(t *testing.T) {
	t.Fatal("Not implemented")
}
