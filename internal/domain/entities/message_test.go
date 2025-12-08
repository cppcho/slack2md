package entities

import (
	"testing"
	"time"
)

func TestNewMessage(t *testing.T) {
	timestamp := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	text := "Hello, world!"
	userDisplayName := "John Doe"

	msg := NewMessage(timestamp, text, userDisplayName)

	if msg == nil {
		t.Fatal("Expected message to be created, got nil")
	}

	if !msg.Timestamp.Equal(timestamp) {
		t.Errorf("Expected timestamp = %v, got = %v", timestamp, msg.Timestamp)
	}

	if msg.Text != text {
		t.Errorf("Expected text = %q, got = %q", text, msg.Text)
	}

	if msg.UserDisplayName != userDisplayName {
		t.Errorf("Expected userDisplayName = %q, got = %q", userDisplayName, msg.UserDisplayName)
	}

	if msg.Replies == nil {
		t.Error("Expected Replies to be initialized (empty slice), got nil")
	}

	if len(msg.Replies) != 0 {
		t.Errorf("Expected Replies to be empty, got %d replies", len(msg.Replies))
	}

	if msg.IsParent {
		t.Error("Expected IsParent to be false for new message, got true")
	}

	if msg.ThreadTS != "" {
		t.Errorf("Expected ThreadTS to be empty, got %q", msg.ThreadTS)
	}
}

func TestNewMessage_WithEmptyFields(t *testing.T) {
	timestamp := time.Time{} // Zero time
	emptyText := ""
	emptyUser := ""

	// NewMessage doesn't validate - it's a simple constructor
	msg := NewMessage(timestamp, emptyText, emptyUser)

	if msg == nil {
		t.Fatal("Expected message to be created, got nil")
	}

	if msg.Text != emptyText {
		t.Errorf("Expected empty text, got = %q", msg.Text)
	}

	if msg.UserDisplayName != emptyUser {
		t.Errorf("Expected empty userDisplayName, got = %q", msg.UserDisplayName)
	}
}

func TestMessage_AddReply(t *testing.T) {
	parentTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	parent := NewMessage(parentTime, "Parent message", "Alice")

	// Verify initial state
	if parent.IsParent {
		t.Error("Expected new message to have IsParent = false initially")
	}
	if len(parent.Replies) != 0 {
		t.Error("Expected new message to have empty Replies initially")
	}

	// Add first reply
	reply1Time := time.Date(2024, 1, 15, 10, 5, 0, 0, time.UTC)
	reply1 := NewMessage(reply1Time, "First reply", "Bob")
	parent.AddReply(*reply1)

	if !parent.IsParent {
		t.Error("Expected IsParent to be true after adding reply")
	}
	if len(parent.Replies) != 1 {
		t.Errorf("Expected 1 reply, got %d", len(parent.Replies))
	}
	if parent.Replies[0].Text != "First reply" {
		t.Errorf("Expected reply text = %q, got = %q", "First reply", parent.Replies[0].Text)
	}

	// Add second reply
	reply2Time := time.Date(2024, 1, 15, 10, 10, 0, 0, time.UTC)
	reply2 := NewMessage(reply2Time, "Second reply", "Charlie")
	parent.AddReply(*reply2)

	if len(parent.Replies) != 2 {
		t.Errorf("Expected 2 replies, got %d", len(parent.Replies))
	}
	if parent.Replies[1].Text != "Second reply" {
		t.Errorf("Expected second reply text = %q, got = %q", "Second reply", parent.Replies[1].Text)
	}
}

func TestMessage_AddReply_Order(t *testing.T) {
	parent := NewMessage(time.Now(), "Parent", "User")

	// Add multiple replies
	for i := 1; i <= 5; i++ {
		replyTime := time.Now().Add(time.Duration(i) * time.Minute)
		reply := NewMessage(replyTime, "", "User")
		reply.Text = "Reply " + string(rune('0'+i))
		parent.AddReply(*reply)
	}

	// Verify order maintained
	if len(parent.Replies) != 5 {
		t.Fatalf("Expected 5 replies, got %d", len(parent.Replies))
	}

	for i := 0; i < 5; i++ {
		expectedText := "Reply " + string(rune('0'+i+1))
		if parent.Replies[i].Text != expectedText {
			t.Errorf("Reply %d: expected text = %q, got = %q", i, expectedText, parent.Replies[i].Text)
		}
	}
}

func TestMessage_AddReply_DoesNotModifyReply(t *testing.T) {
	parent := NewMessage(time.Now(), "Parent", "User")
	reply := NewMessage(time.Now(), "Reply", "User")

	// Store original reply state
	originalText := reply.Text
	originalUser := reply.UserDisplayName
	originalIsParent := reply.IsParent

	parent.AddReply(*reply)

	// Verify reply itself wasn't modified (pass by value)
	if reply.Text != originalText {
		t.Error("AddReply modified the reply's text")
	}
	if reply.UserDisplayName != originalUser {
		t.Error("AddReply modified the reply's userDisplayName")
	}
	if reply.IsParent != originalIsParent {
		t.Error("AddReply modified the reply's IsParent flag")
	}
}
