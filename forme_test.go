package forme

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRenderReturnsPDFBytes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.Write([]byte("%PDF-1.7 fake pdf content"))
	}))
	defer server.Close()

	client := New("forme_sk_test", WithBaseURL(server.URL))
	result, err := client.Render("invoice", map[string]any{"customer": "Acme"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != "%PDF-1.7 fake pdf content" {
		t.Fatalf("unexpected result: %s", result)
	}
}

func TestRenderSendsCorrectAuthHeader(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/pdf")
		w.Write([]byte("%PDF"))
	}))
	defer server.Close()

	client := New("forme_sk_test", WithBaseURL(server.URL))
	client.Render("invoice", nil)

	if gotAuth != "Bearer forme_sk_test" {
		t.Fatalf("expected 'Bearer forme_sk_test', got %q", gotAuth)
	}
}

func TestRenderSendsCorrectContentTypeAndBody(t *testing.T) {
	var gotContentType string
	var gotBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/pdf")
		w.Write([]byte("%PDF"))
	}))
	defer server.Close()

	client := New("forme_sk_test", WithBaseURL(server.URL))
	client.Render("invoice", map[string]any{"customer": "Acme"})

	if gotContentType != "application/json" {
		t.Fatalf("expected 'application/json', got %q", gotContentType)
	}

	var body map[string]any
	json.Unmarshal(gotBody, &body)
	if body["customer"] != "Acme" {
		t.Fatalf("expected customer=Acme, got %v", body["customer"])
	}
}

func TestRenderSendsCorrectPath(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/pdf")
		w.Write([]byte("%PDF"))
	}))
	defer server.Close()

	client := New("forme_sk_test", WithBaseURL(server.URL))
	client.Render("my-template", nil)

	if gotPath != "/v1/render/my-template" {
		t.Fatalf("expected '/v1/render/my-template', got %q", gotPath)
	}
}

func TestRenderReturnsFormeErrorOnNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(map[string]string{"error": "Template not found"})
	}))
	defer server.Close()

	client := New("forme_sk_test", WithBaseURL(server.URL))
	_, err := client.Render("missing", nil)
	if err == nil {
		t.Fatal("expected error")
	}

	fErr, ok := err.(*FormeError)
	if !ok {
		t.Fatalf("expected *FormeError, got %T", err)
	}
	if fErr.Status != 404 {
		t.Fatalf("expected status 404, got %d", fErr.Status)
	}
	if fErr.Message != "Template not found" {
		t.Fatalf("expected 'Template not found', got %q", fErr.Message)
	}
}

func TestRenderErrorFallbackMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(502)
		w.Write([]byte("Bad Gateway"))
	}))
	defer server.Close()

	client := New("forme_sk_test", WithBaseURL(server.URL))
	_, err := client.Render("invoice", nil)

	fErr, ok := err.(*FormeError)
	if !ok {
		t.Fatalf("expected *FormeError, got %T", err)
	}
	if fErr.Status != 502 {
		t.Fatalf("expected status 502, got %d", fErr.Status)
	}
	if fErr.Message != "Request failed with status 502" {
		t.Fatalf("expected fallback message, got %q", fErr.Message)
	}
}

func TestRenderErrorWithMessageField(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]string{"message": "Internal server error"})
	}))
	defer server.Close()

	client := New("forme_sk_test", WithBaseURL(server.URL))
	_, err := client.Render("invoice", nil)

	fErr := err.(*FormeError)
	if fErr.Message != "Internal server error" {
		t.Fatalf("expected 'Internal server error', got %q", fErr.Message)
	}
}

func TestRenderAsyncReturnsJobIDAndStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"jobId": "job-123", "status": "pending"})
	}))
	defer server.Close()

	client := New("forme_sk_test", WithBaseURL(server.URL))
	result, err := client.RenderAsync("invoice", map[string]any{"customer": "Acme"}, AsyncOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.JobID != "job-123" {
		t.Fatalf("expected jobId 'job-123', got %q", result.JobID)
	}
	if result.Status != "pending" {
		t.Fatalf("expected status 'pending', got %q", result.Status)
	}
}

func TestRenderAsyncSendsWebhookURL(t *testing.T) {
	var gotBody []byte
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"jobId": "job-456", "status": "pending"})
	}))
	defer server.Close()

	client := New("forme_sk_test", WithBaseURL(server.URL))
	client.RenderAsync("invoice", map[string]any{"x": 1}, AsyncOptions{WebhookURL: "https://hook.example.com"})

	if gotPath != "/v1/render/invoice/async" {
		t.Fatalf("expected '/v1/render/invoice/async', got %q", gotPath)
	}

	var body map[string]any
	json.Unmarshal(gotBody, &body)
	if body["webhookUrl"] != "https://hook.example.com" {
		t.Fatalf("expected webhookUrl, got %v", body["webhookUrl"])
	}
}

func TestRenderAsyncReturnsFormeErrorOnError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]string{"error": "Render failed"})
	}))
	defer server.Close()

	client := New("forme_sk_test", WithBaseURL(server.URL))
	_, err := client.RenderAsync("invoice", nil, AsyncOptions{})

	fErr, ok := err.(*FormeError)
	if !ok {
		t.Fatalf("expected *FormeError, got %T", err)
	}
	if fErr.Status != 500 {
		t.Fatalf("expected status 500, got %d", fErr.Status)
	}
}

func TestGetJobReturnsResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"id":        "job-123",
			"status":    "complete",
			"pdfBase64": "JVBER...",
		})
	}))
	defer server.Close()

	client := New("forme_sk_test", WithBaseURL(server.URL))
	result, err := client.GetJob("job-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "job-123" {
		t.Fatalf("expected id 'job-123', got %q", result.ID)
	}
	if result.Status != "complete" {
		t.Fatalf("expected status 'complete', got %q", result.Status)
	}
	if result.PDFBase64 != "JVBER..." {
		t.Fatalf("expected pdfBase64, got %q", result.PDFBase64)
	}
}

func TestGetJobSendsAuthAndGET(t *testing.T) {
	var gotAuth, gotMethod, gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"id": "job-1", "status": "pending"})
	}))
	defer server.Close()

	client := New("forme_sk_secret", WithBaseURL(server.URL))
	client.GetJob("job-1")

	if gotAuth != "Bearer forme_sk_secret" {
		t.Fatalf("expected 'Bearer forme_sk_secret', got %q", gotAuth)
	}
	if gotMethod != "GET" {
		t.Fatalf("expected GET, got %q", gotMethod)
	}
	if gotPath != "/v1/jobs/job-1" {
		t.Fatalf("expected '/v1/jobs/job-1', got %q", gotPath)
	}
}

func TestGetJobReturnsFormeErrorOn404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(map[string]string{"error": "Job not found"})
	}))
	defer server.Close()

	client := New("forme_sk_test", WithBaseURL(server.URL))
	_, err := client.GetJob("nonexistent")

	fErr, ok := err.(*FormeError)
	if !ok {
		t.Fatalf("expected *FormeError, got %T", err)
	}
	if fErr.Status != 404 {
		t.Fatalf("expected status 404, got %d", fErr.Status)
	}
}

func TestMergeSendsMultipart(t *testing.T) {
	var gotContentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		// Verify multipart
		r.ParseMultipartForm(10 << 20)
		files := r.MultipartForm.File["files"]
		if len(files) != 2 {
			w.WriteHeader(400)
			return
		}
		w.Header().Set("Content-Type", "application/pdf")
		w.Write([]byte("%PDF-merged"))
	}))
	defer server.Close()

	client := New("forme_sk_test", WithBaseURL(server.URL))
	result, err := client.Merge([][]byte{[]byte("%PDF-1"), []byte("%PDF-2")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != "%PDF-merged" {
		t.Fatalf("unexpected result: %s", result)
	}
	if !strings.Contains(gotContentType, "multipart/form-data") {
		t.Fatalf("expected multipart content type, got %q", gotContentType)
	}
}

func TestMergeReturnsFormeErrorOnError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]string{"error": "At least 2 PDFs required"})
	}))
	defer server.Close()

	client := New("forme_sk_test", WithBaseURL(server.URL))
	_, err := client.Merge([][]byte{[]byte("%PDF-1")})

	fErr, ok := err.(*FormeError)
	if !ok {
		t.Fatalf("expected *FormeError, got %T", err)
	}
	if fErr.Status != 400 {
		t.Fatalf("expected status 400, got %d", fErr.Status)
	}
}

func TestExtractReturnsData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"customer": "Acme", "total": float64(100)},
		})
	}))
	defer server.Close()

	client := New("forme_sk_test", WithBaseURL(server.URL))
	result, err := client.Extract([]byte("%PDF-fake"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["customer"] != "Acme" {
		t.Fatalf("expected customer=Acme, got %v", result["customer"])
	}
}

func TestExtractReturnsNilOnNoEmbeddedData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(map[string]string{"error": "No embedded data found"})
	}))
	defer server.Close()

	client := New("forme_sk_test", WithBaseURL(server.URL))
	result, err := client.Extract([]byte("%PDF-fake"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil, got %v", result)
	}
}

func TestExtractReturnsErrorOnOtherErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]string{"error": "Server error"})
	}))
	defer server.Close()

	client := New("forme_sk_test", WithBaseURL(server.URL))
	_, err := client.Extract([]byte("%PDF-fake"))

	fErr, ok := err.(*FormeError)
	if !ok {
		t.Fatalf("expected *FormeError, got %T", err)
	}
	if fErr.Status != 500 {
		t.Fatalf("expected status 500, got %d", fErr.Status)
	}
}

func TestExtractSendsPDFContentType(t *testing.T) {
	var gotContentType string
	var gotBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{}})
	}))
	defer server.Close()

	client := New("forme_sk_test", WithBaseURL(server.URL))
	client.Extract([]byte("%PDF-fake"))

	if gotContentType != "application/pdf" {
		t.Fatalf("expected 'application/pdf', got %q", gotContentType)
	}
	if string(gotBody) != "%PDF-fake" {
		t.Fatalf("expected body '%s', got '%s'", "%PDF-fake", gotBody)
	}
}

func TestWithBaseURLIsRespected(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.Write([]byte("%PDF"))
	}))
	defer server.Close()

	client := New("sk", WithBaseURL(server.URL))
	_, err := client.Render("tpl", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWithBaseURLTrailingSlashStripped(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/pdf")
		w.Write([]byte("%PDF"))
	}))
	defer server.Close()

	client := New("sk", WithBaseURL(server.URL+"/"))
	client.Render("tpl", nil)

	if gotPath != "/v1/render/tpl" {
		t.Fatalf("expected '/v1/render/tpl', got %q", gotPath)
	}
}

func TestWithHTTPClientIsUsed(t *testing.T) {
	called := false
	customTransport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		called = true
		return &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"application/pdf"}},
			Body:       io.NopCloser(strings.NewReader("%PDF")),
		}, nil
	})
	customClient := &http.Client{Transport: customTransport}

	client := New("sk", WithHTTPClient(customClient))
	client.Render("tpl", nil)

	if !called {
		t.Fatal("custom HTTP client transport was not called")
	}
}

func TestRenderS3ReturnsURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var data map[string]any
		json.Unmarshal(body, &data)

		s3, ok := data["s3"].(map[string]any)
		if !ok {
			w.WriteHeader(400)
			fmt.Fprint(w, `{"error":"missing s3"}`)
			return
		}
		if s3["bucket"] != "my-bucket" {
			w.WriteHeader(400)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"url": "https://s3.example.com/invoice.pdf"})
	}))
	defer server.Close()

	client := New("forme_sk_test", WithBaseURL(server.URL))
	result, err := client.RenderS3("invoice", map[string]any{"customer": "Acme"}, S3Options{
		Bucket:         "my-bucket",
		Key:            "invoice.pdf",
		AccessKeyID:    "AK",
		SecretAccessKey: "SK",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.URL != "https://s3.example.com/invoice.pdf" {
		t.Fatalf("expected S3 URL, got %q", result.URL)
	}
}

// roundTripperFunc is a helper to create custom http.RoundTripper from a function.
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
