package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/arbourd/concourse-slack-alert-resource/concourse"
	"github.com/arbourd/concourse-slack-alert-resource/slack"
)

func TestOut(t *testing.T) {
	ok := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ok.Close()
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer bad.Close()

	env := map[string]string{
		"ATC_EXTERNAL_URL":    "https://ci.example.com",
		"BUILD_TEAM_NAME":     "main",
		"BUILD_PIPELINE_NAME": "demo",
		"BUILD_JOB_NAME":      "test",
		"BUILD_NAME":          "2",
	}

	cases := map[string]struct {
		outRequest *concourse.OutRequest
		want       *concourse.OutResponse
		env        map[string]string
		err        bool
	}{
		"default alert": {
			outRequest: &concourse.OutRequest{
				Source: concourse.Source{URL: ok.URL},
			},
			want: &concourse.OutResponse{
				Version: concourse.Version{"ver": "static"},
				Metadata: []concourse.Metadata{
					{Name: "type", Value: "default"},
					{Name: "channel", Value: ""},
					{Name: "alerted", Value: "true"},
				},
			},
			env: env,
		},
		"success alert": {
			outRequest: &concourse.OutRequest{
				Source: concourse.Source{URL: ok.URL},
				Params: concourse.OutParams{AlertType: "success"},
			},
			want: &concourse.OutResponse{
				Version: concourse.Version{"ver": "static"},
				Metadata: []concourse.Metadata{
					{Name: "type", Value: "success"},
					{Name: "channel", Value: ""},
					{Name: "alerted", Value: "true"},
				},
			},
			env: env,
		},
		"failed alert": {
			outRequest: &concourse.OutRequest{
				Source: concourse.Source{URL: ok.URL},
				Params: concourse.OutParams{AlertType: "failed"},
			},
			want: &concourse.OutResponse{
				Version: concourse.Version{"ver": "static"},
				Metadata: []concourse.Metadata{
					{Name: "type", Value: "failed"},
					{Name: "channel", Value: ""},
					{Name: "alerted", Value: "true"},
				},
			},
			env: env,
		},
		"started alert": {
			outRequest: &concourse.OutRequest{
				Source: concourse.Source{URL: ok.URL},
				Params: concourse.OutParams{AlertType: "started"},
			},
			want: &concourse.OutResponse{
				Version: concourse.Version{"ver": "static"},
				Metadata: []concourse.Metadata{
					{Name: "type", Value: "started"},
					{Name: "channel", Value: ""},
					{Name: "alerted", Value: "true"},
				},
			},
			env: env,
		},
		"aborted alert": {
			outRequest: &concourse.OutRequest{
				Source: concourse.Source{URL: ok.URL},
				Params: concourse.OutParams{AlertType: "aborted"},
			},
			want: &concourse.OutResponse{
				Version: concourse.Version{"ver": "static"},
				Metadata: []concourse.Metadata{
					{Name: "type", Value: "aborted"},
					{Name: "channel", Value: ""},
					{Name: "alerted", Value: "true"},
				},
			},
			env: env,
		},
		"custom alert": {
			outRequest: &concourse.OutRequest{
				Source: concourse.Source{URL: ok.URL},
				Params: concourse.OutParams{
					AlertType: "non-existent-type",
					Message:   "Deploying",
					Color:     "#ffffff",
				},
			},
			want: &concourse.OutResponse{
				Version: concourse.Version{"ver": "static"},
				Metadata: []concourse.Metadata{
					{Name: "type", Value: "default"},
					{Name: "channel", Value: ""},
					{Name: "alerted", Value: "true"},
				},
			},
			env: env,
		},
		"override channel at Source": {
			outRequest: &concourse.OutRequest{
				Source: concourse.Source{URL: ok.URL, Channel: "#source"},
			},
			want: &concourse.OutResponse{
				Version: concourse.Version{"ver": "static"},
				Metadata: []concourse.Metadata{
					{Name: "type", Value: "default"},
					{Name: "channel", Value: "#source"},
					{Name: "alerted", Value: "true"},
				},
			},
			env: env,
		},
		"override channel at Params": {
			outRequest: &concourse.OutRequest{
				Source: concourse.Source{URL: ok.URL, Channel: "#source"},
				Params: concourse.OutParams{Channel: "#params"},
			},
			want: &concourse.OutResponse{
				Version: concourse.Version{"ver": "static"},
				Metadata: []concourse.Metadata{
					{Name: "type", Value: "default"},
					{Name: "channel", Value: "#params"},
					{Name: "alerted", Value: "true"},
				},
			},
			env: env,
		},
		"disable alert": {
			outRequest: &concourse.OutRequest{
				Source: concourse.Source{URL: bad.URL},
				Params: concourse.OutParams{Disable: true},
			},
			want: &concourse.OutResponse{
				Version: concourse.Version{"ver": "static"},
				Metadata: []concourse.Metadata{
					{Name: "type", Value: "default"},
					{Name: "channel", Value: ""},
					{Name: "alerted", Value: "false"},
				},
			},
			env: env,
		},
		"error without Slack URL": {
			outRequest: &concourse.OutRequest{
				Source: concourse.Source{URL: ""},
			},
			env: env,
			err: true,
		},
		"error with bad request": {
			outRequest: &concourse.OutRequest{
				Source: concourse.Source{URL: bad.URL},
			},
			env: env,
			err: true,
		},
		"error without basic auth for fixed type": {
			outRequest: &concourse.OutRequest{
				Source: concourse.Source{URL: ok.URL, Username: "", Password: ""},
				Params: concourse.OutParams{AlertType: "fixed"},
			},
			env: env,
			err: true,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			for k, v := range c.env {
				os.Setenv(k, v)
			}

			got, err := out(c.outRequest, "")
			if err != nil && !c.err {
				t.Fatalf("unexpected error from out:\n\t(ERR): %s", err)
			} else if err == nil && c.err {
				t.Fatalf("expected an error from out:\n\t(GOT): nil")
			} else if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("unexpected concourse.OutResponse value from out:\n\t(GOT): %#v\n\t(WNT): %#v", got, c.want)
			}
		})
	}
}
func TestBuildMessage(t *testing.T) {
	cases := map[string]struct {
		alert Alert
		want  *slack.Message
	}{
		"empty channel": {
			alert: Alert{
				Type:    "default",
				Color:   "#ffffff",
				IconURL: "",
				Message: "Testing",
			},
			want: &slack.Message{
				Attachments: []slack.Attachment{
					{
						Fallback:   "Testing: demo/test/1 -- https://ci.example.com/teams/main/pipelines/demo/jobs/test/builds/1",
						Color:      "#ffffff",
						AuthorName: "Testing",
						Fields: []slack.Field{
							{Title: "Job", Value: "demo/test", Short: true},
							{Title: "Build", Value: "1", Short: true},
						},
						Footer: "https://ci.example.com/teams/main/pipelines/demo/jobs/test/builds/1", FooterIcon: ""},
				},
				Channel: ""},
		},
		"channel and url set": {
			alert: Alert{
				Type:    "default",
				Channel: "general",
				Color:   "#ffffff",
				IconURL: "",
				Message: "Testing",
			},
			want: &slack.Message{
				Attachments: []slack.Attachment{
					{
						Fallback:   "Testing: demo/test/1 -- https://ci.example.com/teams/main/pipelines/demo/jobs/test/builds/1",
						Color:      "#ffffff",
						AuthorName: "Testing",
						Fields: []slack.Field{
							{Title: "Job", Value: "demo/test", Short: true},
							{Title: "Build", Value: "1", Short: true},
						},
						Footer: "https://ci.example.com/teams/main/pipelines/demo/jobs/test/builds/1", FooterIcon: ""},
				},
				Channel: "general"},
		},
		"message file": {
			alert: Alert{
				Type:        "default",
				Message:     "Testing",
				MessageFile: "message_file",
			},
			want: &slack.Message{
				Attachments: []slack.Attachment{
					{
						Fallback:   "message file: demo/test/1 -- https://ci.example.com/teams/main/pipelines/demo/jobs/test/builds/1",
						AuthorName: "message file",
						Fields: []slack.Field{
							{Title: "Job", Value: "demo/test", Short: true},
							{Title: "Build", Value: "1", Short: true},
						},
						Footer: "https://ci.example.com/teams/main/pipelines/demo/jobs/test/builds/1", FooterIcon: ""},
				},
			},
		},
		"message file failure": {
			alert: Alert{
				Type:        "default",
				Message:     "Testing",
				MessageFile: "bad file",
			},
			want: &slack.Message{
				Attachments: []slack.Attachment{
					{
						Fallback:   "Testing: demo/test/1 -- https://ci.example.com/teams/main/pipelines/demo/jobs/test/builds/1",
						AuthorName: "Testing",
						Fields: []slack.Field{
							{Title: "Job", Value: "demo/test", Short: true},
							{Title: "Build", Value: "1", Short: true},
						},
						Footer: "https://ci.example.com/teams/main/pipelines/demo/jobs/test/builds/1", FooterIcon: ""},
				},
			},
		},
	}

	metadata := concourse.BuildMetadata{
		Host:         "https://ci.example.com",
		TeamName:     "main",
		PipelineName: "demo",
		JobName:      "test",
		BuildName:    "1",
		URL:          "https://ci.example.com/teams/main/pipelines/demo/jobs/test/builds/1",
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			path := ""
			if c.alert.MessageFile != "" {
				dir, err := ioutil.TempDir("", "example")
				if err != nil {
					t.Fatal(err)
				}
				path = dir

				defer os.RemoveAll(dir)
				if err := ioutil.WriteFile(filepath.Join(dir, "message_file"), []byte("message file"), 0666); err != nil {
					t.Fatal(err)
				}
			}

			got := buildMessage(c.alert, metadata, path)
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("unexpected slack.Message value from buildSlackMessage:\n\t(GOT): %#v\n\t(WNT): %#v", got, c.want)
			}
		})
	}
}
