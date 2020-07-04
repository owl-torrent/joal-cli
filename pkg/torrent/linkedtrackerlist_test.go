package torrent

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func Test_linkedTrackerList_isFirst(t *testing.T) {
	type fields struct {
		current      ITrackerAnnouncer
		currentIndex uint
		list         []ITrackerAnnouncer
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
			l := linkedTrackerList{
				ITrackerAnnouncer: tt.fields.current,
				currentIndex:      tt.fields.currentIndex,
				list:              tt.fields.list,
				lock:              tt.fields.lock,
			}
			if got := l.isFirst(); got != tt.want {
				t.Errorf("isFirst() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_linkedTrackerList_ShouldGoToNextAndFinallyGetBackToFirst(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var ts []ITrackerAnnouncer
	for i := 0; i < 50; i++ {
		ts = append(ts, NewMockITrackerAnnouncer(ctrl))
	}

	list, _ := newLinkedTrackerList(ts)

	assert.Equal(t, list.currentIndex, uint(0))
	assert.Equal(t, list.ITrackerAnnouncer, ts[0])
	for i := 1; i < len(ts); i++ {
		list.next()
		assert.Equal(t, list.ITrackerAnnouncer, ts[i])
		assert.Equal(t, list.currentIndex, uint(i))
	}

	list.next()
	assert.Equal(t, list.currentIndex, uint(0))
	assert.Equal(t, list.ITrackerAnnouncer, ts[0])
}

func Test_linkedTrackerList_ShouldPromoteToFirst(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t1 := NewMockITrackerAnnouncer(ctrl)
	t2 := NewMockITrackerAnnouncer(ctrl)
	t3 := NewMockITrackerAnnouncer(ctrl)
	t4 := NewMockITrackerAnnouncer(ctrl)

	list, _ := newLinkedTrackerList([]ITrackerAnnouncer{t1, t2, t3, t4})

	assert.Equal(t, list.list[0], t1)
	assert.Equal(t, list.list[1], t2)
	assert.Equal(t, list.list[2], t3)
	assert.Equal(t, list.list[3], t4)

	list.next()
	list.PromoteCurrent()
	assert.Equal(t, list.currentIndex, uint(0))
	assert.Equal(t, list.list[0], t2)
	assert.Equal(t, list.list[1], t1)
	assert.Equal(t, list.list[2], t3)
	assert.Equal(t, list.list[3], t4)

	list.next()
	list.next()
	list.PromoteCurrent()
	assert.Equal(t, list.currentIndex, uint(0))
	assert.Equal(t, list.list[0], t3)
	assert.Equal(t, list.list[1], t2)
	assert.Equal(t, list.list[2], t1)
	assert.Equal(t, list.list[3], t4)
}

func Test_linkedTrackerList_ShouldPromoteToFirstWithSingleTracker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t1 := NewMockITrackerAnnouncer(ctrl)

	list, _ := newLinkedTrackerList([]ITrackerAnnouncer{t1})

	assert.Equal(t, list.list[0], t1)

	list.PromoteCurrent()
	assert.Equal(t, list.currentIndex, uint(0))
	assert.Equal(t, list.list[0], t1)
	assert.Equal(t, list.ITrackerAnnouncer, t1)

	list.next()
	list.PromoteCurrent()
	assert.Equal(t, list.currentIndex, uint(0))
	assert.Equal(t, list.list[0], t1)
	assert.Equal(t, list.ITrackerAnnouncer, t1)
}

func Test_linkedTrackerList_ShouldPromoteFirstToFirst(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t1 := NewMockITrackerAnnouncer(ctrl)
	t2 := NewMockITrackerAnnouncer(ctrl)
	t3 := NewMockITrackerAnnouncer(ctrl)
	t4 := NewMockITrackerAnnouncer(ctrl)

	list, _ := newLinkedTrackerList([]ITrackerAnnouncer{t1, t2, t3, t4})

	assert.Equal(t, list.list[0], t1)
	assert.Equal(t, list.list[1], t2)
	assert.Equal(t, list.list[2], t3)
	assert.Equal(t, list.list[3], t4)

	list.PromoteCurrent()
	assert.Equal(t, list.currentIndex, uint(0))
	assert.Equal(t, list.list[0], t1)
	assert.Equal(t, list.list[1], t2)
	assert.Equal(t, list.list[2], t3)
	assert.Equal(t, list.list[3], t4)
}

func Test_linkedTrackerList_ShouldWorkWithOnlyOneEntry(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ts := []ITrackerAnnouncer{NewMockITrackerAnnouncer(ctrl)}

	list, _ := newLinkedTrackerList(ts)

	assert.Equal(t, list.currentIndex, uint(0))
	assert.Equal(t, list.ITrackerAnnouncer, ts[0])

	list.next()
	assert.Equal(t, list.currentIndex, uint(0))
	assert.Equal(t, list.ITrackerAnnouncer, ts[0])
}
