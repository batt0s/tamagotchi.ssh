package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
)

func Test_teaHandler(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		s     ssh.Session
		want  tea.Model
		want2 []tea.ProgramOption
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got2 := teaHandler(tt.s)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("teaHandler() = %v, want %v", got, tt.want)
			}
			if true {
				t.Errorf("teaHandler() = %v, want %v", got2, tt.want2)
			}
		})
	}
}
