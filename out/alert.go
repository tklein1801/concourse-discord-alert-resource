package main

import (
	"strconv"

	"github.com/tklein1801/concourse-discord-alert-resource/concourse"
)

// An Alert defines the notification that will be sent to Slack.
type Alert struct {
	Type        string
	Color       uint
	IconURL     string
	Message     string
	MessageFile string
	Text        string
	TextFile    string
	Disabled    bool
}

// NewAlert constructs and returns an Alert.
func NewAlert(input *concourse.OutRequest) Alert {
	var alert Alert
	defaultColor := uint(0x35495c) // Dark blue
	switch input.Params.AlertType {
	case "success":
		alert = Alert{
			Type:    "success",
			Color:   0x32cd32, // Green
			IconURL: "https://ci.concourse-ci.org/public/images/favicon-succeeded.png",
			Message: "Success",
		}
	case "failed":
		alert = Alert{
			Type:    "failed",
			Color:   0xd00000, // Red
			IconURL: "https://ci.concourse-ci.org/public/images/favicon-failed.png",
			Message: "Failed",
		}
	case "started":
		alert = Alert{
			Type:    "started",
			Color:   0xf7cd42, // Yellow
			IconURL: "https://ci.concourse-ci.org/public/images/favicon-started.png",
			Message: "Started",
		}
	case "aborted":
		alert = Alert{
			Type:    "aborted",
			Color:   0x8d4b32, // Brown
			IconURL: "https://ci.concourse-ci.org/public/images/favicon-aborted.png",
			Message: "Aborted",
		}
	case "fixed":
		alert = Alert{
			Type:    "fixed",
			Color:   0x32cd32, // Green
			IconURL: "https://ci.concourse-ci.org/public/images/favicon-succeeded.png",
			Message: "Fixed",
		}
	case "broke":
		alert = Alert{
			Type:    "broke",
			Color:   0xd00000, // Red
			IconURL: "https://ci.concourse-ci.org/public/images/favicon-failed.png",
			Message: "Broke",
		}
	case "errored":
		alert = Alert{
			Type:    "errored",
			Color:   0xf5a623, // Orange
			IconURL: "https://ci.concourse-ci.org/public/images/favicon-errored.png",
			Message: "Errored",
		}
	default:
		alert = Alert{
			Type:    "default",
			Color:   defaultColor, // Dark blue
			IconURL: "https://ci.concourse-ci.org/public/images/favicon-pending.png",
			Message: "",
		}
	}

	alert.Disabled = input.Params.Disable
	if !alert.Disabled {
		alert.Disabled = input.Source.Disable
	}

	if input.Params.Message != "" {
		alert.Message = input.Params.Message
	}
	alert.MessageFile = input.Params.MessageFile

	if input.Params.Color != "" {
		colorValue, err := strconv.ParseUint(input.Params.Color, 16, 32)
		if err == nil {
			alert.Color = uint(colorValue)
		} else {
			alert.Color = defaultColor // Default color setting
		}
	} else {
		alert.Color = defaultColor // Default color setting
	}

	alert.Text = input.Params.Text
	alert.TextFile = input.Params.TextFile
	return alert
}
