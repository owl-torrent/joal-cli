package torrent

import (
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
			t := &Tracker{
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
			t := &Tracker{
				isCurrentlyAnnouncing: tt.fields.isCurrentlyAnnouncing,
			}
			assert.Equalf(t1, tt.want, t.isAnnouncing(), "isAnnouncing()")
		})
	}
}

func TestTracker_canAnnounce(t1 *testing.T) {
	type fields struct {
		disabled              TrackerDisabled
		nextAnnounce          time.Time
		isCurrentlyAnnouncing bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{name: "disabled && time elapsed && not announcing", fields: fields{
			disabled:              TrackerDisabled{disabled: true},
			nextAnnounce:          time.Now().Add(-1 * time.Hour),
			isCurrentlyAnnouncing: false,
		}, want: false},
		{name: "disabled && time future && not announcing", fields: fields{
			disabled:              TrackerDisabled{disabled: true},
			nextAnnounce:          time.Now().Add(5 * time.Minute),
			isCurrentlyAnnouncing: false,
		}, want: false},
		{name: "disabled && time elapsed && announcing", fields: fields{
			disabled:              TrackerDisabled{disabled: true},
			nextAnnounce:          time.Now().Add(-1 * time.Hour),
			isCurrentlyAnnouncing: true,
		}, want: false},
		{name: "disabled && time elapsed && not announcing", fields: fields{
			disabled:              TrackerDisabled{disabled: true},
			nextAnnounce:          time.Now().Add(-1 * time.Hour),
			isCurrentlyAnnouncing: false,
		}, want: false},
		{name: "not disabled && time elapsed && announcing", fields: fields{
			disabled:              TrackerDisabled{disabled: false},
			nextAnnounce:          time.Now().Add(-1 * time.Hour),
			isCurrentlyAnnouncing: true,
		}, want: false},
		{name: "not disabled && time future && not announcing", fields: fields{
			disabled:              TrackerDisabled{disabled: false},
			nextAnnounce:          time.Now().Add(5 * time.Minute),
			isCurrentlyAnnouncing: false,
		}, want: false},
		{name: "not disabled && time elapsed && not announcing", fields: fields{
			disabled:              TrackerDisabled{disabled: false},
			nextAnnounce:          time.Now().Add(-1 * time.Hour),
			isCurrentlyAnnouncing: false,
		}, want: true},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Tracker{
				disabled:              tt.fields.disabled,
				nextAnnounce:          tt.fields.nextAnnounce,
				isCurrentlyAnnouncing: tt.fields.isCurrentlyAnnouncing,
			}
			assert.Equalf(t1, tt.want, t.canAnnounce(), "canAnnounce()")
		})
	}
}

func TestTracker_HasAnnouncedStart(t1 *testing.T) {
	type fields struct {
		startSent bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{name: "not sent", fields: fields{startSent: false}, want: false},
		{name: "sent", fields: fields{startSent: true}, want: true},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Tracker{
				startSent: tt.fields.startSent,
			}
			assert.Equalf(t1, tt.want, t.HasAnnouncedStart(), "HasAnnouncedStart()")
		})
	}
}

func TestTracker_announcing(t1 *testing.T) {
	t := Tracker{isCurrentlyAnnouncing: false}
	t.announcing()

	assert.True(t1, t.isCurrentlyAnnouncing)
}

func TestTracker_announceSucceed(t1 *testing.T) {
	t := Tracker{
		nextAnnounce:          time.Now().Add(-5 * time.Hour),
		isCurrentlyAnnouncing: true,
		announcesHistory:      []AnnounceHistory{},
		consecutiveFails:      10,
	}

	at := time.Now()

	t.announceSucceed(AnnounceHistory{
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
	t := Tracker{
		nextAnnounce:          time.Now().Add(-5 * time.Hour),
		isCurrentlyAnnouncing: true,
		announcesHistory:      []AnnounceHistory{},
		consecutiveFails:      6,
	}

	at := time.Now()

	t.announceFailed(AnnounceHistory{
		at:    at,
		error: "failed",
	})

	assert.False(t1, t.isCurrentlyAnnouncing)
	assert.Equal(t1, 7, t.consecutiveFails)
	assert.True(t1, t.nextAnnounce.After(at))
	assert.Len(t1, t.announcesHistory, 1)
}

func TestTracker_announceFailed_shouldNotOverflowConsecutiveFails(t1 *testing.T) {
	t := Tracker{
		isCurrentlyAnnouncing: true,
		announcesHistory:      []AnnounceHistory{},
		consecutiveFails:      math.MaxInt,
	}

	t.announceFailed(AnnounceHistory{
		at:    time.Now(),
		error: "failed",
	})

	assert.Equal(t1, math.MaxInt, t.consecutiveFails)
}

func TestTracker_announceFailed_ShouldIncrementAnnounceIntervalEachTimeItFailsAndBeCappedAtMaximumRetryDelay_IfAnnounceHistoryDoesNotContainsInterval(t1 *testing.T) {
	t := Tracker{
		nextAnnounce:          time.Now().Add(-5 * time.Hour),
		isCurrentlyAnnouncing: true,
		announcesHistory:      []AnnounceHistory{},
	}

	at := time.Now()

	// delta between next announce time and time.Now
	nextAnnounceDeltas := []time.Duration{}

	for i := 0; i < 100; i++ {
		t.announceFailed(AnnounceHistory{
			at:    at,
			error: "failed",
		})
		nextAnnounceDeltas = append(nextAnnounceDeltas, t.nextAnnounce.Sub(time.Now()))
	}

	for i := 1; i < len(nextAnnounceDeltas); i++ {
		if nextAnnounceDeltas[i] != maximumRetryDelay { // reached the peak amount (maximumRetryDelay)
			assert.Greater(t1, nextAnnounceDeltas[i].Seconds(), nextAnnounceDeltas[i-1].Seconds())
		}
		assert.LessOrEqual(t1, nextAnnounceDeltas[i].Seconds(), maximumRetryDelay.Seconds())
	}
}

func TestTracker_announceFailed_ShouldIncrementAnnounceIntervalEachTimeItFailsAndBeCappedAtMaximumRetryDelayButShouldPreferIntervalIfGreater(t1 *testing.T) {
	t := Tracker{
		nextAnnounce:          time.Now().Add(-5 * time.Hour),
		isCurrentlyAnnouncing: true,
		announcesHistory:      []AnnounceHistory{},
	}

	at := time.Now()

	nextAnnounceDeltas := []time.Duration{}

	interval := 25 * time.Minute
	for i := 0; i < 100; i++ {
		t.announceFailed(AnnounceHistory{
			at:       at,
			interval: interval,
			error:    "failed",
		})
		nextAnnounceDeltas = append(nextAnnounceDeltas, t.nextAnnounce.Sub(time.Now()))
	}

	for i := 1; i < len(nextAnnounceDeltas); i++ {
		// never lower than interval
		assert.GreaterOrEqual(t1, nextAnnounceDeltas[i].Seconds(), interval.Seconds())
		if nextAnnounceDeltas[i] != interval && nextAnnounceDeltas[i] != maximumRetryDelay {
			assert.Greater(t1, nextAnnounceDeltas[i].Seconds(), nextAnnounceDeltas[i-1].Seconds())
		}
		assert.LessOrEqual(t1, nextAnnounceDeltas[i].Seconds(), maximumRetryDelay.Seconds())
	}
}

func TestTracker_appendToAnnounceHistory_ShouldNotStoreMoreThanXEntries(t1 *testing.T) {
	var history []AnnounceHistory

	maxLen := 6

	for i := 0; i < maxLen+2; i++ {
		history = appendToAnnounceHistory(history, AnnounceHistory{
			at:    time.Now(),
			error: "failed !",
		}, maxLen)
	}

	assert.Len(t1, history, maxLen)
}
