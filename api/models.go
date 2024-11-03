package api

type OTPUser struct {
	AccountName string `uri:"account-name" binding:"required,email"`
}

type OneTimeCode struct {
	Code string `uri:"code" binding:"required"`
}
