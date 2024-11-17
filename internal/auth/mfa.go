package auth

import (
	"errors"
	"github.com/VEDA95/OpenBoard-API/internal/db"
	"github.com/VEDA95/OpenBoard-API/internal/settings"
	"github.com/VEDA95/OpenBoard-API/internal/util"
	"github.com/doug-martin/goqu/v9"
	"sync"
	"time"
)

type MultiAuthMethodInterface interface {
	CreateAuthChallenge(challengeType string, payload util.Payload) (util.Payload, error)
	VerifyAuthChallenge(challengeType string, payload util.Payload) (*ProviderAuthResult, error)
}

type MultiAuthMethodMap map[string]*MultiAuthMethodInterface

type MultiAuthMethodStore struct {
	mutex       *sync.RWMutex
	authMethods MultiAuthMethodMap
}

type MultiAuthMethod struct {
	Id             string                 `db:"id"`
	DateCreated    time.Time              `db:"date_created"`
	DateUpdated    time.Time              `db:"date_updated"`
	User           *User                  `db:"user,omitempty"`
	Name           string                 `db:"name"`
	Type           string                 `db:"type"`
	CredentialData map[string]interface{} `db:"credential_data,omitempty"`
}

type MultiAuthChallenge struct {
	Id          string                 `db:"id"`
	DateCreated time.Time              `db:"date_created"`
	DateUpdated time.Time              `db:"date_updated"`
	ExpiresOn   time.Time              `db:"expires_on"`
	User        User                   `db:"user"`
	AuthMethod  *MultiAuthMethod       `db:"auth_method,omitempty"`
	Data        map[string]interface{} `db:"data,omitempty"`
}

var MultiAuthMethods *MultiAuthMethodStore

func NewMultiAuthMethodStore() *MultiAuthMethodStore {
	return &MultiAuthMethodStore{
		mutex:       &sync.RWMutex{},
		authMethods: make(map[string]*MultiAuthMethodInterface),
	}
}

func GetMultiAuthMethodInstance(name string) MultiAuthMethodInterface {
	switch name {
	case "otp":
		return nil

	case "authenticator":
		return nil

	case "webauthn":
		webAuthn, err := NewWebAuthnMultiAuth()

		if err != nil {
			panic(err)
		}

		return webAuthn

	default:
		break
	}

	return nil
}

func UpdateMultiAuthChallenge(payload util.Payload) error {
	token, ok := payload["token"].(string)

	if !ok {
		return errors.New("token is missing")
	}

	updatePayload := goqu.Record{}
	expiresOn, okOne := payload["expires_on"].(time.Time)
	authMethodId, okTwo := payload["auth_method_id"].(string)
	data, okThree := payload["data"].(map[string]interface{})

	if okOne {
		updatePayload["expires_on"] = expiresOn
	}

	if okTwo {
		updatePayload["auth_method_id"] = authMethodId
	}

	if okThree {
		updatePayload["data"] = data
	}

	if len(updatePayload) > 0 {
		updatePayload["date_updated"] = time.Now()
		updateQuery := db.Instance.From("open_board_multi_auth_challenge").Prepared(true).
			Update().
			Where(goqu.Ex{"token": token}).
			Set(updatePayload).
			Executor()

		if _, err := updateQuery.Exec(); err != nil {
			return err
		}
	}

	return nil
}

func InitializeMultiAuthMethodStore() {
	settingsInterface := *settings.Instance.GetSettings("auth")
	authSettings := settingsInterface.(*settings.AuthSettings)
	MultiAuthMethods = NewMultiAuthMethodStore()

	if !authSettings.MultiAuthEnabled {
		return
	}

	availableMethods := make([]string, 0)

	if authSettings.OTPEnabled {
		availableMethods = append(availableMethods, "otp")
	}

	if authSettings.AuthenticatorEnabled {
		availableMethods = append(availableMethods, "authenticator")
	}

	if authSettings.WebAuthnEnabled {
		availableMethods = append(availableMethods, "webauthn")
	}

	if len(availableMethods) == 0 {
		return
	}

	MultiAuthMethods.RegisterAuthMethods(availableMethods)
}

func (multiAuthStore *MultiAuthMethodStore) RegisterAuthMethods(availableMethods []string) {
	multiAuthStore.mutex.Lock()
	defer multiAuthStore.mutex.Unlock()

	if len(multiAuthStore.authMethods) > 0 {
		for index, _ := range multiAuthStore.authMethods {
			delete(multiAuthStore.authMethods, index)
		}
	}

	for _, method := range availableMethods {
		multiAuthMethod := GetMultiAuthMethodInstance(method)

		multiAuthStore.authMethods[method] = &multiAuthMethod
	}
}

func (multiAuthStore *MultiAuthMethodStore) GetAllMultiAuthMethods() *MultiAuthMethodMap {
	return &multiAuthStore.authMethods
}

func (multiAuthStore *MultiAuthMethodStore) GetMultiAuthMethod(method string) *MultiAuthMethodInterface {
	multiAuthStore.mutex.RLock()
	authMethod, ok := multiAuthStore.authMethods[method]
	multiAuthStore.mutex.RUnlock()

	if !ok {
		return nil
	}

	return authMethod
}

func (multiAuthStore *MultiAuthMethodStore) SetMultiAuthMethod(name string) {
	multiAuthMethod := GetMultiAuthMethodInstance(name)

	if multiAuthMethod == nil {
		return
	}

	multiAuthStore.mutex.Lock()
	multiAuthStore.authMethods[name] = &multiAuthMethod
	multiAuthStore.mutex.Unlock()
}

func (multiAuthStore *MultiAuthMethodStore) RemoveMultiAuthMethod(method string) {
	multiAuthStore.mutex.Lock()
	defer multiAuthStore.mutex.Unlock()

	_, ok := multiAuthStore.authMethods[method]

	if ok {
		return
	}

	delete(multiAuthStore.authMethods, method)
}
