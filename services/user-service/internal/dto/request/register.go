package request

type RegisterRequest struct {
	Email    string `binding:"required,email"`
	Password string `binding:"required,min=8"`
	FullName string `binding:"required"`
	Phone    string `binding:"required"`
}
