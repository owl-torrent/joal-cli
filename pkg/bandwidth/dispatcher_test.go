package bandwidth

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type testSwarm struct {
	seeders  uint64
	leechers uint64
}

func (s *testSwarm) getSeeders() uint64 {
	return s.seeders
}
func (s *testSwarm) getLeechers() uint64 {
	return s.leechers
}

func Test_calculateWeightShouldNeverGoBelowZero(t *testing.T) {
	type args struct {
		swarm ISwarm
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{name: "with0Leechers0Seeders", args: args{swarm: &testSwarm{seeders: 0, leechers: 0}}, want: 0},
		{name: "with0Leechers1Seeders", args: args{swarm: &testSwarm{seeders: 1, leechers: 0}}, want: 0},
		{name: "with1Leechers0Seeders", args: args{swarm: &testSwarm{seeders: 0, leechers: 1}}, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateWeight(tt.args.swarm); got != tt.want {
				t.Errorf("calculateWeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calculateWeightShouldNotFailIfSwarmIsNil(t *testing.T) {
	assert.Zero(t, calculateWeight(nil))
}

func Test_calculateWeightShouldPromoteSwarmWithMoreLeechers(t *testing.T) {
	first := calculateWeight(&testSwarm{seeders: 10, leechers: 10});
	second := calculateWeight(&testSwarm{seeders: 10, leechers: 30});
	third := calculateWeight(&testSwarm{seeders: 10, leechers: 100});
	fourth := calculateWeight(&testSwarm{seeders: 10, leechers: 200});
	assert.Greater(t, fourth, third, "should be greater")
	assert.Greater(t, third, second, "should be greater")
	assert.Greater(t, second, first, "should be greater")
}

func Test_calculateWeightShouldProvidePreciseValues(t *testing.T) {
	assert.Zero(t, calculateWeight(&testSwarm{seeders: 0, leechers: 1}))
	assert.Zero(t, calculateWeight(&testSwarm{seeders: 1, leechers: 0}))
	assert.Zero(t, calculateWeight(&testSwarm{seeders: 0, leechers: 100}))
	assert.Zero(t, calculateWeight(&testSwarm{seeders: 0, leechers: 0}))
	assert.Equal(t, 25.0, calculateWeight(&testSwarm{seeders: 1, leechers: 1}))
	assert.Equal(t, 50000.0, calculateWeight(&testSwarm{seeders: 2000, leechers: 2000}))
	assert.InDelta(t, float64(11.1), calculateWeight(&testSwarm{seeders: 2, leechers: 1}), 0.1)
	assert.InDelta(t, float64(0.104058273), calculateWeight(&testSwarm{seeders: 30, leechers: 1}), 0.00000001)
	assert.InDelta(t, float64(9611.687812), calculateWeight(&testSwarm{seeders: 2, leechers: 100}), 0.0001)
	assert.InDelta(t, float64(73.01243916), calculateWeight(&testSwarm{seeders: 2000, leechers: 150}), 0.00001)
	assert.InDelta(t, float64(173066.5224), calculateWeight(&testSwarm{seeders: 150, leechers: 2000}), 0.01)
	assert.InDelta(t, float64(184911.2426), calculateWeight(&testSwarm{seeders: 80, leechers: 2000}), 0.1)
}
