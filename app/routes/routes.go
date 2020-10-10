package routes

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"vepa/app/controllers"
	"vepa/app/util"
)

func Routes() {
	r := gin.Default()
	// Server static files
	r.Static("/assets", util.GoDotEnvVariable("APP_ASSETS"))

	// Load all templates
	r.LoadHTMLGlob(util.GoDotEnvVariable("APP_TEMPLATES"))

	//Routes
	dashboardRouter := r.Group("/dashboard")
	{
		dashboardRouter.GET("", controllers.GetDashboard)
		dashboardRouter.GET("/getAttendants", controllers.GetAttendants)
		dashboardRouter.GET("/getAddAttendant", controllers.GetAddAttendant)
		dashboardRouter.POST("/postAddAttendant", controllers.PostAddAttendant)


	}

	//Users
	r.POST("/register", controllers.RegisterHandler)
	r.POST("/login", controllers.LoginHandler)
	r.GET("/profile/:id", controllers.ProfileHandler)
	r.PUT("/editProfile", controllers.EditProfile)
	r.PUT("/token/:id", controllers.FCMTokenHandler)
	//Vehicles
	r.POST("/addVehicle", controllers.AddVehicleHandler)
	r.GET("/vehicle/:vehicleReg", controllers.GetVehicleHandler)
	r.PUT("/editVehicle/:id", controllers.EditVehicleHandler)
	r.GET("/userVehicles", controllers.UserVehiclesHandler)
	r.DELETE("/deleteVehicle/:id", controllers.DeleteVehicleHandler)
	r.GET("/isWaitingClamp", controllers.VehiclesWaitingClamp)
	r.GET("/isClamped", controllers.ClampedVehicle)
	r.GET("/isVehicleClamped/:vehicleReg", controllers.CheckIfVehicleIsClampedHandler)
	//Payment
	r.POST("/makePayment", controllers.PaymentHandler)
	r.POST("/rcb", controllers.CallBackHandler)
	r.GET("/userPayments", controllers.UserPaymentsHandler)
	r.GET("/fetchPaidDays/:vehicleReg", controllers.GetPaidDays)
	r.GET("/verifyPayment/:vehicleReg", controllers.VerificationHandler)
	r.GET("/unpaidVehicleHistory/vehicleReg", controllers.UnpaidVehicleHistoryHandler)
	r.GET("/clampVehicle/:vehicleReg", controllers.ClampVehicleHandler)
	r.POST("/clearclampfee/:id", controllers.ClearClampFeeHandler)
	r.POST("/clamprcb", controllers.ClampCallBackHandler)

	port := util.GetPort()
	util.Log("Starting app on port 👍 ✓ ⌛ :", port)
	log.Fatal(http.ListenAndServe(port, r))
}
