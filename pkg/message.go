package pkg

// message sent by admin over WS.
type AdminMessage struct {
	Target  string `json:"target"`
	Message string `json:"message"`
}
