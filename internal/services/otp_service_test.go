package services

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/gabrielricci/otp-manager/api"
	"github.com/pquerna/otp/totp"
)

type MockOTPRepository struct {
	MockSaveSecret func(accountName string, secret string) error
	MockGetSecret  func(accountName string) (string, error)
}

func (m *MockOTPRepository) SaveSecret(accountName string, secret string) error {
	return m.MockSaveSecret(accountName, secret)
}
func (m *MockOTPRepository) GetSecret(accountName string) (string, error) {
	return m.MockGetSecret(accountName)
}

func setup() *OTPService {
	os.Setenv("OTP_ISSUER", "Test")

	mockRepo := &MockOTPRepository{
		MockSaveSecret: func(accountName string, secret string) error {
			if accountName == "valid" {
				return nil
			}
			return errors.New("SIMULATED")
		},
		MockGetSecret: func(accountName string) (secret string, err error) {
			if accountName == "valid" {
				return "secret", nil
			}
			return "", errors.New("ACCOUNT_NOT_FOUND")
		},
	}

	return NewOTPService(mockRepo)
}

func tearDown() {
	os.Setenv("OTP_ISSUER", "")
}

func TestReturnNilAfterSuccessfullySavingSecret(t *testing.T) {
	// arrange
	service := setup()

	// act
	_, err := service.CreateOTPSecret(&api.OTPUser{AccountName: "valid"})

	// assert
	if err != nil {
		t.Errorf("expected nil, got %s", err)
	}

	tearDown()
}

func TestReturnsKeyCorrectlyWhenSuccessfullyCreatingOTP(t *testing.T) {
	// arrange
	service := setup()

	// act
	key, err := service.CreateOTPSecret(&api.OTPUser{AccountName: "valid"})

	// assert
	if err != nil {
		t.Errorf("expected nil, got %s", err)
	}

	if key.Issuer() != "Test" {
		t.Errorf("expected 'Test', got %s", key.Issuer())
	}

	if key.AccountName() != "valid" {
		t.Errorf("expected 'valid', got %s", key.AccountName())
	}

	tearDown()
}

func TestValidatesCodeCorrectly(t *testing.T) {
	// arrange
	service := setup()

	// act
	key, err := service.CreateOTPSecret(&api.OTPUser{AccountName: "valid"})
	if err != nil {
		t.Errorf("creating key: expected nil, got %s", err)
		return
	}

	code, err := totp.GenerateCode(key.Secret(), time.Now())
	if err != nil {
		t.Errorf("generating code: expected nil, got %s", err)
		return
	}

	correct_code_validated := service.ValidateOTPCode(key.Secret(), code)
	wrong_code_validated := service.ValidateOTPCode(key.Secret(), "111111")

	if !correct_code_validated || wrong_code_validated {
		t.Errorf("validating code: expected true, got false")
	}

	tearDown()
}

func TestReturnErrAfterFailSavingSecret(t *testing.T) {
	// arrange
	service := setup()

	// act
	_, err := service.CreateOTPSecret(&api.OTPUser{AccountName: "blabla"})

	// assert
	if err.Error() != "SIMULATED" {
		t.Errorf("expected 'SIMULATED', got %s", err)
	}

	tearDown()
}

func TestReturnsSecretCorrectly(t *testing.T) {
	// arrange
	service := setup()

	// act
	secret, err := service.GetOTPSecret(&api.OTPUser{AccountName: "valid"})

	// assert
	if err != nil {
		t.Errorf("expected nil, got %s", err)
	}

	if secret != "secret" {
		t.Errorf("expected 'secret', got %s", secret)
	}

	tearDown()
}

func TestReturnsErrorWhenSecretNotFound(t *testing.T) {
	// arrange
	service := setup()

	// act
	secret, err := service.GetOTPSecret(&api.OTPUser{AccountName: "blablabla"})

	// assert
	if err.Error() != "ACCOUNT_NOT_FOUND" {
		t.Errorf("expected 'ACCOUNT_NOT_FOUND', got %s", err)
	}

	if secret != "" {
		t.Errorf("expected '', got %s", secret)
	}

	tearDown()
}
