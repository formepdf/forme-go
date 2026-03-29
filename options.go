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

// JobResult is returned from GetJob.
type JobResult struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	PDFBase64 string `json:"pdfBase64,omitempty"`
	URL       string `json:"url,omitempty"`
	Error     string `json:"error,omitempty"`
}
