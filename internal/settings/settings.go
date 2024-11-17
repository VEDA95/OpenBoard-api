package settings

import (
	"log"
	"sync"
)

type SettingsInterface interface {
	Load() error
	Save() error
}

type Settings struct {
	mutex    *sync.RWMutex
	settings map[string]*SettingsInterface
}

var Instance *Settings

func NewSettings() *Settings {
	return &Settings{
		mutex:    &sync.RWMutex{},
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
	default:
		break
	}

	return nil
}

func (settingsInstance *Settings) RegisterSettings(names []string) {
	settingsInstance.mutex.Lock()
	defer settingsInstance.mutex.Unlock()

	if len(settingsInstance.settings) > 0 {
		for index := range settingsInstance.settings {
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
	settingsInstance.mutex.RLock()
	settingInstance, ok := settingsInstance.settings[name]
	settingsInstance.mutex.RUnlock()

	if !ok {
		return nil
	}

	return settingInstance
}

func (settingsInstance *Settings) UpdateSettings(name string, settingInstance SettingsInterface) {
	settingsInstance.mutex.Lock()
	defer settingsInstance.mutex.Unlock()

	if _, ok := settingsInstance.settings[name]; !ok {
		return
	}

	settingsInstance.settings[name] = &settingInstance
}
