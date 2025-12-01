package cdc

type DebeziumPayload struct {
	Schema  struct{} `json:"schema"`
	Payload struct {
		Before   map[string]interface{} `json:"before"`
		After    map[string]interface{} `json:"after"`
		Source   map[string]interface{} `json:"source"`
		Op       string                 `json:"op"`
		TsMs     int64                  `json:"ts_ms"`
		Sequence string                 `json:"sequence"`
	} `json:"payload"`
}
