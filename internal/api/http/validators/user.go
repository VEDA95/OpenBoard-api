package validators

type UserCreate struct {
	Username        string  `json:"username" validate:"required,min=3,max=32"`
	Email           string  `json:"email" validate:"required,email"`
	Password        string  `json:"password" validate:"required,min=8"`
	ConfirmPassword string  `json:"confirm_password" validate:"required,min=8,eqfield=Password"`
	FirstName       *string `json:"first_name,omitempty"`
	LastName        *string `json:"last_name,omitempty"`
}

type UserUpdate struct {
	Username  *string `json:"username,omitempty" validate:"min=3,max=32"`
	Email     *string `json:"email,omitempty" validate:"email"`
	Thumbnail *string `json:"thumbnail,omitempty" validate:"uuid"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	DarkMode  *bool   `json:"dark_mode,omitempty" default:"false"`
}

type PasswordConfirmationPrompt struct {
	Password string `json:"password" validate:"required,min=8"`
}

type PasswordUpdate struct {
	ResetToken      string `json:"reset_token" validate:"required,min=1"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,min=8,eqfield=Password"`
}
