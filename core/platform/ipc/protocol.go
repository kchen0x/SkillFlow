package ipc

type Request struct {
	Token   string `json:"token"`
	Command string `json:"command"`
}

type Response struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}
