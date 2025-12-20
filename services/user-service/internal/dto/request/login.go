package request

type LoginRequest struct {
	Email    string `binding:"required,email"`
	Password string `binding:"required"`
}
