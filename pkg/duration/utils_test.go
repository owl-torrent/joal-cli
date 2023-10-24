package duration

import (
	"testing"
	"time"
)

func TestMin(t *testing.T) {
	type args struct {
		d1 time.Duration
		d2 time.Duration
	}
	tests := []struct {
		name string
		args args
		want time.Duration
	}{
		{name: "bothEqual", args: args{d1: 1 * time.Second, d2: 1 * time.Second}, want: 1 * time.Second},
		{name: "firstGreater", args: args{d1: 50 * time.Second, d2: 1 * time.Second}, want: 1 * time.Second},
		{name: "secondGreater", args: args{d1: 1 * time.Second, d2: 50 * time.Second}, want: 1 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Min(tt.args.d1, tt.args.d2); got != tt.want {
				t.Errorf("Min() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMax(t *testing.T) {
	type args struct {
		d1 time.Duration
		d2 time.Duration
	}
	tests := []struct {
		name string
		args args
		want time.Duration
	}{
		{name: "bothEqual", args: args{d1: 1 * time.Second, d2: 1 * time.Second}, want: 1 * time.Second},
		{name: "firstGreater", args: args{d1: 50 * time.Second, d2: 1 * time.Second}, want: 50 * time.Second},
		{name: "secondGreater", args: args{d1: 1 * time.Second, d2: 50 * time.Second}, want: 50 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Max(tt.args.d1, tt.args.d2); got != tt.want {
				t.Errorf("Max() = %v, want %v", got, tt.want)
			}
		})
	}
}
