package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"github.com/VEDA95/OpenBoard-API/internal/util"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"strings"
)

type RefreshCredentials struct {
	Selector        string `json:"selector"`
	Validator       string `json:"validator"`
	HashedValidator string `json:"-"`
}

func CreateSessionToken() (string, error) {
	token, err := util.GenerateRandomBytes(32)

	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(token), nil
}

func CreateRefreshCredentials() (*RefreshCredentials, error) {
	credentialBytes, err := util.GenerateRandomBytes(32)

	if err != nil {
		return nil, err
	}

	selector, err := uuid.FromBytes(credentialBytes)

	if err != nil {
		return nil, err
	}

	hashedValidator := sha256.Sum256(credentialBytes)

	return &RefreshCredentials{
		Selector:        selector.String(),
		Validator:       base64.URLEncoding.EncodeToString(credentialBytes),
		HashedValidator: base64.URLEncoding.EncodeToString(hashedValidator[:]),
	}, nil
}

func ExtractRefreshCredential(token string) (*RefreshCredentials, error) {
	credentials := strings.Split(token, ":")

	if len(credentials) != 2 {
		return nil, errors.New("invalid token")
	}

	_, err := uuid.Parse(credentials[0])

	if err != nil {
		return nil, err
	}

	return &RefreshCredentials{Selector: credentials[0], Validator: credentials[1]}, nil
}

func CompareValidatorToken(validatorToken string, validatorTokenHash string) error {
	validatorBytes, err := base64.URLEncoding.DecodeString(validatorToken)

	if err != nil {
		return err
	}

	validatorHash := sha256.Sum256(validatorBytes)
	hashToken := base64.URLEncoding.EncodeToString(validatorHash[:])

	if hashToken != validatorTokenHash {
		return errors.New("validator is invalid")
	}

	return nil
}

func ExtractSessionToken(context *fiber.Ctx, cookieField string) (string, error) {
	tokenHeader := string(context.Request().Header.Peek("Authorization")[:])
	splitHeader := strings.Split(tokenHeader, " ")

	if len(cookieField) == 0 {
		cookieField = "open_board_session"
	}

	if len(splitHeader) != 2 {
		sessionCookie := context.Cookies(cookieField)

		if len(sessionCookie) == 0 {
			return "", errors.New("no session cookie found")
		}

		return sessionCookie, nil
	}

	return splitHeader[1], nil
}
