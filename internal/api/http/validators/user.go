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

type PasswordResetUnlockValidator struct {
	Password   string `json:"password" validate:"required,min=8"`
	ReturnType string `json:"return_type,omitempty" validate:"omitempty,oneof=token session"`
}

type PasswordResetValidator struct {
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,min=8,eqfield=Password"`
	Token           string `json:"token,omitempty" validate:"min=8,max=32"`
}

type UserIdParamValidator struct {
	Id     string `json:"id" validate:"required,uuid4"`
	RoleId string `validate:"omitempty,uuid4"`
}

type UserRolesUpdateValidator struct {
	RoleIds []string `json:"ids" validate:"required,uuid4"`
}
