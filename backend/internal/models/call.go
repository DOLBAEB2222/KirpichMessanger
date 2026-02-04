package models

import (
    "encoding/json"
    "time"

    "github.com/google/uuid"
)

type CallType string

const (
    CallTypeVoice CallType = "voice"
    CallTypeVideo CallType = "video"
)

type CallStatus string

const (
    CallStatusRinging  CallStatus = "ringing"
    CallStatusAccepted CallStatus = "accepted"
    CallStatusRejected CallStatus = "rejected"
    CallStatusMissed   CallStatus = "missed"
    CallStatusEnded    CallStatus = "ended"
    CallStatusBusy     CallStatus = "busy"
)

type Call struct {
    ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
    ChatID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"chat_id"`
    InitiatorID uuid.UUID  `gorm:"type:uuid;not null;index" json:"initiator_id"`
    RecipientID uuid.UUID  `gorm:"type:uuid;not null;index" json:"recipient_id"`
    Type        CallType   `gorm:"type:varchar(20);not null" json:"type"`
    Status      CallStatus `gorm:"type:varchar(20);not null;default:'ringing'" json:"status"`
    Duration    int64      `gorm:"default:0" json:"duration"`
    StartedAt   *time.Time `json:"started_at,omitempty"`
    EndedAt     *time.Time `json:"ended_at,omitempty"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`

    Chat      *Chat `gorm:"foreignKey:ChatID" json:"chat,omitempty"`
    Initiator *User `gorm:"foreignKey:InitiatorID" json:"initiator,omitempty"`
    Recipient *User `gorm:"foreignKey:RecipientID" json:"recipient,omitempty"`
}

type CallSignal struct {
    ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
    CallID    uuid.UUID `gorm:"type:uuid;not null;index" json:"call_id"`
    Type      string    `gorm:"type:varchar(20);not null" json:"type"`
    Data      string    `gorm:"type:text" json:"data"`
    CreatedAt time.Time `json:"created_at"`

    Call *Call `gorm:"foreignKey:CallID" json:"call,omitempty"`
}

type InitiateCallRequest struct {
    ChatID      string   `json:"chat_id" validate:"required,uuid"`
    RecipientID string   `json:"recipient_id" validate:"required,uuid"`
    CallType    CallType `json:"call_type" validate:"required,oneof=voice video"`
}

type CallResponseRequest struct {
    Accept bool `json:"accept"`
}

type CallResponse struct {
    ID          uuid.UUID  `json:"id"`
    ChatID      uuid.UUID  `json:"chat_id"`
    InitiatorID uuid.UUID  `json:"initiator_id"`
    RecipientID uuid.UUID  `json:"recipient_id"`
    Type        CallType   `json:"type"`
    Status      CallStatus `json:"status"`
    Duration    int64      `json:"duration"`
    StartedAt   *time.Time `json:"started_at,omitempty"`
    EndedAt     *time.Time `json:"ended_at,omitempty"`
    CreatedAt   time.Time  `json:"created_at"`
}

type WebRTCSignal struct {
    Type      string          `json:"type"`
    CallID    string          `json:"call_id"`
    UserID    string          `json:"user_id"`
    Offer     json.RawMessage `json:"offer,omitempty"`
    Answer    json.RawMessage `json:"answer,omitempty"`
    Candidate json.RawMessage `json:"candidate,omitempty"`
    Timestamp int64           `json:"timestamp"`
}

type ICEServer struct {
    URLs       []string `json:"urls"`
    Username   string   `json:"username,omitempty"`
    Credential string   `json:"credential,omitempty"`
}

type ICEServersResponse struct {
    ICEServers []ICEServer `json:"iceServers"`
}

func (c *Call) ToResponse() CallResponse {
    return CallResponse{
        ID:          c.ID,
        ChatID:      c.ChatID,
        InitiatorID: c.InitiatorID,
        RecipientID: c.RecipientID,
        Type:        c.Type,
        Status:      c.Status,
        Duration:    c.Duration,
        StartedAt:   c.StartedAt,
        EndedAt:     c.EndedAt,
        CreatedAt:   c.CreatedAt,
    }
}

func (c *Call) IsActive() bool {
    return c.Status == CallStatusRinging || c.Status == CallStatusAccepted
}

func (c *Call) CanJoin(userID uuid.UUID) bool {
    if !c.IsActive() {
        return false
    }
    return c.InitiatorID == userID || c.RecipientID == userID
}

func (c *Call) SetStarted() {
    now := time.Now()
    c.StartedAt = &now
    c.Status = CallStatusAccepted
}

func (c *Call) SetEnded() {
    now := time.Now()
    c.EndedAt = &now
    c.Status = CallStatusEnded
    if c.StartedAt != nil {
        c.Duration = int64(now.Sub(*c.StartedAt).Seconds())
    }
}
