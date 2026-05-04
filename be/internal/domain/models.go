package domain

import (
	"time"

	"github.com/google/uuid"
)

// ── Users ────────────────────────────────────────────────────────────────────

type Role string

const (
	RoleLearner     Role = "learner"
	RoleContributor Role = "contributor"
	RoleAdmin       Role = "admin"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ── Languages ────────────────────────────────────────────────────────────────

type Language struct {
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Region    string    `json:"region,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// ── Cultural Context ─────────────────────────────────────────────────────────

type CulturalContext struct {
	ID             uuid.UUID `json:"id"`
	LanguageCode   string    `json:"language_code"`
	Region         string    `json:"region,omitempty"`
	UsageSituation string    `json:"usage_situation,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// ── Phrases ──────────────────────────────────────────────────────────────────

type PhraseStatus string

const (
	StatusPending  PhraseStatus = "pending"
	StatusApproved PhraseStatus = "approved"
	StatusRejected PhraseStatus = "rejected"
	StatusFlagged  PhraseStatus = "flagged"
	StatusAIFailed PhraseStatus = "ai_failed"
)

type ScriptStatus string

const (
	ScriptStatusNone     ScriptStatus = "none"
	ScriptStatusPending  ScriptStatus = "pending"
	ScriptStatusApproved ScriptStatus = "approved"
	ScriptStatusRejected ScriptStatus = "rejected"
)

type AudioStatus string

const (
	AudioStatusNone     AudioStatus = "none"
	AudioStatusPending  AudioStatus = "pending"
	AudioStatusApproved AudioStatus = "audio_approved"
	AudioStatusRejected AudioStatus = "audio_rejected"
)

type ScriptType string

const (
	ScriptLatin     ScriptType = "latin"
	ScriptJavanese  ScriptType = "javanese"
	ScriptSundanese ScriptType = "sundanese"
	ScriptBalinese  ScriptType = "balinese"
	ScriptLontara   ScriptType = "lontara"
	ScriptBatak     ScriptType = "batak"
	ScriptOther     ScriptType = "other"
)

type Tone string

const (
	ToneFormal Tone = "formal"
	ToneNetral Tone = "netral"
	ToneKasar  Tone = "kasar"
)

type Phrase struct {
	ID                   uuid.UUID    `json:"id"`
	TextLatin            string       `json:"text_latin"`
	TextNativeScript     *string      `json:"text_native_script,omitempty"`
	ScriptType           *ScriptType  `json:"script_type,omitempty"`
	Translation          string       `json:"translation"`
	LanguageCode         string       `json:"language_code"`
	Tone                 *Tone        `json:"tone,omitempty"`
	Status               PhraseStatus `json:"status"`
	ScriptStatus         ScriptStatus `json:"script_status"`
	ContributorID        uuid.UUID    `json:"contributor_id"`
	CulturalContextID    *uuid.UUID   `json:"cultural_context_id,omitempty"`
	AudioURL             *string      `json:"audio_url,omitempty"`
	AudioDurationSeconds *float64     `json:"audio_duration_seconds,omitempty"`
	AudioStatus          AudioStatus  `json:"audio_status"`
	NativeScriptImageURL *string      `json:"native_script_image_url,omitempty"`
	ModeratedBy          *uuid.UUID   `json:"moderated_by,omitempty"`
	ModeratedAt          *time.Time   `json:"moderated_at,omitempty"`
	UpvoteCount          int          `json:"upvote_count"`
	DownvoteCount        int          `json:"downvote_count"`
	FlagCount            int          `json:"flag_count"`
	AudioUpvoteCount     int          `json:"audio_upvote_count"`
	AudioDownvoteCount   int          `json:"audio_downvote_count"`
	ScriptUpvoteCount    int          `json:"script_upvote_count"`
	ScriptDownvoteCount  int          `json:"script_downvote_count"`
	CreatedAt            time.Time    `json:"created_at"`
	UpdatedAt            time.Time    `json:"updated_at"`

	// Joined fields (populated on read)
	Words           []Word           `json:"words,omitempty"`
	CulturalContext *CulturalContext `json:"cultural_context,omitempty"`
}

// ── Words ────────────────────────────────────────────────────────────────────

type Word struct {
	ID                      uuid.UUID   `json:"id"`
	SurfaceFormLatin        string      `json:"surface_form_latin"`
	SurfaceFormNativeScript *string     `json:"surface_form_native_script,omitempty"`
	RootFormLatin           string      `json:"root_form_latin"`
	RootFormNativeScript    *string     `json:"root_form_native_script,omitempty"`
	ScriptType              *ScriptType `json:"script_type,omitempty"`
	PartOfSpeech            *string     `json:"part_of_speech,omitempty"`
	LanguageCode            string      `json:"language_code"`
	Position                int         `json:"position,omitempty"`
	CreatedAt               time.Time   `json:"created_at"`
}

// ── Votes & Flags ────────────────────────────────────────────────────────────

type VoteType string

const (
	VoteUpvote   VoteType = "upvote"
	VoteDownvote VoteType = "downvote"
)

type Vote struct {
	ID            uuid.UUID `json:"id"`
	PhraseID      uuid.UUID `json:"phrase_id"`
	ContributorID uuid.UUID `json:"contributor_id"`
	VoteType      VoteType  `json:"vote_type"`
	CreatedAt     time.Time `json:"created_at"`
}

type FlagReason string

const (
	FlagInaccurateTranslation FlagReason = "inaccurate_translation"
	FlagInappropriateContent  FlagReason = "inappropriate_content"
	FlagDuplicate             FlagReason = "duplicate"
)

type Flag struct {
	ID        uuid.UUID  `json:"id"`
	PhraseID  uuid.UUID  `json:"phrase_id"`
	UserID    uuid.UUID  `json:"user_id"`
	Reason    FlagReason `json:"reason"`
	CreatedAt time.Time  `json:"created_at"`
}

// ── Practice Results ─────────────────────────────────────────────────────────

type PracticeResult string

const (
	PracticeResultTahu      PracticeResult = "tahu"
	PracticeResultBelumTahu PracticeResult = "belum_tahu"
)

type PhrasePracticeResult struct {
	ID        uuid.UUID      `json:"id"`
	LearnerID uuid.UUID      `json:"learner_id"`
	PhraseID  uuid.UUID      `json:"phrase_id"`
	Result    PracticeResult `json:"result"`
	CreatedAt time.Time      `json:"created_at"`
}
