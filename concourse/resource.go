package concourse

// A Source is the resource's source configuration.
type Source struct {
	URL          string `json:"url"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	ConcourseURL string `json:"concourse_url"`
	Disable      bool   `json:"disable"`
}

// Metadata are a key-value pair that must be included for in the in and out
// operation responses.
type Metadata struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

// Version is the key-value pair that the resource is checking, getting or putting.
type Version map[string]string

// CheckResponse is the output for the check operation.
type CheckResponse []Version

// InResponse is the output for the in operation.
type InResponse struct {
	Version  Version    `json:"version"`
	Metadata []Metadata `json:"metadata"`
}

// OutParams are the parameters that can be configured for the out operation.
type OutParams struct {
	AlertType   string `json:"alert_type"`
	Color       string `json:"color"`
	Message     string `json:"message"`
	MessageFile string `json:"message_file"`
	Text        string `json:"text"`
	TextFile    string `json:"text_file"`
	Disable     bool   `json:"disable"`
}

// OutRequest is in the input for the out operation.
type OutRequest struct {
	Source Source    `json:"source"`
	Params OutParams `json:"params"`
}

// OutResponse is the output for the out operation.
type OutResponse struct {
	Version  Version    `json:"version"`
	Metadata []Metadata `json:"metadata"`
}
