package tui

import "testing"

func Test_shorten(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		fp   string
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shorten(tt.fp)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("shorten() = %v, want %v", got, tt.want)
			}
		})
	}
}
