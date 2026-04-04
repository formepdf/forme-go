package forme

// RenderOptions configures a render request.
type RenderOptions struct {
	// EmbedData causes the data to be embedded in the PDF as a JSON attachment.
	EmbedData bool
}

// S3Options configures S3 upload for a render request.
type S3Options struct {
	Bucket          string `json:"bucket"`
	Key             string `json:"key"`
	AccessKeyID     string `json:"accessKeyId"`
	SecretAccessKey  string `json:"secretAccessKey"`
	Region          string `json:"region,omitempty"`
	SessionToken    string `json:"sessionToken,omitempty"`
}

// AsyncOptions configures an async render request.
type AsyncOptions struct {
	WebhookURL string
}

// S3Result is returned from RenderS3.
type S3Result struct {
	URL string `json:"url"`
}

// AsyncResult is returned from RenderAsync.
type AsyncResult struct {
	JobID  string `json:"jobId"`
	Status string `json:"status"`
}

// CertifyOptions configures a certify request.
type CertifyOptions struct {
	Reason   string `json:"reason,omitempty"`
	Location string `json:"location,omitempty"`
	Contact  string `json:"contact,omitempty"`
}

// RedactOptions configures a redact request.
type RedactOptions struct {
	Redactions []RedactionRegion  `json:"redactions,omitempty"`
	Patterns   []RedactionPattern `json:"patterns,omitempty"`
	Presets    []string           `json:"presets,omitempty"`
	Template   string             `json:"template,omitempty"`
}

// RedactionRegion specifies a coordinate-based redaction area.
type RedactionRegion struct {
	Page   int     `json:"page"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Color  string  `json:"color,omitempty"`
}

// RedactionPattern specifies a text search pattern for redaction.
type RedactionPattern struct {
	Pattern     string `json:"pattern"`
	PatternType string `json:"pattern_type"` // "Literal" or "Regex"
	Page        *int   `json:"page,omitempty"`
	Color       string `json:"color,omitempty"`
}

// RasterizeOptions configures a rasterize request.
type RasterizeOptions struct {
	DPI int `json:"dpi,omitempty"` // 72-300, default 150
}

// JobResult is returned from GetJob.
type JobResult struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	PDFBase64 string `json:"pdfBase64,omitempty"`
	URL       string `json:"url,omitempty"`
	Error     string `json:"error,omitempty"`
}
