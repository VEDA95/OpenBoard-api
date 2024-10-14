package settings

import "log"

type SettingsInterface interface {
	Load() error
	Save() error
}

type Settings struct {
	settings map[string]*SettingsInterface
}

var Instance *Settings

func NewSettings() *Settings {
	return &Settings{
		settings: make(map[string]*SettingsInterface),
	}
}

func InitializeSettingsInstance(names []string) {
	Instance = NewSettings()

	Instance.RegisterSettings(names)
}

func GetSettingsInstance(name string) SettingsInterface {
	switch name {
	case "auth":
		return &AuthSettings{}
	}

	return nil
}

func (settingsInstance *Settings) RegisterSettings(names []string) {
	if len(settingsInstance.settings) > 0 {
		for index, _ := range settingsInstance.settings {
			delete(settingsInstance.settings, index)
		}
	}

	for _, name := range names {
		settingInstance := GetSettingsInstance(name)

		if err := settingInstance.Load(); err != nil {
			log.Print(err.Error())
			continue
		}

		settingsInstance.settings[name] = &settingInstance
	}
}

func (settingsInstance *Settings) GetSettings(name string) *SettingsInterface {
	settingInstance, ok := settingsInstance.settings[name]

	if !ok {
		return nil
	}

	return settingInstance
}

func (settingsInstance *Settings) UpdateSettings(name string, settingInstance SettingsInterface) {
	if _, ok := settingsInstance.settings[name]; !ok {
		return
	}

	settingsInstance.settings[name] = &settingInstance
}
