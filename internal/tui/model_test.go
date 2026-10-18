package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

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

func TestModel_View(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		d      Deps
		width  int
		height int
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(tt.d, tt.width, tt.height)
			got := m.View()
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("View() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModel_Update(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		d      Deps
		width  int
		height int
		// Named input parameters for target function.
		msg   tea.Msg
		want  tea.Model
		want2 tea.Cmd
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(tt.d, tt.width, tt.height)
			got, got2 := m.Update(tt.msg)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("Update() = %v, want %v", got, tt.want)
			}
			if true {
				t.Errorf("Update() = %v, want %v", got2, tt.want2)
			}
		})
	}
}
