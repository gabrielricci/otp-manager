package services

import (
	"github.com/gabrielricci/otp-manager/api"
	"github.com/gabrielricci/otp-manager/internal/db"
	"os"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type OTPService struct {
	repo db.OTPRepository
}

func NewOTPService(repo db.OTPRepository) *OTPService {
	return &OTPService{
		repo: repo,
	}
}

func (s *OTPService) CreateOTPSecret(otpUser *api.OTPUser) (key *otp.Key, err error) {
	key, err = s.GenerateOTPSecret(otpUser)
	if err != nil {
		return
	}

	err = s.repo.SaveSecret(otpUser.AccountName, key.Secret())
	return
}

func (s *OTPService) GenerateOTPSecret(otpUser *api.OTPUser) (*otp.Key, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      os.Getenv("OTP_ISSUER"),
		AccountName: otpUser.AccountName,
	})

	return key, err
}

func (s *OTPService) GetOTPSecret(otpUser *api.OTPUser) (string, error) {
	return s.repo.GetSecret(otpUser.AccountName)
}

func (s *OTPService) ValidateOTPCode(secret string, code string) bool {
	return totp.Validate(code, secret)
}
