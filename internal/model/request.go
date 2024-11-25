package model

type SignupRequest struct {
	Name     string `binding:"required" json:"name"     form:"name"`
	Login    string `binding:"required" json:"login"    form:"login"`
	Phone    string `binding:"required" json:"phone"    form:"phone"`
	Password string `binding:"required" json:"password" form:"password"`
}

type AuthRequest struct {
	Login    string `binding:"required" json:"login"    form:"login"`
	Password string `binding:"required" json:"password" form:"password"`
}

type GetPhoneCodeRequest struct {
	Phone string `binding:"required" form:"p"`
}

type ConfirmPhoneCodeRequest struct {
	Phone string `binding:"required" json:"phone" form:"phone"`
	Code  int    `binding:"required" json:"code"  form:"code"`
}

type ExitRequest struct {
	Id int `binding:"required" form:"id"`
}
