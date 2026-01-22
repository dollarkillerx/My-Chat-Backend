package protocol

import (
	"testing"
)

func TestGetReplyMid(t *testing.T) {
	tests := []struct {
		name    string
		tags    []Tag
		wantMid int64
		wantOk  bool
	}{
		{
			name:    "empty tags",
			tags:    []Tag{},
			wantMid: 0,
			wantOk:  false,
		},
		{
			name: "has reply tag",
			tags: []Tag{
				{Type: TagReply, Value: int64(123)},
			},
			wantMid: 123,
			wantOk:  true,
		},
		{
			name: "reply tag with other tags",
			tags: []Tag{
				{Type: TagMention, Value: "user1"},
				{Type: TagReply, Value: int64(456)},
				{Type: TagMention, Value: "user2"},
			},
			wantMid: 456,
			wantOk:  true,
		},
		{
			name: "no reply tag",
			tags: []Tag{
				{Type: TagMention, Value: "user1"},
				{Type: TagTarget, Value: int64(789)},
			},
			wantMid: 0,
			wantOk:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMid, gotOk := GetReplyMid(tt.tags)
			if gotMid != tt.wantMid {
				t.Errorf("GetReplyMid() mid = %v, want %v", gotMid, tt.wantMid)
			}
			if gotOk != tt.wantOk {
				t.Errorf("GetReplyMid() ok = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestGetMentions(t *testing.T) {
	tests := []struct {
		name     string
		tags     []Tag
		wantUids []string
	}{
		{
			name:     "empty tags",
			tags:     []Tag{},
			wantUids: nil,
		},
		{
			name: "single mention",
			tags: []Tag{
				{Type: TagMention, Value: "user1"},
			},
			wantUids: []string{"user1"},
		},
		{
			name: "multiple mentions",
			tags: []Tag{
				{Type: TagMention, Value: "user1"},
				{Type: TagMention, Value: "user2"},
				{Type: TagMention, Value: "user3"},
			},
			wantUids: []string{"user1", "user2", "user3"},
		},
		{
			name: "mixed tags",
			tags: []Tag{
				{Type: TagReply, Value: int64(100)},
				{Type: TagMention, Value: "user1"},
				{Type: TagTarget, Value: int64(200)},
				{Type: TagMention, Value: "user2"},
			},
			wantUids: []string{"user1", "user2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUids := GetMentions(tt.tags)
			if len(gotUids) != len(tt.wantUids) {
				t.Errorf("GetMentions() length = %v, want %v", len(gotUids), len(tt.wantUids))
				return
			}
			for i, uid := range gotUids {
				if uid != tt.wantUids[i] {
					t.Errorf("GetMentions()[%d] = %v, want %v", i, uid, tt.wantUids[i])
				}
			}
		})
	}
}

func TestGetTargetMid(t *testing.T) {
	tests := []struct {
		name    string
		tags    []Tag
		wantMid int64
		wantOk  bool
	}{
		{
			name:    "empty tags",
			tags:    []Tag{},
			wantMid: 0,
			wantOk:  false,
		},
		{
			name: "has target tag",
			tags: []Tag{
				{Type: TagTarget, Value: int64(999)},
			},
			wantMid: 999,
			wantOk:  true,
		},
		{
			name: "no target tag",
			tags: []Tag{
				{Type: TagReply, Value: int64(100)},
				{Type: TagMention, Value: "user1"},
			},
			wantMid: 0,
			wantOk:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMid, gotOk := GetTargetMid(tt.tags)
			if gotMid != tt.wantMid {
				t.Errorf("GetTargetMid() mid = %v, want %v", gotMid, tt.wantMid)
			}
			if gotOk != tt.wantOk {
				t.Errorf("GetTargetMid() ok = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestNewReplyTag(t *testing.T) {
	tag := NewReplyTag(123)

	if tag.Type != TagReply {
		t.Errorf("Type mismatch: got %d, want %d", tag.Type, TagReply)
	}

	if tag.Value != int64(123) {
		t.Errorf("Value mismatch: got %v, want 123", tag.Value)
	}
}

func TestNewMentionTag(t *testing.T) {
	tag := NewMentionTag("user123")

	if tag.Type != TagMention {
		t.Errorf("Type mismatch: got %d, want %d", tag.Type, TagMention)
	}

	if tag.Value != "user123" {
		t.Errorf("Value mismatch: got %v, want user123", tag.Value)
	}
}

func TestNewMentionAllTag(t *testing.T) {
	tag := NewMentionAllTag()

	if tag.Type != TagMention {
		t.Errorf("Type mismatch: got %d, want %d", tag.Type, TagMention)
	}

	if tag.Value != "all" {
		t.Errorf("Value mismatch: got %v, want all", tag.Value)
	}
}

func TestNewTargetTag(t *testing.T) {
	tag := NewTargetTag(456)

	if tag.Type != TagTarget {
		t.Errorf("Type mismatch: got %d, want %d", tag.Type, TagTarget)
	}

	if tag.Value != int64(456) {
		t.Errorf("Value mismatch: got %v, want 456", tag.Value)
	}
}

func TestNewForwardCidTag(t *testing.T) {
	tag := NewForwardCidTag("conv123")

	if tag.Type != TagForwardCid {
		t.Errorf("Type mismatch: got %d, want %d", tag.Type, TagForwardCid)
	}

	if tag.Value != "conv123" {
		t.Errorf("Value mismatch: got %v, want conv123", tag.Value)
	}
}

func TestNewForwardMidTag(t *testing.T) {
	tag := NewForwardMidTag(789)

	if tag.Type != TagForwardMid {
		t.Errorf("Type mismatch: got %d, want %d", tag.Type, TagForwardMid)
	}

	if tag.Value != int64(789) {
		t.Errorf("Value mismatch: got %v, want 789", tag.Value)
	}
}

func TestParseTags(t *testing.T) {
	tags := []Tag{
		{Type: TagReply, Value: int64(100)},
		{Type: TagMention, Value: "user1"},
		{Type: TagMention, Value: "user2"},
		{Type: TagTarget, Value: int64(200)},
	}

	result := ParseTags(tags)

	if len(result[TagReply]) != 1 {
		t.Errorf("Reply count mismatch: got %d, want 1", len(result[TagReply]))
	}

	if len(result[TagMention]) != 2 {
		t.Errorf("Mention count mismatch: got %d, want 2", len(result[TagMention]))
	}

	if len(result[TagTarget]) != 1 {
		t.Errorf("Target count mismatch: got %d, want 1", len(result[TagTarget]))
	}
}
