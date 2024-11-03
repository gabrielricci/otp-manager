package controllers

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gabrielricci/otp-manager/api"
	"github.com/gabrielricci/otp-manager/internal/services"

	"github.com/gin-gonic/gin"
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

func setup() (router *gin.Engine, secret string) {
	os.Setenv("OTP_ISSUER", "Test")

	key, err := services.NewOTPService(nil).GenerateOTPSecret(&api.OTPUser{AccountName: "gabrielricci+existing@email.com"})
	if err == nil {
		secret = key.Secret()
	}

	localSecret := strings.Clone(secret)

	mockRepo := &MockOTPRepository{
		MockSaveSecret: func(accountName string, secret string) error {
			if accountName == "gabrielricci+new@gmail.com" {
				return nil
			}

			return errors.New("SIMULATED")
		},
		MockGetSecret: func(accountName string) (secret string, err error) {
			if accountName == "gabrielricci+existing@gmail.com" {
				return localSecret, nil
			}
			return "", errors.New("ACCOUNT_NOT_FOUND")
		},
	}

	service := services.NewOTPService(mockRepo)
	controller := NewOTPController(service)

	router = gin.Default()
	router.POST("/account/:account-name", controller.CreateOTPAccount)
	router.POST("/account/:account-name/validate/:code", controller.ValidateCode)

	return
}

func tearDown() {
	os.Setenv("OTP_ISSUER", "")
}

func Test200WhenCreateNewAccountSuccessfully(t *testing.T) {
	// arrange
	router, _ := setup()

	// act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/account/gabrielricci+new@gmail.com", nil)
	router.ServeHTTP(w, req)

	// assert
	if w.Code != 200 {
		t.Errorf("Expected '200', got %d", w.Code)
		t.Errorf("%s", w.Body.String())
	}

	tearDown()
}

func Test400WhenCreatingDuplicateAccount(t *testing.T) {
	// arrange
	router, _ := setup()

	// act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/account/gabrielricci+existing@gmail.com", nil)
	router.ServeHTTP(w, req)

	// assert
	if w.Code != 400 || !bytes.Contains(w.Body.Bytes(), []byte("already created")) {
		t.Errorf("Expected '400', got %d", w.Code)
		t.Errorf("Expected 'already created' inside JSON, got %s", w.Body.String())
	}

	tearDown()
}

func Test400WhenProvidingInvalidEmail(t *testing.T) {
	// arrange
	router, _ := setup()

	// act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/account/gabrielricci", nil)
	router.ServeHTTP(w, req)

	// assert
	if w.Code != 400 || !bytes.Contains(w.Body.Bytes(), []byte("email")) {
		t.Errorf("Expected '400', got %d", w.Code)
		t.Errorf("Expected 'email' inside JSON, got %s", w.Body.String())
	}

	tearDown()
}

func Test500WhenInternalErrorCreatingAccount(t *testing.T) {
	// arrange
	router, _ := setup()

	// act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/account/invalid@gmail.com", nil)
	router.ServeHTTP(w, req)

	// assert
	if w.Code != 500 || !bytes.Contains(w.Body.Bytes(), []byte("SIMULATED")) {
		t.Errorf("Expected '500', got %d", w.Code)
		t.Errorf("Expected 'SIMULATED' inside JSON, got %s", w.Body.String())
	}

	tearDown()
}

func Test404WhenValidatingCodeForNonExistentAccount(t *testing.T) {
	// arrange
	router, _ := setup()

	// act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/account/gabrielricci+new@gmail.com/validate/123123", nil)
	router.ServeHTTP(w, req)

	// assert
	if w.Code != 404 || !bytes.Contains(w.Body.Bytes(), []byte("Account not found")) {
		t.Errorf("Expected '404', got %d", w.Code)
		t.Errorf("Expected 'Account not found' inside JSON, got %s", w.Body.String())
	}

	tearDown()
}

func Test200WhenValidatingCorrectCode(t *testing.T) {
	// arrange
	router, secret := setup()
	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		t.Errorf("Error generating code for secret: %s", err)
	}

	var sb strings.Builder
	sb.WriteString("/account/gabrielricci+existing@gmail.com/validate/")
	sb.WriteString(code)
	url := sb.String()

	// act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", url, nil)
	router.ServeHTTP(w, req)

	// assert
	if w.Code != 200 || !bytes.Contains(w.Body.Bytes(), []byte("validated")) {
		t.Errorf("Expected '200', got %d", w.Code)
		t.Errorf("Expected 'validated' inside JSON, got %s", w.Body.String())
	}

	tearDown()
}

func Test403WhenValidatingCorrectCode(t *testing.T) {
	// arrange
	router, _ := setup()

	// act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/account/gabrielricci+existing@gmail.com/validate/123123", nil)
	router.ServeHTTP(w, req)

	// assert
	if w.Code != 403 || !bytes.Contains(w.Body.Bytes(), []byte("Invalid code")) {
		t.Errorf("Expected '403', got %d", w.Code)
		t.Errorf("Expected 'Invalid code' inside JSON, got %s", w.Body.String())
	}

	tearDown()
}
