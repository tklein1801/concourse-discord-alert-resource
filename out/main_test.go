package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tklein1801/concourse-discord-alert-resource/concourse"
	"github.com/tklein1801/concourse-discord-alert-resource/discord"
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
					{Name: "alerted", Value: "false"},
				},
			},
			env: env,
		},
		"error without Discord Webhook URL": {
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

			got, err := out(c.outRequest)
			if err != nil && !c.err {
				t.Fatalf("unexpected error from out:\n\t(ERR): %s", err)
			} else if err == nil && c.err {
				t.Fatalf("expected an error from out:\n\t(GOT): nil")
			} else if !cmp.Equal(got, c.want) {
				t.Fatalf("unexpected concourse.OutResponse value from out:\n\t(GOT): %#v\n\t(WNT): %#v\n\t(DIFF): %v", got, c.want, cmp.Diff(got, c.want))
			}
		})
	}
}
func TestBuildMessage(t *testing.T) {
	cases := map[string]struct {
		alert Alert
		want  *discord.Message
	}{
		"url set": {
			alert: Alert{
				Type:    "default",
				Color:   0xffffff,
				IconURL: "",
				Message: "Defaulted",
			},
			want: &discord.Message{
				Username: "Concourse CI",
				Embeds: []discord.Embed{
					{
						Title:       "Defaulted",
						Description: "The execution of task 'test' in pipeline 'demo' ended with status 'default'.",
						URL:         "https://ci.example.com/teams/main/pipelines/demo/jobs/test/builds/1",
						Color:       0xffffff,
						Image:       &discord.Image{URL: "https://ci.example.com/teams/main/pipelines/demo/jobs/test/builds/1"},
						Fields: []discord.Field{
							{
								Name:   "Step",
								Value:  "demo/test",
								Inline: true,
							},
							{
								Name:   "Build",
								Value:  "1",
								Inline: true,
							},
						},
					},
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

			got := buildMessage(c.alert, metadata)
			if !cmp.Equal(got, c.want) {
				t.Fatalf("unexpected discord.Message value from buildDiscordMessage:\n\t(GOT): %#v\n\t(WNT): %#v\n\t(DIFF): %v", got, c.want, cmp.Diff(got, c.want))
			}
		})
	}
}

func TestPreviousBuildName(t *testing.T) {
	cases := map[string]struct {
		build string
		want  string

		err bool
	}{
		"standard": {
			build: "6",
			want:  "6",
		},
		"rerun 1": {
			build: "6.1",
			want:  "6",
		},
		"rerun x": {
			build: "6.2",
			want:  "6.1",
		},
		"error 1": {
			build: "X",
			err:   true,
		},
		"error x": {
			build: "6.X",
			err:   true,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := previousBuildName(c.build)
			if err != nil && !c.err {
				t.Fatalf("unexpected error from previousBuildName:\n\t(ERR): %s", err)
			} else if err == nil && c.err {
				t.Fatalf("expected an error from previousBuildName:\n\t(GOT): nil")
			} else if err != nil && c.err {
				return
			}

			if err != nil {
				t.Fatalf("unexpected value from previousBuildName:\n\t(GOT): %#v\n\t(WNT): %#v", got, c.want)
			}
		})
	}
}
