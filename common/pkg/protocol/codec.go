package protocol

import (
	"github.com/vmihailenco/msgpack/v5"
)

// Encode 编码为MsgPack
func Encode(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

// Decode 解码MsgPack
func Decode(data []byte, v interface{}) error {
	return msgpack.Unmarshal(data, v)
}

// EncodeEnvelope 编码封包
func EncodeEnvelope(env *Envelope) ([]byte, error) {
	return Encode(env)
}

// DecodeEnvelope 解码封包
func DecodeEnvelope(data []byte) (*Envelope, error) {
	var env Envelope
	if err := Decode(data, &env); err != nil {
		return nil, err
	}
	return &env, nil
}

// EncodeEvent 编码事件
func EncodeEvent(event *Event) ([]byte, error) {
	return Encode(event)
}

// DecodeEvent 解码事件
func DecodeEvent(data []byte) (*Event, error) {
	var event Event
	if err := Decode(data, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

// DecodeEventFromBody 从封包Body解码事件
func DecodeEventFromBody(body interface{}) (*Event, error) {
	// 如果body已经是*Event类型
	if event, ok := body.(*Event); ok {
		return event, nil
	}

	// 如果body是map类型，需要重新编解码
	data, err := Encode(body)
	if err != nil {
		return nil, err
	}
	return DecodeEvent(data)
}
