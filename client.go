package forme

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

// Render renders a template to PDF bytes (synchronous).
func (f *Forme) Render(slug string, data map[string]any) ([]byte, error) {
	return f.RenderWithOptions(slug, data, RenderOptions{})
}

// RenderWithOptions renders a template to PDF bytes with options.
func (f *Forme) RenderWithOptions(slug string, data map[string]any, opts RenderOptions) ([]byte, error) {
	body := make(map[string]any)
	for k, v := range data {
		body[k] = v
	}
	if opts.EmbedData {
		body["embedData"] = true
	}

	respBody, contentType, err := f.doJSON("POST", fmt.Sprintf("/v1/render/%s", slug), body)
	if err != nil {
		return nil, err
	}

	if strings.Contains(contentType, "application/json") {
		return respBody, nil
	}
	return respBody, nil
}

// RenderS3 renders a template and uploads the result to S3.
func (f *Forme) RenderS3(slug string, data map[string]any, opts S3Options) (*S3Result, error) {
	body := make(map[string]any)
	for k, v := range data {
		body[k] = v
	}
	body["s3"] = opts

	respBody, _, err := f.doJSON("POST", fmt.Sprintf("/v1/render/%s", slug), body)
	if err != nil {
		return nil, err
	}

	var result S3Result
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse S3 response: %w", err)
	}
	return &result, nil
}

// RenderAsync starts an asynchronous render job.
func (f *Forme) RenderAsync(slug string, data map[string]any, opts AsyncOptions) (*AsyncResult, error) {
	body := make(map[string]any)
	for k, v := range data {
		body[k] = v
	}
	if opts.WebhookURL != "" {
		body["webhookUrl"] = opts.WebhookURL
	}

	respBody, _, err := f.doJSON("POST", fmt.Sprintf("/v1/render/%s/async", slug), body)
	if err != nil {
		return nil, err
	}

	var result AsyncResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse async response: %w", err)
	}
	return &result, nil
}

// GetJob polls the status of an async render job.
func (f *Forme) GetJob(jobID string) (*JobResult, error) {
	respBody, _, err := f.doJSON("GET", fmt.Sprintf("/v1/jobs/%s", jobID), nil)
	if err != nil {
		return nil, err
	}

	var result JobResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse job response: %w", err)
	}
	return &result, nil
}

// Merge combines multiple PDFs into a single PDF via multipart upload.
func (f *Forme) Merge(pdfs [][]byte) ([]byte, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	for i, pdf := range pdfs {
		part, err := w.CreateFormFile("files", fmt.Sprintf("file%d.pdf", i))
		if err != nil {
			return nil, fmt.Errorf("failed to create form file: %w", err)
		}
		if _, err := part.Write(pdf); err != nil {
			return nil, fmt.Errorf("failed to write PDF data: %w", err)
		}
	}
	w.Close()

	req, err := http.NewRequest("POST", f.baseURL+"/v1/merge", &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+f.apiKey)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, parseError(resp.StatusCode, respBody)
	}

	return respBody, nil
}

// Extract reads embedded data from a PDF.
// Returns nil, nil if the PDF has no embedded data.
func (f *Forme) Extract(pdfBytes []byte) (map[string]any, error) {
	req, err := http.NewRequest("POST", f.baseURL+"/v1/extract", bytes.NewReader(pdfBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+f.apiKey)
	req.Header.Set("Content-Type", "application/pdf")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fErr := parseError(resp.StatusCode, respBody)
		if fErr.Status == 404 && strings.Contains(strings.ToLower(fErr.Message), "no embedded data") {
			return nil, nil
		}
		return nil, fErr
	}

	var envelope struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(respBody, &envelope); err != nil {
		return nil, fmt.Errorf("failed to parse extract response: %w", err)
	}
	return envelope.Data, nil
}

// doJSON sends a JSON request and returns the response body and content type.
func (f *Forme) doJSON(method, path string, body any) ([]byte, string, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, "", fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, f.baseURL+path, reqBody)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Authorization", "Bearer "+f.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", parseError(resp.StatusCode, respBody)
	}

	return respBody, resp.Header.Get("Content-Type"), nil
}

func parseError(status int, body []byte) *FormeError {
	message := fmt.Sprintf("Request failed with status %d", status)

	var errBody map[string]any
	if err := json.Unmarshal(body, &errBody); err == nil {
		if msg, ok := errBody["error"].(string); ok && msg != "" {
			message = msg
		} else if msg, ok := errBody["message"].(string); ok && msg != "" {
			message = msg
		}
	}

	return &FormeError{Status: status, Message: message}
}
