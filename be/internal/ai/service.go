package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/yourusername/bahasa-daerah-platform/internal/domain"
)

// aiWord is the word structure returned by the AI API.
type aiWord struct {
	Surface string `json:"surface"`
	Root    string `json:"root"`
	POS     string `json:"pos"`
}

// aiResponse is the expected JSON response from the AI API.
type aiResponse struct {
	Words []aiWord `json:"words"`
	Tone  string   `json:"tone"`
}

// Service provides the AI processing logic.
type Service struct {
	repo       Repository
	apiURL     string
	apiKey     string
	httpClient *http.Client
}

// NewService creates a new AI service.
func NewService(repo Repository) *Service {
	return &Service{
		repo:   repo,
		apiURL: os.Getenv("AI_API_URL"),
		apiKey: os.Getenv("AI_API_KEY"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Process calls the AI API for a phrase and persists the result.
// On failure after retries, marks the phrase as ai_failed.
func (s *Service) Process(ctx context.Context, phraseID uuid.UUID, textLatin, translation string) {
	const maxRetries = 3

	var lastErr error
	delay := time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(delay)
			delay *= 2
		}

		result, err := s.callAPI(ctx, textLatin, translation)
		if err != nil {
			lastErr = err
			log.Printf("[AI] attempt %d failed for phrase %s: %v", attempt+1, phraseID, err)
			continue
		}

		// Parse tone
		tone := normalizeTone(result.Tone)

		// Map words
		words := make([]domain.Word, 0, len(result.Words))
		for _, w := range result.Words {
			words = append(words, domain.Word{
				SurfaceFormLatin: w.Surface,
				RootFormLatin:    w.Root,
				PartOfSpeech:     &w.POS,
			})
		}

		if err := s.repo.SaveAIResults(ctx, phraseID, tone, words); err != nil {
			lastErr = err
			log.Printf("[AI] save failed for phrase %s: %v", phraseID, err)
			continue
		}

		log.Printf("[AI] processed phrase %s: tone=%s words=%d", phraseID, tone, len(words))
		return
	}

	// All retries exhausted
	log.Printf("[AI] all retries failed for phrase %s: %v", phraseID, lastErr)
	if err := s.repo.UpdatePhraseAIFailed(context.Background(), phraseID); err != nil {
		log.Printf("[AI] failed to mark phrase %s as ai_failed: %v", phraseID, err)
	}
}

// callAPI sends a request to the AI API and decodes the response.
func (s *Service) callAPI(ctx context.Context, textLatin, translation string) (*aiResponse, error) {
	prompt := fmt.Sprintf(
		`Analyze this regional Indonesian language phrase and respond ONLY with valid JSON, no other text.
Phrase: "%s"
Translation (Indonesian): "%s"

Respond with this exact JSON schema:
{"words":[{"surface":"<surface form>","root":"<root form>","pos":"<part of speech>"}],"tone":"<formal|netral|kasar>"}`,
		textLatin, translation,
	)

	reqBody, err := json.Marshal(map[string]any{
		"model": "gpt-4o-mini",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"response_format": map[string]string{"type": "json_object"},
		"max_tokens":      500,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.apiURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI API returned status %d", resp.StatusCode)
	}

	// OpenAI response wrapper
	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, fmt.Errorf("decode openai response: %w", err)
	}
	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in AI response")
	}

	content := openAIResp.Choices[0].Message.Content

	var result aiResponse
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("decode AI JSON content: %w", err)
	}
	return &result, nil
}

func normalizeTone(t string) domain.Tone {
	switch t {
	case "formal":
		return domain.ToneFormal
	case "kasar":
		return domain.ToneKasar
	default:
		return domain.ToneNetral
	}
}