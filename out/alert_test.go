package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tklein1801/concourse-discord-alert-resource/concourse"
)

func TestNewAlert(t *testing.T) {
	cases := map[string]struct {
		input *concourse.OutRequest
		want  Alert
	}{
		// Default and overrides.
		"default": {
			input: &concourse.OutRequest{},
			want:  Alert{Type: "default", Color: "#35495c", IconURL: "https://ci.concourse-ci.org/public/images/favicon-pending.png"},
		},
		"custom params": {
			input: &concourse.OutRequest{
				Source: concourse.Source{},
				Params: concourse.OutParams{Color: "#ffffff", Message: "custom-message", Text: "custom-text", Disable: true, Role: "1342563020215291936"},
			},
			want: Alert{Type: "default", Color: "#ffffff", IconURL: "https://ci.concourse-ci.org/public/images/favicon-pending.png", Message: "custom-message", Text: "custom-text", Disabled: true, Role: "1342563020215291936"},
		},
		"custom source": {
			input: &concourse.OutRequest{
				Source: concourse.Source{Disable: true},
			},
			want: Alert{Type: "default", Color: "#35495c", IconURL: "https://ci.concourse-ci.org/public/images/favicon-pending.png", Disabled: true},
		},
		// Alert types.
		"success": {
			input: &concourse.OutRequest{Params: concourse.OutParams{AlertType: "success"}},
			want:  Alert{Type: "success", Color: "#32cd32", IconURL: "https://ci.concourse-ci.org/public/images/favicon-succeeded.png", Message: "Success"},
		},
		"failed": {
			input: &concourse.OutRequest{Params: concourse.OutParams{AlertType: "failed"}},
			want:  Alert{Type: "failed", Color: "#d00000", IconURL: "https://ci.concourse-ci.org/public/images/favicon-failed.png", Message: "Failed"},
		},
		"started": {
			input: &concourse.OutRequest{Params: concourse.OutParams{AlertType: "started"}},
			want:  Alert{Type: "started", Color: "#f7cd42", IconURL: "https://ci.concourse-ci.org/public/images/favicon-started.png", Message: "Started"},
		},
		"aborted": {
			input: &concourse.OutRequest{Params: concourse.OutParams{AlertType: "aborted"}},
			want:  Alert{Type: "aborted", Color: "#8d4b32", IconURL: "https://ci.concourse-ci.org/public/images/favicon-aborted.png", Message: "Aborted"},
		},
		"fixed": {
			input: &concourse.OutRequest{Params: concourse.OutParams{AlertType: "fixed"}},
			want:  Alert{Type: "fixed", Color: "#32cd32", IconURL: "https://ci.concourse-ci.org/public/images/favicon-succeeded.png", Message: "Fixed"},
		},
		"broke": {
			input: &concourse.OutRequest{Params: concourse.OutParams{AlertType: "broke"}},
			want:  Alert{Type: "broke", Color: "#d00000", IconURL: "https://ci.concourse-ci.org/public/images/favicon-failed.png", Message: "Broke"},
		},
		"errored": {
			input: &concourse.OutRequest{Params: concourse.OutParams{AlertType: "errored"}},
			want:  Alert{Type: "errored", Color: "#f5a623", IconURL: "https://ci.concourse-ci.org/public/images/favicon-errored.png", Message: "Errored"},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got := NewAlert(c.input)
			if !cmp.Equal(got, c.want) {
				t.Fatalf("unexpected Alert from NewAlert:\n\t(GOT): %#v\n\t(WNT): %#v\n\t(DIFF): %v", got, c.want, cmp.Diff(got, c.want))
			}
		})
	}
}
