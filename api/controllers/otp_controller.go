package controllers

import (
	"bytes"
	"image/png"
	"net/http"

	"github.com/gabrielricci/otp-manager/api"
	"github.com/gabrielricci/otp-manager/internal/services"

	"github.com/gin-gonic/gin"
)

type OTPController struct {
	service *services.OTPService
}

func NewOTPController(service *services.OTPService) *OTPController {
	return &OTPController{
		service: service,
	}
}

func (controller *OTPController) CreateOTPAccount(c *gin.Context) {
	var user api.OTPUser
	if err := c.ShouldBindUri(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	_, err := controller.service.GetOTPSecret(&user)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "OTP already created",
		})
		return
	}

	key, err := controller.service.CreateOTPSecret(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	var buf bytes.Buffer
	img, err := key.Image(200, 200)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	png.Encode(&buf, img)

	c.Data(http.StatusOK, "image/png", buf.Bytes())
}

func (controller *OTPController) ValidateCode(c *gin.Context) {
	var user api.OTPUser
	if err := c.ShouldBindUri(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var code api.OneTimeCode
	if err := c.ShouldBindUri(&code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	secret, err := controller.service.GetOTPSecret(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Account not found",
		})
		return
	}

	valid := controller.service.ValidateOTPCode(secret, code.Code)
	if !valid {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Invalid code",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "validated",
	})
}
