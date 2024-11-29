package model

type ExternalSignUpRequest struct {
	Iaid      string `json:"iaid"`
	Eaid      string `json:"eaid"`
	Signature string `json:"signature"`
	Name      string `json:"name"`
	Login     string `json:"login"`
	Phone     string `json:"phone"`
	Password  string `json:"password"`
}

type InternalSignUpRequest struct {
	Ip             string `json:"ip"`
	DeviceName     string `json:"device_name"`
	DevicePlatform string `json:"device_platform"`
}

type ClientSignUpRequest struct {
	Iaid         string `json:"iaid"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type EnterSession struct {
	Iaid      string `json:"iaid"`
	Signature string `json:"signature"`
	InternalSignUpRequest
}
