package res

import "time"

type ResponseHttp[T any] struct {
	Timestamp time.Time `json:"timestamp"`
	Payload   T         `json:"payload"`
	Code      int16     `json:"code"`
	Status    bool      `json:"status"`
	Message   string    `json:"message"`
	Version   uint8     `json:"version"`
	TraceID   string    `json:"trace_id,omitempty"`
	Path      string    `json:"path,omitempty"`
}
