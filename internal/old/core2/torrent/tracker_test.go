package torrent

import (
	trackerlib "github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/internal/old/utils/testutils"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
	"time"
)

func TestTracker_getTier(t1 *testing.T) {
	type fields struct {
		tier int
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{name: "tier 1", fields: fields{tier: 1}, want: 1},
		{name: "tier 22", fields: fields{tier: 22}, want: 22},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &tracker{
				tier: tt.fields.tier,
			}
			assert.Equalf(t1, tt.want, t.getTier(), "getTier()")
		})
	}
}

func TestTracker_isAnnouncing(t1 *testing.T) {
	type fields struct {
		isCurrentlyAnnouncing bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{name: "notAnnouncing", fields: fields{isCurrentlyAnnouncing: false}, want: false},
		{name: "announcing", fields: fields{isCurrentlyAnnouncing: true}, want: true},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &tracker{
				isCurrentlyAnnouncing: tt.fields.isCurrentlyAnnouncing,
			}
			assert.Equalf(t1, tt.want, t.isAnnouncing(), "isAnnouncing()")
		})
	}
}

func TestTracker_canAnnounce(t1 *testing.T) {
	type fields struct {
		disabled              trackerDisabled
		nextAnnounce          time.Time
		isCurrentlyAnnouncing bool
	}
	type args struct {
		at time.Time
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{name: "disabled && time elapsed && not announcing", fields: fields{
			disabled:              trackerDisabled{disabled: true},
			nextAnnounce:          time.Now().Add(-1 * time.Hour),
			isCurrentlyAnnouncing: false,
		}, args: args{at: time.Now()}, want: false},
		{name: "disabled && time future && not announcing", fields: fields{
			disabled:              trackerDisabled{disabled: true},
			nextAnnounce:          time.Now().Add(5 * time.Minute),
			isCurrentlyAnnouncing: false,
		}, args: args{at: time.Now()}, want: false},
		{name: "disabled && time elapsed && announcing", fields: fields{
			disabled:              trackerDisabled{disabled: true},
			nextAnnounce:          time.Now().Add(-1 * time.Hour),
			isCurrentlyAnnouncing: true,
		}, args: args{at: time.Now()}, want: false},
		{name: "disabled && time elapsed && not announcing", fields: fields{
			disabled:              trackerDisabled{disabled: true},
			nextAnnounce:          time.Now().Add(-1 * time.Hour),
			isCurrentlyAnnouncing: false,
		}, args: args{at: time.Now()}, want: false},
		{name: "not disabled && time elapsed && announcing", fields: fields{
			disabled:              trackerDisabled{disabled: false},
			nextAnnounce:          time.Now().Add(-1 * time.Hour),
			isCurrentlyAnnouncing: true,
		}, args: args{at: time.Now()}, want: false},
		{name: "not disabled && time future && not announcing", fields: fields{
			disabled:              trackerDisabled{disabled: false},
			nextAnnounce:          time.Now().Add(5 * time.Minute),
			isCurrentlyAnnouncing: false,
		}, args: args{at: time.Now()}, want: false},
		{name: "not disabled && time elapsed && not announcing", fields: fields{
			disabled:              trackerDisabled{disabled: false},
			nextAnnounce:          time.Now().Add(-1 * time.Hour),
			isCurrentlyAnnouncing: false,
		}, args: args{at: time.Now()}, want: true},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &tracker{
				disabled:              tt.fields.disabled,
				nextAnnounce:          tt.fields.nextAnnounce,
				isCurrentlyAnnouncing: tt.fields.isCurrentlyAnnouncing,
			}
			assert.Equalf(t1, tt.want, t.canAnnounce(tt.args.at), "canAnnounce()")
		})
	}
}

func TestTracker_announce(t1 *testing.T) {
	t := tracker{
		url:          *testutils.MustParseUrl("http://localhost:8081"),
		nextAnnounce: time.Now(),
		startSent:    true,
	}

	var req TrackerAnnounceRequest
	err := t.announce(trackerlib.None, func(request TrackerAnnounceRequest) {
		req = request
	})
	assert.NoError(t1, err)

	assert.Equal(t1, true, t.isCurrentlyAnnouncing)
	assert.Equal(t1, "http://localhost:8081", req.Url.String())
	assert.Equal(t1, trackerlib.None, req.Event)
}

func TestTracker_announce_shouldReplaceAnnounceNoneWithStartedIfNotStartedYet(t1 *testing.T) {
	t := tracker{
		url:          *testutils.MustParseUrl("http://localhost:8081"),
		nextAnnounce: time.Now(),
		startSent:    false,
	}

	var req TrackerAnnounceRequest
	err := t.announce(trackerlib.None, func(request TrackerAnnounceRequest) {
		req = request

	})
	assert.NoError(t1, err)

	assert.Equal(t1, trackerlib.Started, req.Event)
}

func TestTracker_announce_shouldNotAnnounceIfDisabled(t1 *testing.T) {
	t := tracker{
		url:          *testutils.MustParseUrl("http://localhost:8081"),
		nextAnnounce: time.Now(),
		disabled:     trackerDisabled{disabled: true},
	}

	err := t.announce(trackerlib.None, func(request TrackerAnnounceRequest) {})
	assert.Error(t1, err)
	assert.Contains(t1, err.Error(), "tracker is disabled")
}

func TestTracker_announce_shouldNotAnnounceIfEventIsStoppedAndNeverSentStart(t1 *testing.T) {
	t := tracker{
		url:          *testutils.MustParseUrl("http://localhost:8081"),
		nextAnnounce: time.Now(),
		startSent:    false,
	}

	hasAnnounced := false
	err := t.announce(trackerlib.Stopped, func(request TrackerAnnounceRequest) {
		hasAnnounced = true
	})
	assert.NoError(t1, err)
	assert.False(t1, hasAnnounced)
	assert.False(t1, t.isCurrentlyAnnouncing)
}

func TestTracker_announce_shouldAllowAnnounceIfNextAnnounceIsNotElapsed(t1 *testing.T) {
	t := tracker{
		url:          *testutils.MustParseUrl("http://localhost:8081"),
		nextAnnounce: time.Now().Add(1 * time.Hour),
	}
	hasAnnounced := false
	err := t.announce(trackerlib.None, func(request TrackerAnnounceRequest) {
		hasAnnounced = true
	})
	assert.NoError(t1, err)
	assert.True(t1, hasAnnounced)
}

func TestTracker_announceSucceed(t1 *testing.T) {
	t := tracker{
		nextAnnounce:          time.Now().Add(-5 * time.Hour),
		isCurrentlyAnnouncing: true,
		announcesHistory:      []announceHistory{},
		consecutiveFails:      10,
	}

	at := time.Now()

	t.announceSucceed(announceHistory{
		at:       at,
		interval: 3520 * time.Second,
		seeders:  50,
		leechers: 50,
		error:    "",
	})

	assert.False(t1, t.isCurrentlyAnnouncing)
	assert.Equal(t1, 0, t.consecutiveFails)
	assert.Equal(t1, at.Add(3520*time.Second), t.nextAnnounce)
	assert.Len(t1, t.announcesHistory, 1)
}

func TestTracker_announceFailed(t1 *testing.T) {
	t := tracker{
		nextAnnounce:          time.Now().Add(-5 * time.Hour),
		isCurrentlyAnnouncing: true,
		announcesHistory:      []announceHistory{},
		consecutiveFails:      6,
	}

	at := time.Now()

	t.announceFailed(announceHistory{
		at:    at,
		error: "failed",
	})

	assert.False(t1, t.isCurrentlyAnnouncing)
	assert.Equal(t1, 7, t.consecutiveFails)
	assert.True(t1, t.nextAnnounce.After(at))
	assert.Len(t1, t.announcesHistory, 1)
}

func TestTracker_announceFailed_shouldNotOverflowConsecutiveFails(t1 *testing.T) {
	t := tracker{
		isCurrentlyAnnouncing: true,
		announcesHistory:      []announceHistory{},
		consecutiveFails:      math.MaxInt,
	}

	t.announceFailed(announceHistory{
		at:    time.Now(),
		error: "failed",
	})

	assert.Equal(t1, math.MaxInt, t.consecutiveFails)
}

func TestTracker_announceFailed_ShouldIncrementAnnounceIntervalEachTimeItFailsAndBeCappedAtMaximumRetryDelay_IfAnnounceHistoryDoesNotContainsInterval(t1 *testing.T) {
	t := tracker{
		nextAnnounce:          time.Now().Add(-5 * time.Hour),
		isCurrentlyAnnouncing: true,
		announcesHistory:      []announceHistory{},
	}

	at := time.Now()

	// delta between next announce time and time.Now
	var nextAnnounceDeltas []time.Duration

	for i := 0; i < 100; i++ {
		t.announceFailed(announceHistory{
			at:    at,
			error: "failed",
		})
		nextAnnounceDeltas = append(nextAnnounceDeltas, t.nextAnnounce.Sub(time.Now()))
	}

	for i := 1; i < len(nextAnnounceDeltas); i++ {
		if nextAnnounceDeltas[i] != maximumRetryDelay { // reached the peak amount (maximumRetryDelay)
			assert.Greater(t1, math.Ceil(nextAnnounceDeltas[i].Seconds()), nextAnnounceDeltas[i-1].Seconds())
		}
		assert.LessOrEqual(t1, math.Ceil(nextAnnounceDeltas[i].Seconds()), maximumRetryDelay.Seconds())
	}
}

func TestTracker_announceFailed_ShouldIncrementAnnounceIntervalEachTimeItFailsAndBeCappedAtMaximumRetryDelayButShouldPreferIntervalIfGreater(t1 *testing.T) {
	t := tracker{
		nextAnnounce:          time.Now().Add(-5 * time.Hour),
		isCurrentlyAnnouncing: true,
		announcesHistory:      []announceHistory{},
	}

	at := time.Now()

	var nextAnnounceDeltas []time.Duration

	interval := 25 * time.Minute
	for i := 0; i < 100; i++ {
		t.announceFailed(announceHistory{
			at:       at,
			interval: interval,
			error:    "failed",
		})
		nextAnnounceDeltas = append(nextAnnounceDeltas, t.nextAnnounce.Sub(time.Now()))
	}

	for i := 1; i < len(nextAnnounceDeltas); i++ {
		// never lower than interval
		assert.GreaterOrEqual(t1, math.Ceil(nextAnnounceDeltas[i].Seconds()), interval.Seconds())
		if nextAnnounceDeltas[i] != interval && nextAnnounceDeltas[i] != maximumRetryDelay {
			assert.Greater(t1, math.Ceil(nextAnnounceDeltas[i].Seconds()), nextAnnounceDeltas[i-1].Seconds())
		}
		assert.LessOrEqual(t1, math.Ceil(nextAnnounceDeltas[i].Seconds()), maximumRetryDelay.Seconds())
	}
}

func TestTracker_appendToAnnounceHistory_ShouldNotStoreMoreThanXEntries(t1 *testing.T) {
	var history []announceHistory

	maxLen := 6

	for i := 0; i < maxLen+2; i++ {
		history = appendToAnnounceHistory(history, announceHistory{
			at:    time.Now(),
			error: "failed !",
		}, maxLen)
	}

	assert.Len(t1, history, maxLen)
}
