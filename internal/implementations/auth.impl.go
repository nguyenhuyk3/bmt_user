package implementations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"
	"user_service/db/sqlc"
	"user_service/dto/request"
	"user_service/dto/response"
	"user_service/global"
	"user_service/internal/services"
	"user_service/utils/cryptor"
	"user_service/utils/generator"
	"user_service/utils/redis"
	mail "user_service/utils/sender"
	"user_service/utils/token/jwt"

	"github.com/jackc/pgx/v5/pgtype"
)

type authService struct {
	SqlStore *sqlc.SqlStore
	JwtMaker jwt.IMaker
}

func NewAuthService(sqlStore *sqlc.SqlStore, jwtMaker jwt.IMaker) services.IAuth {
	return &authService{
		SqlStore: sqlStore,
		JwtMaker: jwtMaker,
	}
}

// ForgotPassword implements services.IAuthUser.
func (a *authService) ForgotPassword() {
	panic("unimplemented")
}

// Register implements services.IAuthUser.
func (a *authService) SendOtp(ctx context.Context, arg request.SendOtpReq) (int, error) {
	// Check if email has otp in redis or not
	encryptedEmail, _ := cryptor.BcryptHashInput(arg.Email)
	key := fmt.Sprintf("%s%s", global.OTP_KEY, encryptedEmail)
	isExists := redis.ExistsKey(key)
	if isExists {
		return http.StatusConflict, errors.New("email is in registration status")
	}
	// Check if email already exists or not
	isExists, err := a.SqlStore.Queries.CheckAccountExistsByEmail(ctx, arg.Email)
	if isExists && err == nil {
		return http.StatusConflict, errors.New("email already exists")
	}

	expirationTime := int64(10)
	otp, _ := generator.GenerateNumberBasedOnLength(6)
	// Save email and otp is in registration status
	_ = redis.Save(key, request.VerifyOtpReq{
		EncryptedEmail: encryptedEmail,
		Otp:            otp,
	}, expirationTime)

	// Send mail
	err = mail.SendTemplateEmailOtp([]string{arg.Email},
		global.Config.Server.FromEmail, "otp_email.html",
		map[string]interface{}{
			"otp":             otp,
			"from_email":      global.Config.Server.FromEmail,
			"expiration_time": expirationTime,
		})
	if err != nil {
		redis.Delete(key)

		return http.StatusInternalServerError, errors.New("failed to send mail, please try again later")
	}

	return http.StatusOK, nil
}

// VerifyOTP implements services.IAuthUser.
func (a *authService) VerifyOtp(ctx context.Context, arg request.VerifyOtpReq) (int, error) {
	key := fmt.Sprintf("%s%s", global.OTP_KEY, arg.EncryptedEmail)

	var result request.VerifyOtpReq

	err := redis.Get(key, &result)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	isMatch := cryptor.BcryptCheckInput(arg.Email, result.EncryptedEmail)

	if isMatch == nil && arg.Otp == result.Otp {
		_ = redis.Delete(key)
		_ = redis.Save(fmt.Sprintf("%s%s", global.COMPLETE_REGISTRATION_PROCESS, arg.EncryptedEmail),
			map[string]interface{}{
				"encrypted_email": arg.EncryptedEmail,
			}, 10)

		return http.StatusOK, nil
	}

	return http.StatusUnauthorized, fmt.Errorf("invalid email or otp")
}

// CompleteRegister implements services.IAuth.
func (a *authService) CompleteRegistration(ctx context.Context, arg request.CompleteRegistrationReq) (int, error) {
	key := fmt.Sprintf("%s%s", global.COMPLETE_REGISTRATION_PROCESS, arg.EncryptedEmail)
	fmt.Println(arg.EncryptedEmail)
	isExists := redis.ExistsKey(key)
	if !isExists {
		return http.StatusNotFound, errors.New("email is not found in redis")
	}

	isMatch := cryptor.BcryptCheckInput(arg.Account.Email, arg.EncryptedEmail)
	if isMatch != nil {
		return http.StatusBadRequest, errors.New("encrypted email and email don't match")
	}

	err := a.SqlStore.InsertAccountTran(ctx, arg)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to complete registration: %w", err)
	}

	_ = redis.Delete(key)

	return http.StatusCreated, nil
}

// UpdatePassword implements services.IAuthUser.
func (a *authService) UpdatePassword() {
	panic("unimplemented")
}

// Login implements services.IAuthUser.
func (a *authService) Login(ctx context.Context, arg request.LoginReq) (response.LoginRes, int, error) {
	var result response.LoginRes

	user, err := a.SqlStore.Queries.GetUserByEmail(ctx, arg.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return result, http.StatusNotFound, errors.New("user not found")
		}

		return result, http.StatusInternalServerError, fmt.Errorf("failed to fetch user: %w", err)
	}

	isMatch := cryptor.BcryptCheckInput(user.Password, arg.Password)
	if isMatch != nil {
		return result, http.StatusUnauthorized, errors.New("invalid credentials")
	}

	accessToken, payload, err := a.JwtMaker.CreateAccessToken(user.Email, string(user.Role.Roles), time.Duration(60)*time.Minute)
	if err != nil {
		return result, http.StatusInternalServerError, fmt.Errorf("failed to create access token: %w", err)
	}

	refreshToken, err := a.JwtMaker.CreateRefreshToken(user.Email, time.Duration(60)*time.Minute)
	if err != nil {
		return result, http.StatusInternalServerError, fmt.Errorf("failed to create refresh token: %w", err)
	}

	_, err = a.SqlStore.Queries.UpdateAction(ctx, sqlc.UpdateActionParams{
		Column3: user.Email,

		Column1: pgtype.Timestamptz{
			Time:  time.Now(), // Thêm UTC() để đảm bảo đúng định dạng timestamptz
			Valid: true,
		},
		Column2: pgtype.Timestamptz{
			Valid: false,
			Time:  time.Time{}, // Thêm giá trị zero time
		},
	})
	if err != nil {
		return result, http.StatusInternalServerError, fmt.Errorf("failed to update user_action: %w", err)
	}

	result.AccessToken = accessToken
	result.RefreshToken = refreshToken
	result.Payload = payload

	return result, http.StatusOK, nil
}
