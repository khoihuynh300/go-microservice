package events

type UserRegisteredEvent struct {
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Token    string `json:"token"`
}

type EmailVerifySuccessEvent struct {
	Email string `json:"email"`
}

type UserForgotPasswordEvent struct {
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Token    string `json:"token"`
}

type UserPasswordResetSuccessEvent struct {
	Email string `json:"email"`
}
