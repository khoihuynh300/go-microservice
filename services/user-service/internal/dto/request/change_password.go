package request

type ChangePasswordRequest struct {
	CurrentPassword string
	NewPassword     string
}
