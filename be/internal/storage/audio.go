package storage

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

// AudioStorageService defines the interface for object storage operations.
type AudioStorageService interface {
	Upload(ctx context.Context, key string, data []byte, contentType string) (string, error)
	GenerateSignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	Delete(ctx context.Context, key string) error
}

// s3Storage is a minimal AWS S3-compatible client using raw HTTP + SigV4.
type s3Storage struct {
	endpoint  string
	bucket    string
	region    string
	accessKey string
	secretKey string
	client    *http.Client
}

// NewS3Storage creates an S3-compatible storage client without external SDK deps.
func NewS3Storage() (AudioStorageService, error) {
	endpoint := os.Getenv("S3_ENDPOINT")
	bucket := os.Getenv("S3_BUCKET")
	accessKey := os.Getenv("S3_ACCESS_KEY")
	secretKey := os.Getenv("S3_SECRET_KEY")
	region := os.Getenv("S3_REGION")
	if region == "" {
		region = "us-east-1"
	}
	if endpoint == "" || bucket == "" || accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("S3 env vars not configured")
	}
	return &s3Storage{
		endpoint:  strings.TrimRight(endpoint, "/"),
		bucket:    bucket,
		region:    region,
		accessKey: accessKey,
		secretKey: secretKey,
		client:    &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// Upload stores data using a PUT request with AWS SigV4 signature.
func (s *s3Storage) Upload(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	objectURL := fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucket, key)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, objectURL, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("build upload request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	req.ContentLength = int64(len(data))

	s.signRequest(req, data)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated &&
		resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed (status %d): %s", resp.StatusCode, string(body))
	}

	return objectURL, nil
}

// GenerateSignedURL creates a pre-signed URL (query-string auth style).
func (s *s3Storage) GenerateSignedURL(_ context.Context, key string, expiry time.Duration) (string, error) {
	now := time.Now().UTC()
	dateStr := now.Format("20060102")
	datetimeStr := now.Format("20060102T150405Z")
	expirySeconds := int(expiry.Seconds())
	if expirySeconds <= 0 {
		expirySeconds = 3600
	}

	credential := fmt.Sprintf("%s/%s/%s/s3/aws4_request", s.accessKey, dateStr, s.region)
	objectURL := fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucket, key)

	params := url.Values{}
	params.Set("X-Amz-Algorithm", "AWS4-HMAC-SHA256")
	params.Set("X-Amz-Credential", credential)
	params.Set("X-Amz-Date", datetimeStr)
	params.Set("X-Amz-Expires", fmt.Sprintf("%d", expirySeconds))
	params.Set("X-Amz-SignedHeaders", "host")

	parsedURL, _ := url.Parse(objectURL)
	canonicalRequest := strings.Join([]string{
		"GET",
		"/" + s.bucket + "/" + key,
		params.Encode(),
		"host:" + parsedURL.Host + "\n",
		"host",
		"UNSIGNED-PAYLOAD",
	}, "\n")

	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		datetimeStr,
		fmt.Sprintf("%s/%s/s3/aws4_request", dateStr, s.region),
		hashSHA256(canonicalRequest),
	}, "\n")

	signingKey := s.deriveSigningKey(dateStr)
	signature := hex.EncodeToString(hmacSHA256(signingKey, stringToSign))

	params.Set("X-Amz-Signature", signature)
	return objectURL + "?" + params.Encode(), nil
}

// Delete removes an object from S3.
func (s *s3Storage) Delete(ctx context.Context, key string) error {
	objectURL := fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucket, key)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, objectURL, nil)
	if err != nil {
		return fmt.Errorf("build delete request: %w", err)
	}
	s.signRequest(req, nil)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("delete request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed (status %d): %s", resp.StatusCode, string(body))
	}
	return nil
}

// signRequest adds AWS SigV4 Authorization header.
func (s *s3Storage) signRequest(req *http.Request, body []byte) {
	now := time.Now().UTC()
	dateStr := now.Format("20060102")
	datetimeStr := now.Format("20060102T150405Z")

	req.Header.Set("x-amz-date", datetimeStr)
	req.Header.Set("host", req.URL.Host)

	bodyHash := hashSHA256(string(body))
	req.Header.Set("x-amz-content-sha256", bodyHash)

	// Canonical headers (sorted)
	headerKeys := []string{"host", "x-amz-content-sha256", "x-amz-date"}
	if ct := req.Header.Get("Content-Type"); ct != "" {
		req.Header.Set("content-type", ct)
		headerKeys = append(headerKeys, "content-type")
	}
	sort.Strings(headerKeys)

	var canonicalHeaders strings.Builder
	for _, k := range headerKeys {
		canonicalHeaders.WriteString(k + ":" + strings.TrimSpace(req.Header.Get(k)) + "\n")
	}
	signedHeaders := strings.Join(headerKeys, ";")

	canonicalRequest := strings.Join([]string{
		req.Method,
		req.URL.Path,
		req.URL.RawQuery,
		canonicalHeaders.String(),
		signedHeaders,
		bodyHash,
	}, "\n")

	credentialScope := fmt.Sprintf("%s/%s/s3/aws4_request", dateStr, s.region)
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		datetimeStr,
		credentialScope,
		hashSHA256(canonicalRequest),
	}, "\n")

	signingKey := s.deriveSigningKey(dateStr)
	signature := hex.EncodeToString(hmacSHA256(signingKey, stringToSign))

	req.Header.Set("Authorization", fmt.Sprintf(
		"AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		s.accessKey, credentialScope, signedHeaders, signature,
	))
}

func (s *s3Storage) deriveSigningKey(dateStr string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+s.secretKey), dateStr)
	kRegion := hmacSHA256(kDate, s.region)
	kService := hmacSHA256(kRegion, "s3")
	return hmacSHA256(kService, "aws4_request")
}

func hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}

func hashSHA256(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// ── Key helpers ───────────────────────────────────────────────────────────────

// NewObjectKey generates a UUID-based object key for audio files.
func NewObjectKey(ext string) string {
	return fmt.Sprintf("audio/%s.%s", uuid.New().String(), ext)
}

// NewImageKey generates a UUID-based object key for image files.
func NewImageKey() string {
	return fmt.Sprintf("images/%s.png", uuid.New().String())
}

// ExtractKeyFromURL extracts the S3 object key from a full URL.
func ExtractKeyFromURL(rawURL, endpoint, bucket string) string {
	prefix := fmt.Sprintf("%s/%s/", strings.TrimRight(endpoint, "/"), bucket)
	if strings.HasPrefix(rawURL, prefix) {
		return rawURL[len(prefix):]
	}
	return rawURL
}

// ── Audio helpers ─────────────────────────────────────────────────────────────

// FetchAudioDuration estimates audio duration from raw bytes.
func FetchAudioDuration(data []byte, contentType string) (float64, error) {
	if contentType == "audio/wav" || contentType == "audio/x-wav" {
		return estimateWAVDuration(data)
	}
	sizeMB := float64(len(data)) / (1024 * 1024)
	return sizeMB * 60.0, nil
}

func estimateWAVDuration(data []byte) (float64, error) {
	if len(data) < 44 {
		return 0, fmt.Errorf("data too short for WAV")
	}
	byteRate := int(data[28]) | int(data[29])<<8 | int(data[30])<<16 | int(data[31])<<24
	if byteRate == 0 {
		return 0, fmt.Errorf("invalid WAV byte rate")
	}
	return float64(len(data)-44) / float64(byteRate), nil
}

// ── No-op storage for dev ─────────────────────────────────────────────────────

// NoOpStorage is a dev-mode storage that logs but doesn't actually store files.
type NoOpStorage struct{}

func (n *NoOpStorage) Upload(_ context.Context, key string, _ []byte, _ string) (string, error) {
	return fmt.Sprintf("http://localhost:9000/bahasa-daerah/%s", key), nil
}

func (n *NoOpStorage) GenerateSignedURL(_ context.Context, key string, _ time.Duration) (string, error) {
	return fmt.Sprintf("http://localhost:9000/bahasa-daerah/%s?signed=mock", key), nil
}

func (n *NoOpStorage) Delete(_ context.Context, _ string) error {
	return nil
}