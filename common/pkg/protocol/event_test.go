package protocol

import (
	"testing"
)

func TestNewEvent(t *testing.T) {
	event := NewEvent(KindText, "conv123", "user456")

	if event.Cid != "conv123" {
		t.Errorf("Cid mismatch: got %s, want conv123", event.Cid)
	}

	if event.Kind != KindText {
		t.Errorf("Kind mismatch: got %d, want %d", event.Kind, KindText)
	}

	if event.Sender != "user456" {
		t.Errorf("Sender mismatch: got %s, want user456", event.Sender)
	}

	if event.Version != 1 {
		t.Errorf("Version mismatch: got %d, want 1", event.Version)
	}
}

func TestSetText(t *testing.T) {
	event := NewEvent(KindText, "conv123", "user456")
	event.SetText("Hello, World!")

	text := event.GetText()
	if text != "Hello, World!" {
		t.Errorf("Text mismatch: got %s, want Hello, World!", text)
	}
}

func TestSetFileData(t *testing.T) {
	event := NewEvent(KindFile, "conv123", "user456")
	event.SetFileData("fid123", "document.pdf", 1024, "application/pdf", "sha256hash", "https://example.com/file.pdf")

	fd := event.GetFileData()

	if fd.Fid != "fid123" {
		t.Errorf("Fid mismatch: got %s", fd.Fid)
	}

	if fd.Name != "document.pdf" {
		t.Errorf("Name mismatch: got %s", fd.Name)
	}

	if fd.Size != 1024 {
		t.Errorf("Size mismatch: got %d", fd.Size)
	}

	if fd.Mime != "application/pdf" {
		t.Errorf("Mime mismatch: got %s", fd.Mime)
	}

	if fd.SHA256 != "sha256hash" {
		t.Errorf("SHA256 mismatch: got %s", fd.SHA256)
	}

	if fd.URL != "https://example.com/file.pdf" {
		t.Errorf("URL mismatch: got %s", fd.URL)
	}
}

func TestAddReplyTag(t *testing.T) {
	event := NewEvent(KindText, "conv123", "user456")
	event.AddReplyTag(789)

	if len(event.Tags) != 1 {
		t.Fatalf("Tags length mismatch: got %d, want 1", len(event.Tags))
	}

	tag := event.Tags[0]
	if tag.Type != TagReply {
		t.Errorf("Tag type mismatch: got %d, want %d", tag.Type, TagReply)
	}

	if tag.Value != int64(789) {
		t.Errorf("Tag value mismatch: got %v, want 789", tag.Value)
	}
}

func TestAddMentionTag(t *testing.T) {
	event := NewEvent(KindText, "conv123", "user456")
	event.AddMentionTag("user1")
	event.AddMentionTag("user2")
	event.AddMentionTag("user3")

	if len(event.Tags) != 3 {
		t.Fatalf("Tags length mismatch: got %d, want 3", len(event.Tags))
	}

	for i, tag := range event.Tags {
		if tag.Type != TagMention {
			t.Errorf("Tag[%d] type mismatch: got %d, want %d", i, tag.Type, TagMention)
		}
	}
}

func TestSetRevokeData(t *testing.T) {
	event := NewEvent(KindRevoke, "conv123", "user456")
	event.SetRevokeData(100, 1, "inappropriate content")

	if len(event.Tags) != 1 {
		t.Fatalf("Tags length mismatch: got %d, want 1", len(event.Tags))
	}

	tag := event.Tags[0]
	if tag.Type != TagTarget {
		t.Errorf("Tag type mismatch: got %d, want %d", tag.Type, TagTarget)
	}

	if tag.Value != int64(100) {
		t.Errorf("Tag value mismatch: got %v, want 100", tag.Value)
	}

	if event.Data[0] != 1 {
		t.Errorf("Scope mismatch: got %v, want 1", event.Data[0])
	}

	if event.Data[1] != "inappropriate content" {
		t.Errorf("Reason mismatch: got %v", event.Data[1])
	}
}

func TestSetEditData(t *testing.T) {
	event := NewEvent(KindEdit, "conv123", "user456")
	event.SetEditData(200, "updated content", 2)

	if len(event.Tags) != 1 {
		t.Fatalf("Tags length mismatch: got %d, want 1", len(event.Tags))
	}

	tag := event.Tags[0]
	if tag.Type != TagTarget {
		t.Errorf("Tag type mismatch: got %d, want %d", tag.Type, TagTarget)
	}

	if event.Data[0] != "updated content" {
		t.Errorf("Content mismatch: got %v", event.Data[0])
	}

	if event.Data[1] != 2 {
		t.Errorf("Version mismatch: got %v", event.Data[1])
	}
}

func TestSetReadReceipt(t *testing.T) {
	event := NewEvent(KindReadReceipt, "conv123", "user456")
	event.SetReadReceipt(500)

	if event.Data[0] != int64(500) {
		t.Errorf("LastReadMid mismatch: got %v, want 500", event.Data[0])
	}
}

func TestSetTyping(t *testing.T) {
	event := NewEvent(KindTyping, "conv123", "user456")
	event.SetTyping(1)

	if event.Data[0] != 1 {
		t.Errorf("Typing state mismatch: got %v, want 1", event.Data[0])
	}
}

func TestSetReaction(t *testing.T) {
	event := NewEvent(KindReaction, "conv123", "user456")
	event.SetReaction(300, "👍", 1)

	if len(event.Tags) != 1 {
		t.Fatalf("Tags length mismatch: got %d, want 1", len(event.Tags))
	}

	if event.Data[0] != "👍" {
		t.Errorf("Emoji mismatch: got %v", event.Data[0])
	}

	if event.Data[1] != 1 {
		t.Errorf("Action mismatch: got %v", event.Data[1])
	}
}

func TestSetForward(t *testing.T) {
	event := NewEvent(KindForward, "conv123", "user456")
	event.SetForward("originalCid", 999, ForwardTypeSingle, nil)

	if len(event.Tags) != 2 {
		t.Fatalf("Tags length mismatch: got %d, want 2", len(event.Tags))
	}

	// Check forward cid tag
	cidTag := event.Tags[0]
	if cidTag.Type != TagForwardCid {
		t.Errorf("First tag type mismatch: got %d, want %d", cidTag.Type, TagForwardCid)
	}
	if cidTag.Value != "originalCid" {
		t.Errorf("First tag value mismatch: got %v, want originalCid", cidTag.Value)
	}

	// Check forward mid tag
	midTag := event.Tags[1]
	if midTag.Type != TagForwardMid {
		t.Errorf("Second tag type mismatch: got %d, want %d", midTag.Type, TagForwardMid)
	}
	if midTag.Value != int64(999) {
		t.Errorf("Second tag value mismatch: got %v, want 999", midTag.Value)
	}
}

func TestAddMentionAllTag(t *testing.T) {
	event := NewEvent(KindText, "conv123", "user456")
	event.AddMentionAllTag()

	if len(event.Tags) != 1 {
		t.Fatalf("Tags length mismatch: got %d, want 1", len(event.Tags))
	}

	tag := event.Tags[0]
	if tag.Type != TagMention {
		t.Errorf("Tag type mismatch: got %d, want %d", tag.Type, TagMention)
	}

	if tag.Value != "all" {
		t.Errorf("Tag value mismatch: got %v, want all", tag.Value)
	}
}
