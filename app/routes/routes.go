package routes

import (
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"

	"log"
	"net/http"
	"vepa/app/controllers"
	"vepa/app/util"
)

func Routes() {
	r := gin.Default()
	// Session to use in auth
	r.Use(sessions.Sessions("vepa_session", sessions.NewCookieStore([]byte(util.GoDotEnvVariable("SESSION_KEY")))))

	// Server static files
	r.Static("/assets", util.GoDotEnvVariable("APP_ASSETS"))

	// Load all templates
	r.LoadHTMLGlob(util.GoDotEnvVariable("APP_TEMPLATES"))

	//Routes
	dashboardRouter := r.Group("/dashboard")
	dashboardRouter.Use(util.AuthRequired())

	{
		dashboardRouter.GET("", controllers.GetDashboard)
		dashboardRouter.GET("/getAttendants", controllers.GetAttendants)
		dashboardRouter.GET("/getCustomers", controllers.GetCustomers)
		dashboardRouter.GET("/getAddAttendant", controllers.GetAddAttendant)
		//dashboardRouter.GET("/getAdminLogin", controllers.GetAdminLoginHandler)
		dashboardRouter.POST("/postAddAttendant", controllers.PostAddAttendant)
		//dashboardRouter.POST("/adminRegistration", controllers.RegisterAdminHandler)
		//dashboardRouter.POST("/adminLogin", controllers.AdminLoginHandler)

	}
	r.POST("/auth/adminLogin", controllers.AdminLoginHandler)
	r.GET("/auth/getAdminLogin", controllers.GetAdminLoginHandler)
	r.GET("/logout", controllers.GetLogout)
	r.POST("/auth/adminRegistration", controllers.RegisterAdminHandler)

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
	r.GET("/nonRegisteredClamped", controllers.NonRegisteredClampedVehiclesHandler)
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
	//Attendant
	r.POST("/attendantLogin", controllers.AttendantLoginHandler)
	//FCM
	r.PUT("/attendantToken/:id", controllers.SaveAttendantsFCMHandler)

	port := util.GetPort()
	util.Log("Starting app on port 👍 ✓ ⌛ :", port)
	log.Fatal(http.ListenAndServe(port, r))
}
