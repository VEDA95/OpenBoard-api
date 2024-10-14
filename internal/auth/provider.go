package auth

type ProviderInterface interface {
	Login(payload ProviderPayload) (*ProviderAuthResult, error)
	Logout(payload ProviderPayload) error
	GetUser(payload ProviderPayload) (*User, error)
	GetUsers(payload ProviderPayload) (*[]User, error)
}

type ProviderPayload map[string]interface{}

type ProviderAuthResult struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	UserId       string `json:"user_id,omitempty"`
}

type ProviderRegistrationEntry struct {
	Config interface{}
	Key    string
	Name   string
}

type ProvidersInstance struct {
	providers map[string]*ProviderInterface
}

var Instance *ProvidersInstance

func GetProviderInstance(driver string, config interface{}) ProviderInterface {
	switch driver {
	case "local":
		return &LocalAuthProvider{Config: config}
	}
	return nil
}

func NewProvidersInstance() *ProvidersInstance {
	return &ProvidersInstance{
		providers: make(map[string]*ProviderInterface),
	}
}

func InitializeProvidersInstance(providerEntries []ProviderRegistrationEntry) {
	Instance = NewProvidersInstance()

	Instance.RegisterProviders(providerEntries)
}

func (providersInstance *ProvidersInstance) RegisterProviders(providerEntries []ProviderRegistrationEntry) {
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
	provider, ok := providersInstance.providers[key]

	if !ok {
		return nil
	}

	return provider
}

func (providersInstance *ProvidersInstance) RemoveProvider(name string) {
	if _, ok := providersInstance.providers[name]; ok {
		delete(providersInstance.providers, name)
	}
}
