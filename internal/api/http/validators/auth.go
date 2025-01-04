package validators

type LoginValidator struct {
	Username   string `json:"username" validate:"required,min=1,max=32"`
	Password   string `json:"password" validate:"required,min=8,max=32"`
	RememberMe bool   `json:"remember_me,omitempty" validate:"omitempty,oneof=true false"`
	ReturnType string `json:"type" validate:"required,oneof=token session"`
}

type MFASelectValidator struct {
	MFAType    string `json:"type,omitempty" validate:"omitempty,min=1,max=7,oneof=otp authenticator webauthn"`
	Skip       bool   `json:"skip,omitempty" validate:"omitempty,oneof=true false" default:"false"`
	ReturnType string `json:"return_type,omitempty" validate:"omitempty,oneof=token session"`
}

type OTPValidator struct {
	Otp string `json:"otp" validate:"required,min=6,max=6"`
}

type AuthenticatorValidator struct {
	Passcode string `json:"passcode" validate:"required,min=6,max=32"`
}

type LogoutValidator struct {
	ReturnType string `json:"return_type,omitempty" validate:"omitempty,oneof=token session"`
}

type RoleIDValidator struct {
	Id string `validate:"required,uuid4"`
}

type RoleValidator struct {
	Name        string   `json:"name" validate:"required,min=1,max=32"`
	Permissions []string `json:"permissions,omitempty" validate:"omitempty,uuid4"`
}

type RoleUpdateValidator struct {
	Name        string   `json:"name,omitempty" validate:"omitempty,min=1,max=32"`
	Permissions []string `json:"permissions,omitempty" validate:"omitempty,uuid4"`
}

type PermissionValidator struct {
	Path string `json:"path" validate:"required,min=1,max=32"`
}

type PermissionUpdateValidator struct {
	Path string `json:"path,omitempty" validate:"omitempty,min=1,max=32"`
}

type UserRegisterValidator struct {
	Username        string  `json:"username" validate:"required,min=3,max=32"`
	Email           string  `json:"email" validate:"required,email"`
	Password        string  `json:"password" validate:"required,min=8"`
	ConfirmPassword string  `json:"confirm_password" validate:"required,min=8,eqfield=Password"`
	FirstName       *string `json:"first_name,omitempty"`
	LastName        *string `json:"last_name,omitempty"`
	ReturnType      string  `json:"type" validate:"required,oneof=token session"`
}
