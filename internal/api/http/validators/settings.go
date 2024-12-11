package validators

type SettingsParamsValidator struct {
	Name string `validate:"required,oneof=auth notification general"`
}
