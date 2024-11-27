package model

type AuthDTO struct{ AccessToken, RefreshToken, Signature string }

type DeviceInfo struct {
	DeviceName    string
	OSName        string
	OSVersion     string
	Client        string
	ClientVersion string
}
