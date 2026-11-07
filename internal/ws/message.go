package ws

import "encoding/json"

// inbound types (client → server)
const (
	TypeJoin  = "join"
	TypeLeave = "leave"
	TypePing  = "ping"
)

// outbound types (server → client)
const (
	TypeJoined = "joined"
	TypeLeft   = "left"
	TypePong   = "pong"
	TypeError  = "error"
)

// Message is the one JSON envelope used in both directions:
// {"type":"comment_added","data":{...}}
type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

func encode(typ string, data any) []byte {
	m := Message{Type: typ}
	if data != nil {
		if b, err := json.Marshal(data); err == nil {
			m.Data = b
		}
	}
	out, _ := json.Marshal(m)
	return out
}

func errMsg(code string) []byte {
	return encode(TypeError, map[string]string{"code": code})
}
