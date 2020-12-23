package orchestrator

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTierStateCalculator_ShouldNotPublishEventCreation(t *testing.T) {
	providers := []ITrackerAnnouncer{
		&mockedTrackerAnnouncer{},
		&mockedTrackerAnnouncer{},
		&mockedTrackerAnnouncer{},
		&mockedTrackerAnnouncer{},
	}

	calculator := NewTierStateCalculator(providers)

	assertEmpty(t, calculator.States())
}

func TestTierStateCalculator_ShouldPublishAliveOnFirstSuccess(t *testing.T) {
	providers := []ITrackerAnnouncer{
		&mockedTrackerAnnouncer{},
		&mockedTrackerAnnouncer{},
		&mockedTrackerAnnouncer{},
		&mockedTrackerAnnouncer{},
	}

	calculator := NewTierStateCalculator(providers)

	calculator.setIndividualState(providers[0], true)

	select {
	case st := <-calculator.States():
		assert.Equal(t, ALIVE, st)
	default:
		t.Fatal("channel should not be empty")
	}
}

func TestTierStateCalculator_ShouldPublishDeadWhenAllReportFail(t *testing.T) {
	providers := []ITrackerAnnouncer{
		&mockedTrackerAnnouncer{},
		&mockedTrackerAnnouncer{},
		&mockedTrackerAnnouncer{},
		&mockedTrackerAnnouncer{},
	}

	calculator := NewTierStateCalculator(providers)

	calculator.setIndividualState(providers[0], false)
	assertEmpty(t, calculator.States())
	calculator.setIndividualState(providers[1], false)
	assertEmpty(t, calculator.States())
	calculator.setIndividualState(providers[2], false)
	assertEmpty(t, calculator.States())
	calculator.setIndividualState(providers[3], false)

	select {
	case st := <-calculator.States():
		assert.Equal(t, DEAD, st)
	default:
		t.Fatal("should not send to channel on creation")
	}
}

func TestTierStateCalculator_ShouldPublishAliveAfterDeadIfOneReportSuccess(t *testing.T) {
	providers := []ITrackerAnnouncer{
		&mockedTrackerAnnouncer{},
		&mockedTrackerAnnouncer{},
	}

	calculator := NewTierStateCalculator(providers)

	calculator.setIndividualState(providers[0], false)
	assertEmpty(t, calculator.States())
	calculator.setIndividualState(providers[1], false)

	select {
	case st := <-calculator.States():
		assert.Equal(t, DEAD, st)
	default:
		t.Fatal("chan should not be empty")
	}

	calculator.setIndividualState(providers[0], true)

	select {
	case st := <-calculator.States():
		assert.Equal(t, ALIVE, st)
	default:
		t.Fatal("chan should not be empty")
	}
}

func assertEmpty(t *testing.T, c <-chan tierState) {
	select {
	case <-c:
		t.Fatal("channel should be empty")
	default:
	}
}
