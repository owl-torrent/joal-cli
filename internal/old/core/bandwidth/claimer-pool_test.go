package bandwidth

/*
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
	first := calculateWeight(&testSwarm{seeders: 10, leechers: 10})
	second := calculateWeight(&testSwarm{seeders: 10, leechers: 30})
	third := calculateWeight(&testSwarm{seeders: 10, leechers: 100})
	fourth := calculateWeight(&testSwarm{seeders: 10, leechers: 200})
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
	assert.InDelta(t, 11.1, calculateWeight(&testSwarm{seeders: 2, leechers: 1}), 0.1)
	assert.InDelta(t, 0.104058273, calculateWeight(&testSwarm{seeders: 30, leechers: 1}), 0.00000001)
	assert.InDelta(t, 9611.687812, calculateWeight(&testSwarm{seeders: 2, leechers: 100}), 0.0001)
	assert.InDelta(t, 73.01243916, calculateWeight(&testSwarm{seeders: 2000, leechers: 150}), 0.00001)
	assert.InDelta(t, 173066.5224, calculateWeight(&testSwarm{seeders: 150, leechers: 2000}), 0.01)
	assert.InDelta(t, 184911.2426, calculateWeight(&testSwarm{seeders: 80, leechers: 2000}), 0.1)
}

func TestClaimerPool_ShouldCalculateWeightAndAddToTotalWeightOnAdd(t *testing.T) {
	pool := NewWeightedClaimerPool()
	c1 := &mockedBandwidthClaimable{
		infoHash:    metainfo.NewHashFromHex("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"),
		uploaded:    0,
		swarm:       &dumbSwarm{seeders: 10, leechers: 10},
		addUploaded: nil,
	}
	c2 := &mockedBandwidthClaimable{
		infoHash:    metainfo.NewHashFromHex("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAB"),
		uploaded:    0,
		swarm:       &dumbSwarm{seeders: 10, leechers: 10000},
		addUploaded: nil,
	}
	pool.AddOrUpdate(c1)
	pool.AddOrUpdate(c2)

	assert.Greater(t, pool.claimers[c1.InfoHash()].weight, 0.0)
	assert.Greater(t, pool.claimers[c2.InfoHash()].weight, 0.0)

	assert.Equal(t, pool.totalWeight, pool.claimers[c1.InfoHash()].weight+pool.claimers[c2.InfoHash()].weight)
}

func TestClaimerPool_ShouldRemoveFromTotalWeightOnRemove(t *testing.T) {
	pool := NewWeightedClaimerPool()
	c1 := &mockedBandwidthClaimable{
		infoHash:    metainfo.NewHashFromHex("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"),
		uploaded:    0,
		swarm:       &dumbSwarm{seeders: 10, leechers: 10},
		addUploaded: nil,
	}
	c2 := &mockedBandwidthClaimable{
		infoHash:    metainfo.NewHashFromHex("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAB"),
		uploaded:    0,
		swarm:       &dumbSwarm{seeders: 10, leechers: 10000},
		addUploaded: nil,
	}
	pool.AddOrUpdate(c1)
	pool.AddOrUpdate(c2)

	assert.Greater(t, pool.totalWeight, 0.0)
	weightBeforeDelete := pool.totalWeight

	pool.RemoveFromPool(c2)

	assert.Less(t, pool.totalWeight, weightBeforeDelete)
	assert.Greater(t, pool.totalWeight, 0.0)
}

func TestClaimerPool_ShouldReturnWeights(t *testing.T) {
	pool := NewWeightedClaimerPool()
	c1 := &mockedBandwidthClaimable{
		infoHash:    metainfo.NewHashFromHex("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"),
		uploaded:    0,
		swarm:       &dumbSwarm{seeders: 10, leechers: 10},
		addUploaded: nil,
	}
	c2 := &mockedBandwidthClaimable{
		infoHash:    metainfo.NewHashFromHex("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAB"),
		uploaded:    0,
		swarm:       &dumbSwarm{seeders: 10, leechers: 10000},
		addUploaded: nil,
	}
	pool.AddOrUpdate(c1)
	pool.AddOrUpdate(c2)

	weights, totalWeight := pool.GetWeights()

	assert.Len(t, weights, 2)
	assert.Equal(t, weights[0].weight, pool.claimers[weights[0].InfoHash()].weight)
	assert.Equal(t, weights[1].weight, pool.claimers[weights[1].InfoHash()].weight)

	assert.Equal(t, totalWeight, pool.claimers[c1.InfoHash()].weight+pool.claimers[c2.InfoHash()].weight)
	assert.Greater(t, pool.claimers[c1.InfoHash()].weight, 0.0)
	assert.Greater(t, pool.claimers[c2.InfoHash()].weight, 0.0)
}*/
