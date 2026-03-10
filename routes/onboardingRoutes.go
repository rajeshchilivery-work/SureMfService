package routes

import (
	"SureMFService/controller"
	"SureMFService/middleware"

	"github.com/gin-gonic/gin"
)

func onboardingRoutes(rg *gin.RouterGroup) {
	onboarding := rg.Group("/onboarding")
	onboarding.Use(middleware.AuthMiddleware())
	{
		onboarding.GET("/status", controller.GetOnboardingStatus)
		onboarding.GET("/pre-verification/:fp_id", controller.GetPreVerificationStatus)
		onboarding.GET("/kyc-check", controller.KYCCheck)
		onboarding.POST("/investor-profile", controller.CreateInvestorProfile)
		onboarding.POST("/phone", controller.AddPhone)
		onboarding.POST("/email", controller.AddEmail)
		onboarding.POST("/address", controller.AddAddress)
		onboarding.POST("/bank", controller.AddBankAccount)
		onboarding.POST("/nominee", controller.AddNominee)
		onboarding.POST("/activate", controller.ActivateAccount)
	}
}
