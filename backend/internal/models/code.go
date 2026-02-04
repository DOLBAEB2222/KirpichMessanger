package models

import (
	"time"

	"github.com/google/uuid"
)

type CodeLanguage string

const (
	CodeLanguageJavaScript CodeLanguage = "javascript"
	CodeLanguageTypeScript CodeLanguage = "typescript"
	CodeLanguagePython     CodeLanguage = "python"
	CodeLanguageGo         CodeLanguage = "go"
	CodeLanguageJava       CodeLanguage = "java"
	CodeLanguageC          CodeLanguage = "c"
	CodeLanguageCPP        CodeLanguage = "cpp"
	CodeLanguageRust       CodeLanguage = "rust"
	CodeLanguagePHP        CodeLanguage = "php"
	CodeLanguageRuby       CodeLanguage = "ruby"
	CodeLanguageSQL        CodeLanguage = "sql"
	CodeLanguageHTML       CodeLanguage = "html"
	CodeLanguageCSS        CodeLanguage = "css"
	CodeLanguageBash       CodeLanguage = "bash"
	CodeLanguageJSON       CodeLanguage = "json"
	CodeLanguageXML        CodeLanguage = "xml"
	CodeLanguageMarkdown   CodeLanguage = "markdown"
	CodeLanguageOther      CodeLanguage = "other"
)

type CodeSnippet struct {
	ID          uuid.UUID     `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	MessageID   uuid.UUID     `gorm:"type:uuid;not null;unique" json:"message_id"`
	ChatID      uuid.UUID     `gorm:"type:uuid;not null;index:idx_code_chat" json:"chat_id"`
	Language    CodeLanguage  `gorm:"type:varchar(50);not null" json:"language"`
	Code        string        `gorm:"type:text;not null" json:"code"`
	FileName    *string       `gorm:"type:varchar(255)" json:"file_name,omitempty"`
	CreatedByID uuid.UUID     `gorm:"type:uuid;not null" json:"created_by_id"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`

	CreatedBy *User  `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`
	Message   *Message `gorm:"foreignKey:MessageID" json:"message,omitempty"`
	Chat      *Chat  `gorm:"foreignKey:ChatID" json:"chat,omitempty"`
}

type CreateCodeSnippetRequest struct {
	MessageID string       `json:"message_id" validate:"required,uuid"`
	ChatID    string       `json:"chat_id" validate:"required,uuid"`
	Language  CodeLanguage `json:"language" validate:"required"`
	Code      string       `json:"code" validate:"required"`
	FileName  *string      `json:"file_name" validate:"omitempty,max=255"`
}

type UpdateCodeSnippetRequest struct {
	Language *CodeLanguage `json:"language" validate:"omitempty"`
	Code     *string       `json:"code" validate:"omitempty,min=1"`
	FileName *string       `json:"file_name" validate:"omitempty,max=255"`
}

type CodeSnippetResponse struct {
	ID          uuid.UUID     `json:"id"`
	MessageID   uuid.UUID     `json:"message_id"`
	ChatID      uuid.UUID     `json:"chat_id"`
	Language    CodeLanguage  `json:"language"`
	Code        string        `json:"code"`
	FileName    *string       `json:"file_name,omitempty"`
	CreatedByID uuid.UUID     `json:"created_by_id"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	CreatedBy   *UserResponse `json:"created_by,omitempty"`
}

func (c *CodeSnippet) ToResponse() CodeSnippetResponse {
	resp := CodeSnippetResponse{
		ID:          c.ID,
		MessageID:   c.MessageID,
		ChatID:      c.ChatID,
		Language:    c.Language,
		Code:        c.Code,
		FileName:    c.FileName,
		CreatedByID: c.CreatedByID,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}

	if c.CreatedBy != nil {
		userResp := c.CreatedBy.ToResponse()
		resp.CreatedBy = &userResp
	}

	return resp
}
