package auth

import (
	"github.com/VEDA95/OpenBoard-API/internal/util"
	"sync"
)

type ProviderInterface interface {
	Login(payload util.Payload) (*ProviderAuthResult, error)
	Logout(payload util.Payload) error
	GetUser(payload util.Payload) (*User, error)
	GetUsers(payload util.Payload) (*[]User, error)
}

type ProviderAuthResult struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Validator    string `json:"validator,omitempty"`
	UserId       string `json:"user_id,omitempty"`
	MFARequired  bool   `json:"-"`
}

type ProviderRegistrationEntry struct {
	Config interface{}
	Key    string
	Name   string
}

type ProvidersInstance struct {
	mutex     *sync.RWMutex
	providers map[string]*ProviderInterface
}

var Instance *ProvidersInstance

func GetProviderInstance(driver string, config interface{}) ProviderInterface {
	switch driver {
	case "local":
		return &LocalAuthProvider{Config: config}
	default:
		break
	}
	return nil
}

func NewProvidersInstance() *ProvidersInstance {
	return &ProvidersInstance{
		mutex:     &sync.RWMutex{},
		providers: make(map[string]*ProviderInterface),
	}
}

func InitializeProvidersInstance(providerEntries []ProviderRegistrationEntry) {
	Instance = NewProvidersInstance()

	Instance.RegisterProviders(providerEntries)
}

func (providersInstance *ProvidersInstance) RegisterProviders(providerEntries []ProviderRegistrationEntry) {
	providersInstance.mutex.Lock()
	defer providersInstance.mutex.Unlock()

	if len(providersInstance.providers) > 0 {
		for index := range providersInstance.providers {
			delete(providersInstance.providers, index)
		}
	}

	for _, providerEntry := range providerEntries {
		provider := GetProviderInstance(providerEntry.Name, providerEntry.Config)
		providersInstance.providers[providerEntry.Key] = &provider
	}
}

func (providersInstance *ProvidersInstance) GetProvider(key string) *ProviderInterface {
	providersInstance.mutex.RLock()
	provider, ok := providersInstance.providers[key]
	providersInstance.mutex.RUnlock()

	if !ok {
		return nil
	}

	return provider
}

func (providersInstance *ProvidersInstance) AddProvider(entry ProviderRegistrationEntry) {
	provider := GetProviderInstance(entry.Name, entry.Config)

	providersInstance.mutex.Lock()
	defer providersInstance.mutex.Unlock()

	providersInstance.providers[entry.Key] = &provider
}

func (providersInstance *ProvidersInstance) RemoveProvider(name string) {
	providersInstance.mutex.Lock()
	defer providersInstance.mutex.Unlock()

	if _, ok := providersInstance.providers[name]; ok {
		delete(providersInstance.providers, name)
	}
}
