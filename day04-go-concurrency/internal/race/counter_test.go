package race

import "testing"

func TestCountersAreCorrect(t *testing.T) {
	tests := []struct {
		name string
		c    Counter
	}{
		{"mutex", &MutexCounter{}},
		{"atomic", &AtomicCounter{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := Hammer(tc.c, 100, 1000); got != 100000 {
				t.Fatalf("got %d want 100000", got)
			}
		})
	}
}

func TestHammerZero(t *testing.T) {
	if got := Hammer(&AtomicCounter{}, 0, 10); got != 0 {
		t.Fatalf("got %d", got)
	}
}
