// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package injectors

import (
	"user_service/db/sqlc"
	"user_service/internal/controllers"
	"user_service/internal/implementations/customer"
	"user_service/internal/implementations/forgot_password"
	"user_service/internal/implementations/login"
	"user_service/internal/implementations/logout"
	"user_service/internal/implementations/oauth2"
	"user_service/internal/implementations/registration"
	"user_service/internal/injectors/provider"
	"user_service/internal/message_broker"
	"user_service/internal/middlewares"
	"user_service/utils/redis"
	"user_service/utils/token/jwt"
)

// Injectors from auth.controller.wire.go:

func InitAuthController() (*controllers.AuthController, error) {
	pool := provider.ProvidePgxPool()
	iStore := sqlc.NewStore(pool)
	iRedis := redis.NewRedisClient()
	iMessageBroker := messagebroker.NewKafkaMessageBroker()
	iRegistration := registration.NewRegistrationService(iStore, iRedis, iMessageBroker)
	string2 := provider.ProvideSecretKey()
	iMaker, err := jwt.NewJWTMaker(string2)
	if err != nil {
		return nil, err
	}
	iLogin := login.NewLoginService(iStore, iMaker)
	iForgotPassword := forgotpassword.NewForgotPasswordSevice(iStore, iRedis, iMessageBroker)
	ioAuth2 := oauth2.NewOAuth2Service(iStore, iMaker)
	iLogout := logout.NewLogoutService(iStore)
	googleOAuthConfig := provider.ProvideGoogleOAuthConfig()
	facebookOAuthConfig := provider.ProvideFacebookOAuthConfig()
	authController := controllers.NewAuthController(iRegistration, iLogin, iForgotPassword, ioAuth2, iLogout, googleOAuthConfig, facebookOAuthConfig)
	return authController, nil
}

// Injectors from auth.middleware.wire.go:

func InitAuthMiddleware() (*middlewares.AuthMiddleware, error) {
	string2 := provider.ProvideSecretKey()
	iMaker, err := jwt.NewJWTMaker(string2)
	if err != nil {
		return nil, err
	}
	iRedis := redis.NewRedisClient()
	authMiddleware := middlewares.NewAuthMiddleware(iMaker, iRedis)
	return authMiddleware, nil
}

// Injectors from customer.controller.wire.go:

func InitCustomerController() (*controllers.CustomerController, error) {
	pool := provider.ProvidePgxPool()
	iStore := sqlc.NewStore(pool)
	iCustomer := customer.NewCustomerService(iStore)
	customerController := controllers.NewCustomerController(iCustomer)
	return customerController, nil
}
