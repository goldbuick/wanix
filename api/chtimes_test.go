package api

import (
	"math"
	"testing"
	"time"
)

func TestAsfloat64(t *testing.T) {
	cases := []struct {
		in   any
		want float64
		ok   bool
	}{
		{float64(1.5), 1.5, true},
		{float32(2), 2, true},
		{int(3), 3, true},
		{int64(4), 4, true},
		{uint64(5), 5, true},
		{"nope", 0, false},
		{nil, 0, false},
	}
	for _, tc := range cases {
		got, ok := asfloat64(tc.in)
		if ok != tc.ok {
			t.Fatalf("asfloat64(%T)=(%v,%v) want ok=%v", tc.in, got, ok, tc.ok)
		}
		if ok && got != tc.want {
			t.Fatalf("asfloat64(%T)=%v want %v", tc.in, got, tc.want)
		}
	}
}

func TestSecstofloattime(t *testing.T) {
	sec := 1700000000.25
	got := secstofloattime(sec)
	want := time.Unix(1700000000, int64(0.25*1e9))
	if !got.Equal(want) {
		t.Fatalf("secstofloattime(%v)=%v want %v", sec, got, want)
	}
	if math.Abs(float64(got.UnixNano()-want.UnixNano())) > 1 {
		t.Fatalf("nsec mismatch: %d vs %d", got.UnixNano(), want.UnixNano())
	}
}
