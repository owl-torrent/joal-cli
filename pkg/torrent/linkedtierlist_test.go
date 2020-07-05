package torrent

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func Test_linkedTierList_isFirst(t *testing.T) {
	type fields struct {
		current      ITierAnnouncer
		currentIndex uint
		list         []ITierAnnouncer
		lock         *sync.RWMutex
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{name: "shouldBeFirst", want: true, fields: fields{currentIndex: 0, lock: &sync.RWMutex{}}},
		{name: "shouldNotBeFirst", want: false, fields: fields{currentIndex: 1, lock: &sync.RWMutex{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := linkedTierList{
				ITierAnnouncer: tt.fields.current,
				currentIndex:   tt.fields.currentIndex,
				list:           tt.fields.list,
				lock:           tt.fields.lock,
			}
			if got := l.isFirst(); got != tt.want {
				t.Errorf("isFirst() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_linkedTierList_ShouldGoToNextAndFinalyGetBackToFirst(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var ts []ITierAnnouncer
	for i := 0; i < 50; i++ {
		ts = append(ts, NewMockITierAnnouncer(ctrl))
	}

	list, _ := newLinkedTierList(ts)

	assert.Equal(t, list.currentIndex, uint(0))
	assert.Same(t, list.ITierAnnouncer, ts[0])
	for i := 1; i < len(ts); i++ {
		list.next()
		assert.Equal(t, list.currentIndex, uint(i))
		assert.Same(t, list.ITierAnnouncer, ts[i])
	}

	list.next()
	assert.Equal(t, list.currentIndex, uint(0))
	assert.Same(t, list.ITierAnnouncer, ts[0])
}

func Test_linkedTierList_ShouldForwardToFirst(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var ts []ITierAnnouncer
	for i := 0; i < 50; i++ {
		ts = append(ts, NewMockITierAnnouncer(ctrl))
	}

	list, _ := newLinkedTierList(ts)
	list.backToFirst()

	assert.Equal(t, list.currentIndex, uint(0))
	assert.Same(t, list.ITierAnnouncer, ts[0])
	for i := 1; i < len(ts)/2; i++ {
		list.next()
		assert.Equal(t, list.currentIndex, uint(i))
		assert.Same(t, list.ITierAnnouncer, ts[i])
	}
	list.backToFirst()

	assert.Equal(t, list.currentIndex, uint(0))
	assert.Same(t, list.ITierAnnouncer, ts[0])
}

func Test_linkedTierList_ShouldWorkWithOnlyOneEntry(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ts := []ITierAnnouncer{NewMockITierAnnouncer(ctrl)}

	list, _ := newLinkedTierList(ts)

	assert.Equal(t, list.currentIndex, uint(0))
	assert.Same(t, list.ITierAnnouncer, ts[0])

	list.next()
	assert.Equal(t, list.currentIndex, uint(0))
	assert.Same(t, list.ITierAnnouncer, ts[0])

	list.backToFirst()
	assert.Equal(t, list.currentIndex, uint(0))
	assert.Same(t, list.ITierAnnouncer, ts[0])
}
