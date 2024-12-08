package validators

type LoginValidator struct {
	Username   string `json:"username" validate:"required,min=1,max=32"`
	Password   string `json:"password" validate:"required,min=8,max=32"`
	RememberMe bool   `json:"remember_me,omitempty" validate:"omitempty,oneof=true false"`
	ReturnType string `json:"type" validate:"required,min=1,max=7,oneof=token session"`
}

type MFASelectValidator struct {
	MFAType    string `json:"type,omitempty" validate:"omitempty,min=1,max=7,oneof=otp authenticator webauthn"`
	Skip       bool   `json:"skip,omitempty" validate:"omitempty,oneof=true false" default:"false"`
	ReturnType string `json:"return_type,omitempty" validate:"omitempty,min=1,max=7,oneof=token session"`
}

type OTPValidator struct {
	Otp string `json:"otp" validate:"required,min=6,max=6"`
}

type AuthenticatorValidator struct {
	Passcode string `json:"passcode" validate:"required,min=6,max=32"`
}

type LogoutValidator struct {
	ReturnType string `json:"return_type,omitempty" validate:"omitempty,min=1,max=7,oneof=token session"`
}
