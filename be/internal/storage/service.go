package storage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	maxAudioSizeBytes    = 5 * 1024 * 1024 // 5MB
	maxAudioDurationSecs = 30.0
	signedURLExpiry      = time.Hour
)

var (
	ErrInvalidFormat   = errors.New("invalid audio format: only WAV and MP3 are supported")
	ErrFileTooLarge    = errors.New("audio file exceeds 5MB limit")
	ErrDurationTooLong = errors.New("audio duration exceeds 30 seconds")
)

// Service handles audio and image upload business logic.
type Service struct {
	repo    Repository
	storage AudioStorageService
}

// NewService creates a new storage service.
func NewService(repo Repository, storage AudioStorageService) *Service {
	return &Service{repo: repo, storage: storage}
}

// UploadAudio validates and uploads audio, then updates the phrase record.
func (s *Service) UploadAudio(ctx context.Context, phraseID uuid.UUID, data []byte, filename string) (url string, duration float64, err error) {
	ext, contentType, err := detectAudioFormat(filename, data)
	if err != nil {
		return "", 0, err
	}
	if len(data) > maxAudioSizeBytes {
		return "", 0, ErrFileTooLarge
	}
	duration, _ = FetchAudioDuration(data, contentType) // non-fatal
	if duration > maxAudioDurationSecs {
		return "", 0, ErrDurationTooLong
	}

	key := NewObjectKey(ext)
	url, err = s.storage.Upload(ctx, key, data, contentType)
	if err != nil {
		return "", 0, fmt.Errorf("upload audio: %w", err)
	}
	if err = s.repo.UpdatePhraseAudioURL(ctx, phraseID, url, duration); err != nil {
		return "", 0, fmt.Errorf("save audio url: %w", err)
	}
	return url, duration, nil
}

// UploadImage validates and uploads a PNG, then updates the phrase record.
func (s *Service) UploadImage(ctx context.Context, phraseID uuid.UUID, data []byte) (string, error) {
	key := NewImageKey()
	url, err := s.storage.Upload(ctx, key, data, "image/png")
	if err != nil {
		return "", fmt.Errorf("upload image: %w", err)
	}
	if err = s.repo.UpdatePhraseImageURL(ctx, phraseID, url); err != nil {
		return "", fmt.Errorf("save image url: %w", err)
	}
	return url, nil
}

// DeleteAudio removes the audio file from storage for a given phrase.
func (s *Service) DeleteAudio(ctx context.Context, phraseID uuid.UUID) error {
	url, err := s.repo.GetPhraseAudioURL(ctx, phraseID)
	if err != nil || url == "" {
		return nil
	}
	key := ExtractKeyFromURL(url, os.Getenv("S3_ENDPOINT"), os.Getenv("S3_BUCKET"))
	return s.storage.Delete(ctx, key)
}

// DeleteImage removes the image file from storage for a given phrase.
func (s *Service) DeleteImage(ctx context.Context, phraseID uuid.UUID) error {
	url, err := s.repo.GetPhraseImageURL(ctx, phraseID)
	if err != nil || url == "" {
		return nil
	}
	key := ExtractKeyFromURL(url, os.Getenv("S3_ENDPOINT"), os.Getenv("S3_BUCKET"))
	return s.storage.Delete(ctx, key)
}

// detectAudioFormat inspects extension and magic bytes to determine format.
func detectAudioFormat(filename string, data []byte) (ext, contentType string, err error) {
	lower := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lower, ".mp3"):
		return "mp3", "audio/mpeg", nil
	case strings.HasSuffix(lower, ".wav"):
		return "wav", "audio/wav", nil
	}
	// Fall back to magic bytes
	if len(data) >= 4 {
		if data[0] == 'R' && data[1] == 'I' && data[2] == 'F' && data[3] == 'F' {
			return "wav", "audio/wav", nil
		}
		if data[0] == 0xFF && (data[1]&0xE0 == 0xE0) {
			return "mp3", "audio/mpeg", nil
		}
		if data[0] == 'I' && data[1] == 'D' && data[2] == '3' {
			return "mp3", "audio/mpeg", nil
		}
	}
	return "", "", ErrInvalidFormat
}