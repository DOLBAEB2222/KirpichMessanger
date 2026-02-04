package tests

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/messenger/backend/internal/models"
	"github.com/messenger/backend/internal/services"
)

func TestCallService_InitiateCall(t *testing.T) {
	db := setupTestDB(t)
	chatService := services.NewChatService(db, nil)

	user1 := createTestUser(t, db, "+1000000001")
	user2 := createTestUser(t, db, "+1000000002")

	chat, err := chatService.GetOrCreateDMChat(nil, user1.ID, user2.ID)
	if err != nil {
		t.Fatalf("failed to create chat: %v", err)
	}

	t.Run("Create voice call", func(t *testing.T) {
		call := models.Call{
			ChatID:      chat.ID,
			InitiatorID: user1.ID,
			RecipientID: user2.ID,
			Type:        models.CallTypeVoice,
			Status:      models.CallStatusRinging,
		}

		if err := db.Create(&call).Error; err != nil {
			t.Fatalf("failed to create call: %v", err)
		}

		if call.ID == uuid.Nil {
			t.Error("expected call ID to be set")
		}

		if call.Status != models.CallStatusRinging {
			t.Errorf("expected status ringing, got %s", call.Status)
		}

		if call.Type != models.CallTypeVoice {
			t.Errorf("expected type voice, got %s", call.Type)
		}
	})

	t.Run("Create video call", func(t *testing.T) {
		call := models.Call{
			ChatID:      chat.ID,
			InitiatorID: user1.ID,
			RecipientID: user2.ID,
			Type:        models.CallTypeVideo,
			Status:      models.CallStatusRinging,
		}

		if err := db.Create(&call).Error; err != nil {
			t.Fatalf("failed to create call: %v", err)
		}

		if call.Type != models.CallTypeVideo {
			t.Errorf("expected type video, got %s", call.Type)
		}
	})
}

func TestCallService_AcceptRejectCall(t *testing.T) {
	db := setupTestDB(t)
	chatService := services.NewChatService(db, nil)

	user1 := createTestUser(t, db, "+1000000003")
	user2 := createTestUser(t, db, "+1000000004")

	chat, err := chatService.GetOrCreateDMChat(nil, user1.ID, user2.ID)
	if err != nil {
		t.Fatalf("failed to create chat: %v", err)
	}

	t.Run("Accept call", func(t *testing.T) {
		call := models.Call{
			ChatID:      chat.ID,
			InitiatorID: user1.ID,
			RecipientID: user2.ID,
			Type:        models.CallTypeVoice,
			Status:      models.CallStatusRinging,
		}
		db.Create(&call)

		call.SetStarted()
		if err := db.Save(&call).Error; err != nil {
			t.Fatalf("failed to update call: %v", err)
		}

		if call.Status != models.CallStatusAccepted {
			t.Errorf("expected status accepted, got %s", call.Status)
		}

		if call.StartedAt == nil {
			t.Error("expected started_at to be set")
		}
	})

	t.Run("Reject call", func(t *testing.T) {
		call := models.Call{
			ChatID:      chat.ID,
			InitiatorID: user1.ID,
			RecipientID: user2.ID,
			Type:        models.CallTypeVoice,
			Status:      models.CallStatusRinging,
		}
		db.Create(&call)

		call.Status = models.CallStatusRejected
		now := time.Now()
		call.EndedAt = &now

		if err := db.Save(&call).Error; err != nil {
			t.Fatalf("failed to update call: %v", err)
		}

		if call.Status != models.CallStatusRejected {
			t.Errorf("expected status rejected, got %s", call.Status)
		}
	})
}

func TestCallService_EndCall(t *testing.T) {
	db := setupTestDB(t)
	chatService := services.NewChatService(db, nil)

	user1 := createTestUser(t, db, "+1000000005")
	user2 := createTestUser(t, db, "+1000000006")

	chat, err := chatService.GetOrCreateDMChat(nil, user1.ID, user2.ID)
	if err != nil {
		t.Fatalf("failed to create chat: %v", err)
	}

	t.Run("End active call", func(t *testing.T) {
		call := models.Call{
			ChatID:      chat.ID,
			InitiatorID: user1.ID,
			RecipientID: user2.ID,
			Type:        models.CallTypeVoice,
			Status:      models.CallStatusAccepted,
		}
		call.SetStarted()
		// Simulate some time passing
		time.Sleep(100 * time.Millisecond)
		db.Create(&call)

		call.SetEnded()
		if err := db.Save(&call).Error; err != nil {
			t.Fatalf("failed to end call: %v", err)
		}

		if call.Status != models.CallStatusEnded {
			t.Errorf("expected status ended, got %s", call.Status)
		}

		if call.EndedAt == nil {
			t.Error("expected ended_at to be set")
		}

		if call.Duration < 0 {
			t.Error("expected positive duration")
		}
	})
}

func TestCallService_CallUniqueness(t *testing.T) {
	db := setupTestDB(t)
	chatService := services.NewChatService(db, nil)

	user1 := createTestUser(t, db, "+1000000007")
	user2 := createTestUser(t, db, "+1000000008")

	chat, err := chatService.GetOrCreateDMChat(nil, user1.ID, user2.ID)
	if err != nil {
		t.Fatalf("failed to create chat: %v", err)
	}

	t.Run("Prevent duplicate active calls", func(t *testing.T) {
		call1 := models.Call{
			ChatID:      chat.ID,
			InitiatorID: user1.ID,
			RecipientID: user2.ID,
			Type:        models.CallTypeVoice,
			Status:      models.CallStatusRinging,
		}
		db.Create(&call1)

		// Check for existing active call
		var existingCall models.Call
		db.Where("(initiator_id = ? OR recipient_id = ?) AND (initiator_id = ? OR recipient_id = ?) AND status IN ?",
			user1.ID, user1.ID, user2.ID, user2.ID, []models.CallStatus{models.CallStatusRinging, models.CallStatusAccepted}).
			First(&existingCall)

		if existingCall.ID == uuid.Nil {
			t.Error("expected to find existing active call")
		}

		if existingCall.ID != call1.ID {
			t.Error("expected to find the same call")
		}
	})
}

func TestCallService_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		status   models.CallStatus
		expected bool
	}{
		{"Ringing", models.CallStatusRinging, true},
		{"Accepted", models.CallStatusAccepted, true},
		{"Rejected", models.CallStatusRejected, false},
		{"Ended", models.CallStatusEnded, false},
		{"Missed", models.CallStatusMissed, false},
		{"Busy", models.CallStatusBusy, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := models.Call{Status: tt.status}
			if call.IsActive() != tt.expected {
				t.Errorf("IsActive() = %v, want %v", call.IsActive(), tt.expected)
			}
		})
	}
}

func TestCallService_CanJoin(t *testing.T) {
	db := setupTestDB(t)
	user1 := createTestUser(t, db, "+1000000009")
	user2 := createTestUser(t, db, "+1000000010")
	user3 := createTestUser(t, db, "+1000000011")

	chatService := services.NewChatService(db, nil)
	chat, _ := chatService.GetOrCreateDMChat(nil, user1.ID, user2.ID)

	call := models.Call{
		ChatID:      chat.ID,
		InitiatorID: user1.ID,
		RecipientID: user2.ID,
		Type:        models.CallTypeVoice,
		Status:      models.CallStatusAccepted,
	}

	t.Run("Initiator can join", func(t *testing.T) {
		if !call.CanJoin(user1.ID) {
			t.Error("expected initiator to be able to join")
		}
	})

	t.Run("Recipient can join", func(t *testing.T) {
		if !call.CanJoin(user2.ID) {
			t.Error("expected recipient to be able to join")
		}
	})

	t.Run("Third party cannot join", func(t *testing.T) {
		if call.CanJoin(user3.ID) {
			t.Error("expected third party to not be able to join")
		}
	})

	t.Run("Cannot join ended call", func(t *testing.T) {
		call.Status = models.CallStatusEnded
		if call.CanJoin(user1.ID) {
			t.Error("expected cannot join ended call")
		}
	})
}

func TestCallModel_ToResponse(t *testing.T) {
	now := time.Now()
	call := models.Call{
		ID:          uuid.New(),
		ChatID:      uuid.New(),
		InitiatorID: uuid.New(),
		RecipientID: uuid.New(),
		Type:        models.CallTypeVoice,
		Status:      models.CallStatusAccepted,
		Duration:    125,
		CreatedAt:   now,
	}
	call.SetStarted()
	call.SetEnded()

	resp := call.ToResponse()

	if resp.ID != call.ID {
		t.Error("ID mismatch")
	}
	if resp.ChatID != call.ChatID {
		t.Error("ChatID mismatch")
	}
	if resp.InitiatorID != call.InitiatorID {
		t.Error("InitiatorID mismatch")
	}
	if resp.RecipientID != call.RecipientID {
		t.Error("RecipientID mismatch")
	}
	if resp.Type != call.Type {
		t.Error("Type mismatch")
	}
	if resp.Status != call.Status {
		t.Error("Status mismatch")
	}
	if resp.Duration != call.Duration {
		t.Error("Duration mismatch")
	}
}

func TestCallSignalModel(t *testing.T) {
	db := setupTestDB(t)
	chatService := services.NewChatService(db, nil)

	user1 := createTestUser(t, db, "+1000000012")
	user2 := createTestUser(t, db, "+1000000013")

	chat, _ := chatService.GetOrCreateDMChat(nil, user1.ID, user2.ID)

	call := models.Call{
		ChatID:      chat.ID,
		InitiatorID: user1.ID,
		RecipientID: user2.ID,
		Type:        models.CallTypeVoice,
		Status:      models.CallStatusAccepted,
	}
	db.Create(&call)

	t.Run("Create call signal", func(t *testing.T) {
		signal := models.CallSignal{
			CallID: call.ID,
			Type:   "offer",
			Data:   `{"type":"offer","sdp":"v=0\r\no=- 12345..."}`,
		}

		if err := db.Create(&signal).Error; err != nil {
			t.Fatalf("failed to create signal: %v", err)
		}

		if signal.ID == uuid.Nil {
			t.Error("expected signal ID to be set")
		}

		if signal.Type != "offer" {
			t.Errorf("expected type offer, got %s", signal.Type)
		}
	})

	t.Run("Create multiple signals", func(t *testing.T) {
		signals := []models.CallSignal{
			{CallID: call.ID, Type: "offer", Data: "offer_data"},
			{CallID: call.ID, Type: "answer", Data: "answer_data"},
			{CallID: call.ID, Type: "candidate", Data: "candidate_data"},
		}

		for _, signal := range signals {
			if err := db.Create(&signal).Error; err != nil {
				t.Fatalf("failed to create signal: %v", err)
			}
		}

		var count int64
		db.Model(&models.CallSignal{}).Where("call_id = ?", call.ID).Count(&count)
		if count < 3 {
			t.Errorf("expected at least 3 signals, got %d", count)
		}
	})
}

func TestICEServersResponse(t *testing.T) {
	resp := models.ICEServersResponse{
		ICEServers: []models.ICEServer{
			{
				URLs:       []string{"turn:turn.example.com:3478"},
				Username:   "user",
				Credential: "pass",
			},
			{
				URLs: []string{"stun:stun.example.com:3478"},
			},
		},
	}

	if len(resp.ICEServers) != 2 {
		t.Errorf("expected 2 ICE servers, got %d", len(resp.ICEServers))
	}

	if len(resp.ICEServers[0].URLs) != 1 {
		t.Error("expected 1 URL for first server")
	}

	if resp.ICEServers[0].Username != "user" {
		t.Error("username mismatch")
	}

	if resp.ICEServers[1].Username != "" {
		t.Error("STUN server should not have username")
	}
}
