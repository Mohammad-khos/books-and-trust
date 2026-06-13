package contracts

type ResponseSuccess struct {
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
	Data      any    `json:"data_message,omitempty"`
	Timestamp string `json:"timestamp"`
}

type ResponseError struct {
	Status    string `json:"status"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	Details   any    `json:"details,omitempty"`
	Timestamp string `json:"timestamp"`
}
