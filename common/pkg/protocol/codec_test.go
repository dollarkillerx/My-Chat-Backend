package protocol

import (
	"testing"
)

func TestEncode(t *testing.T) {
	env := NewEnvelope(CmdEvent, 1, map[string]string{"key": "value"})

	data, err := Encode(env)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Encode returned empty data")
	}
}

func TestDecode(t *testing.T) {
	original := NewEnvelope(CmdPing, 123, nil)

	data, err := Encode(original)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	var decoded Envelope
	err = Decode(data, &decoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if decoded.Cmd != original.Cmd {
		t.Errorf("Cmd mismatch: got %s, want %s", decoded.Cmd, original.Cmd)
	}

	if decoded.Seq != original.Seq {
		t.Errorf("Seq mismatch: got %d, want %d", decoded.Seq, original.Seq)
	}
}

func TestDecodeEnvelope(t *testing.T) {
	original := &Envelope{
		Version: 1,
		Cmd:     CmdEvent,
		Seq:     456,
		Body:    "test body",
	}

	data, err := Encode(original)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	decoded, err := DecodeEnvelope(data)
	if err != nil {
		t.Fatalf("DecodeEnvelope failed: %v", err)
	}

	if decoded.Cmd != original.Cmd {
		t.Errorf("Cmd mismatch: got %s, want %s", decoded.Cmd, original.Cmd)
	}

	if decoded.Seq != original.Seq {
		t.Errorf("Seq mismatch: got %d, want %d", decoded.Seq, original.Seq)
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	testCases := []struct {
		name string
		env  *Envelope
	}{
		{
			name: "ping",
			env:  NewEnvelope(CmdPing, 1, nil),
		},
		{
			name: "pong",
			env:  NewEnvelope(CmdPong, 2, nil),
		},
		{
			name: "event with body",
			env:  NewEnvelope(CmdEvent, 100, map[string]interface{}{"msg": "hello"}),
		},
		{
			name: "ack",
			env:  NewEnvelope(CmdAck, 50, &AckBody{Seq: 50, Mid: 12345}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := Encode(tc.env)
			if err != nil {
				t.Fatalf("Encode failed: %v", err)
			}

			var decoded Envelope
			err = Decode(data, &decoded)
			if err != nil {
				t.Fatalf("Decode failed: %v", err)
			}

			if decoded.Cmd != tc.env.Cmd {
				t.Errorf("Cmd mismatch: got %s, want %s", decoded.Cmd, tc.env.Cmd)
			}

			if decoded.Seq != tc.env.Seq {
				t.Errorf("Seq mismatch: got %d, want %d", decoded.Seq, tc.env.Seq)
			}
		})
	}
}

func TestEncodeDecodeEvent(t *testing.T) {
	original := NewEvent(KindText, "conv123", "user456")
	original.SetText("Hello, World!")
	original.Mid = 789
	original.Timestamp = 1234567890

	data, err := EncodeEvent(original)
	if err != nil {
		t.Fatalf("EncodeEvent failed: %v", err)
	}

	decoded, err := DecodeEvent(data)
	if err != nil {
		t.Fatalf("DecodeEvent failed: %v", err)
	}

	if decoded.Cid != original.Cid {
		t.Errorf("Cid mismatch: got %s, want %s", decoded.Cid, original.Cid)
	}

	if decoded.Kind != original.Kind {
		t.Errorf("Kind mismatch: got %d, want %d", decoded.Kind, original.Kind)
	}

	if decoded.Sender != original.Sender {
		t.Errorf("Sender mismatch: got %s, want %s", decoded.Sender, original.Sender)
	}

	if decoded.GetText() != "Hello, World!" {
		t.Errorf("Text mismatch: got %s, want Hello, World!", decoded.GetText())
	}
}
