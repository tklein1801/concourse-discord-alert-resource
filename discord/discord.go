package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// Message represents the payload for an Discord webhook message
type Message struct {
	Content   string       `json:"content,omitempty"`    // Main message
	Username  string       `json:"username,omitempty"`   // Displayname of the webhook
	AvatarURL string       `json:"avatar_url,omitempty"` // Customized avatar
	TTS       bool         `json:"tts,omitempty"`        // Activate Text-to-Speech
	Embeds    []Embed      `json:"embeds,omitempty"`     // Embeds
	Files     []Attachment `json:"-"`                    // Attachments
}

// Embed represents an embedded object (Rich Embed)
type Embed struct {
	Title       string     `json:"title,omitempty"`
	Description string     `json:"description,omitempty"`
	URL         string     `json:"url,omitempty"`
	Color       int        `json:"color,omitempty"`
	Timestamp   string     `json:"timestamp,omitempty"`
	Footer      *Footer    `json:"footer,omitempty"`
	Image       *Image     `json:"image,omitempty"`
	Thumbnail   *Thumbnail `json:"thumbnail,omitempty"`
	Author      *Author    `json:"author,omitempty"`
	Fields      []Field    `json:"fields,omitempty"`
}

type Field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type Footer struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url,omitempty"`
}

type Image struct {
	URL string `json:"url"`
}

type Thumbnail struct {
	URL string `json:"url"`
}

type Author struct {
	Name    string `json:"name"`
	URL     string `json:"url,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}

type Attachment struct {
	Filename string
	Data     []byte
}

func (d *Message) ToJSON() ([]byte, error) {
	return json.Marshal(d)
}

// Send sends the message to the webhook URL.
func Send(url string, m *Message, maxRetryTime time.Duration) error {
	buf, err := json.Marshal(m)
	if err != nil {
		return err
	}

	err = backoff.Retry(
		func() error {
			r, err := http.Post(url, "application/json", bytes.NewReader(buf))
			if err != nil {
				return err
			}
			defer r.Body.Close()

			if r.StatusCode > 399 {
				return fmt.Errorf("unexpected response status code: '%d'! Payload: %s", r.StatusCode, buf)
			}
			return nil
		},
		backoff.NewExponentialBackOff(backoff.WithMaxElapsedTime(maxRetryTime)),
	)

	if err != nil {
		return err
	}
	return nil
}
