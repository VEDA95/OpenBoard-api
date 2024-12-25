package routes

import (
	"errors"
	"fmt"
	"github.com/VEDA95/OpenBoard-API/internal/api/http/validators"
	"github.com/VEDA95/OpenBoard-API/internal/settings"
	"github.com/VEDA95/OpenBoard-API/internal/util"
	"github.com/gofiber/fiber/v2"
)

func SettingsGET(context *fiber.Ctx) error {
	settingsName := context.Params("name")
	validationParamData := validators.SettingsParamsValidator{Name: settingsName}

	if errs := validators.Instance.Validate(&validationParamData); len(errs) > 0 {
		return util.CreateValidationError(errs)
	}

	settingsInterface := *settings.Instance.GetSettings(validationParamData.Name)

	if settingsInterface == nil {
		return errors.New(fmt.Sprintf("settings: %s could not be found", validationParamData.Name))
	}

	return util.JSONResponse(context, fiber.StatusOK, settingsInterface)
}

func SettingsPUT(context *fiber.Ctx) error {
	settingsName := context.Params("name")
	validationParamData := validators.SettingsParamsValidator{Name: settingsName}

	if errs := validators.Instance.Validate(&validationParamData); len(errs) > 0 {
		return util.CreateValidationError(errs)
	}

	settingsInterface := *settings.Instance.GetSettings(validationParamData.Name)

	if settingsInterface == nil {
		return errors.New(fmt.Sprintf("settings: %s could not be found", validationParamData.Name))
	}

	var output interface{}

	switch settingsInterface.(type) {
	case *settings.AuthSettings:
		validationData := new(settings.AuthSettings)

		if err := context.BodyParser(&validationData); err != nil {
			return err
		}

		if errs := validators.Instance.Validate(&validationData); len(errs) > 0 {
			return util.CreateValidationError(errs)
		}

		authSettings := settingsInterface.(*settings.AuthSettings)

		if validationData.AccessTokenLifetime != authSettings.AccessTokenLifetime {
			authSettings.AccessTokenLifetime = validationData.AccessTokenLifetime
		}

		if validationData.RefreshTokenLifetime != authSettings.RefreshTokenLifetime {
			authSettings.RefreshTokenLifetime = validationData.RefreshTokenLifetime
		}

		if validationData.RefreshTokenIdleLifetime != authSettings.RefreshTokenIdleLifetime {
			authSettings.RefreshTokenIdleLifetime = validationData.RefreshTokenIdleLifetime
		}

		if validationData.OTPEnabled != authSettings.OTPEnabled {
			authSettings.OTPEnabled = validationData.OTPEnabled
		}

		if validationData.ForceMultiAuthEnabled != authSettings.ForceMultiAuthEnabled {
			authSettings.ForceMultiAuthEnabled = validationData.ForceMultiAuthEnabled
		}

		if validationData.OTPEnabled != authSettings.OTPEnabled {
			authSettings.OTPEnabled = validationData.OTPEnabled
		}

		if validationData.WebAuthnEnabled != authSettings.WebAuthnEnabled {
			authSettings.WebAuthnEnabled = validationData.WebAuthnEnabled
		}

		if validationData.AuthenticatorEnabled != authSettings.AuthenticatorEnabled {
			authSettings.AuthenticatorEnabled = validationData.AuthenticatorEnabled
		}

		err := authSettings.Save()

		if err != nil {
			return err
		}

		settings.Instance.UpdateSettings(validationParamData.Name, authSettings)

		output = authSettings
		break

	case *settings.NotificationSettings:
		validationData := new(settings.NotificationSettings)

		if err := context.BodyParser(&validationData); err != nil {
			return err
		}

		if errs := validators.Instance.Validate(&validationData); len(errs) > 0 {
			return util.CreateValidationError(errs)
		}

		notificationSettings := settingsInterface.(*settings.NotificationSettings)

		if validationData.SMTPServer != notificationSettings.SMTPServer {
			notificationSettings.SMTPServer = validationData.SMTPServer
		}

		if validationData.SMTPPort != notificationSettings.SMTPPort {
			notificationSettings.SMTPPort = validationData.SMTPPort
		}

		if validationData.SMTPUser != notificationSettings.SMTPUser {
			notificationSettings.SMTPUser = validationData.SMTPUser
		}

		if validationData.SMTPPassword != notificationSettings.SMTPPassword {
			notificationSettings.SMTPPassword = validationData.SMTPPassword
		}

		if validationData.Name != notificationSettings.Name {
			notificationSettings.Name = validationData.Name
		}

		if validationData.EmailAddress != notificationSettings.EmailAddress {
			notificationSettings.EmailAddress = validationData.EmailAddress
		}

		if validationData.UseTLS != notificationSettings.UseTLS {
			notificationSettings.UseTLS = validationData.UseTLS
		}

		err := notificationSettings.Save()

		if err != nil {
			return err
		}

		settings.Instance.UpdateSettings(validationParamData.Name, notificationSettings)

		output = notificationSettings
		break

	}

	return util.JSONResponse(context, fiber.StatusOK, output)
}
