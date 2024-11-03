package main

import (
	"github.com/gabrielricci/otp-manager/api/controllers"
	"github.com/gabrielricci/otp-manager/internal/db"
	"github.com/gabrielricci/otp-manager/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func start() {
	godotenv.Load()
	db.Start()
	services.Repo = &db.OTPBadgerRepository{}
}

func stop() {
	db.Stop()
}

func startHealthCheckRoutes(r *gin.Engine) {
	controller := controllers.NewHealthCheckController()

	r.GET("/health-check", controller.GetHealthCheck)
}

func startOTPRoutes(r *gin.Engine) {
	controller := controllers.NewOTPController(
		services.NewOTPService(&db.OTPBadgerRepository{}),
	)

	group := r.Group("/account/")
	{
		group.POST("/:account-name", controller.CreateOTPAccount)
		group.POST("/:account-name/validate/:code", controller.ValidateCode)
	}
}

func main() {
	start()

	r := gin.Default()
	startHealthCheckRoutes(r)
	startOTPRoutes(r)
	r.Run()

	stop()
}
