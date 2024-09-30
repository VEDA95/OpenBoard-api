package validators

type LoginValidator struct {
	Username   string `json:"username" validate:"required,min=1,max=32"`
	Password   string `json:"password" validate:"required,min=8,max=32"`
	RememberMe bool   `json:"remember_me,omitempty" validate:"omitempty,oneof=true false"`
	ReturnType string `json:"type" validate:"required,min=1,max=7,oneof=token session"`
}
