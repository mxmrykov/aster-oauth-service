package model

type Response struct {
	Payload interface{} `json:"payload"`
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Error   bool        `json:"error"`
}

type SignUpResponse struct {
	Signature   string `json:"signature"`
	AccessToken string `json:"access_token"`
}
