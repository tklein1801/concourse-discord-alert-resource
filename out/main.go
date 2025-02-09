package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tklein1801/concourse-discord-alert-resource/concourse"
	"github.com/tklein1801/concourse-discord-alert-resource/discord"
)

func buildMessage(alert Alert, m concourse.BuildMetadata) *discord.Message {
	message := alert.Message
	url := m.URL
	if !strings.HasPrefix(url, "http") {
		url = "http://missing.domain" + url // Ensure URL has http prefix
	}

	embeds := []discord.Embed{
		{
			Title:       message,
			Description: fmt.Sprintf("The execution of task '%s' in pipeline '%s' ended with status '%s'.", m.JobName, m.PipelineName, alert.Type),
			URL:         url,
			Color:       alert.Color,
			Image: &discord.Image{
				URL: url,
			},
			Fields: []discord.Field{
				{
					Name:   "Step",
					Value:  fmt.Sprintf("%s/%s", m.PipelineName, m.JobName),
					Inline: true,
				},
				{
					Name:   "Build",
					Value:  m.BuildName,
					Inline: true,
				},
			},
		},
	}

	return &discord.Message{Username: "Concourse CI", Embeds: embeds}
}

func previousBuildStatus(input *concourse.OutRequest, m concourse.BuildMetadata) (string, error) {
	// Exit early if first build
	if m.BuildName == "1" {
		return "", nil
	}

	c, err := concourse.NewClient(m.Host, m.TeamName, input.Source.Username, input.Source.Password)
	if err != nil {
		return "", fmt.Errorf("error connecting to Concourse: %w", err)
	}

	p, err := previousBuildName(m.BuildName)
	if err != nil {
		return "", fmt.Errorf("error parsing build name: %w", err)
	}

	instanceVars := ""
	instanceVarsIndex := strings.Index(m.URL, "?")
	if instanceVarsIndex > -1 {
		instanceVars = m.URL[instanceVarsIndex:]
	}

	previous, err := c.JobBuild(m.PipelineName, m.JobName, p, instanceVars)
	if err != nil {
		return "", fmt.Errorf("error requesting Concourse build status: %w", err)
	}

	return previous.Status, nil
}

func previousBuildName(s string) (string, error) {
	strs := strings.Split(s, ".")

	if len(strs) == 1 {
		i, err := strconv.Atoi(strs[0])
		if err != nil {
			return "", err
		}

		return strconv.Itoa(i - 1), nil
	}

	i, err := strconv.Atoi(strs[1])
	if err != nil {
		return "", err
	}

	s = fmt.Sprintf("%s.%s", strs[0], strconv.Itoa(i-1))
	return strings.Trim(s, ".0"), nil
}

var maxElapsedTime = 30 * time.Second

func out(input *concourse.OutRequest) (*concourse.OutResponse, error) {
	if input.Source.URL == "" {
		return nil, errors.New("discord webhook url cannot be blank")
	}

	alert := NewAlert(input)
	metadata := concourse.NewBuildMetadata(input.Source.ConcourseURL)
	if alert.Disabled {
		return buildOut(alert.Type, false), nil
	}

	if alert.Type == "fixed" || alert.Type == "broke" {
		pstatus, err := previousBuildStatus(input, metadata)
		if err != nil {
			return nil, fmt.Errorf("error getting last build status: %w", err)
		}

		if (alert.Type == "fixed" && pstatus == "succeeded") || (alert.Type == "broke" && pstatus != "succeeded") {
			return buildOut(alert.Type, false), nil
		}
	}

	message := buildMessage(alert, metadata)
	err := discord.Send(input.Source.URL, message, maxElapsedTime)
	if err != nil {
		return nil, fmt.Errorf("error sending discord message: %w", err)
	}
	return buildOut(alert.Type, true), nil
}

func buildOut(atype string, alerted bool) *concourse.OutResponse {
	return &concourse.OutResponse{
		Version: concourse.Version{"ver": "static"},
		Metadata: []concourse.Metadata{
			{Name: "type", Value: atype},
			{Name: "alerted", Value: strconv.FormatBool(alerted)},
		},
	}
}

func main() {
	var input *concourse.OutRequest
	err := json.NewDecoder(os.Stdin).Decode(&input)
	if err != nil {
		log.Fatalln(fmt.Errorf("error reading stdin: %w", err))
	}

	o, err := out(input)
	if err != nil {
		log.Fatalln(err)
	}

	err = json.NewEncoder(os.Stdout).Encode(o)
	if err != nil {
		log.Fatalln(fmt.Errorf("error writing stdout: %w", err))
	}
}
