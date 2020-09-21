package routes

import (
	"log"
	"net/http"
	"vepa/controllers"
	"vepa/util"

	"github.com/gorilla/mux"
)

//Routes is...
func Routes() {
	util.Log("Initializing routes...")

	r := mux.NewRouter()
	r.HandleFunc("/register", controllers.RegisterHandler).Methods("POST")
	r.HandleFunc("/userLogin", controllers.LoginHandler).Methods("POST")
	r.HandleFunc("/editProfile", controllers.EditProfileHandler).Methods("PUT")
	r.HandleFunc("/addVehicle", controllers.AddVehicleHandler).Methods("POST")
	r.HandleFunc("/editVehicle/{id}", controllers.EditVehicleHandler).Methods("PUT")
	r.HandleFunc("/deleteVehicle/{id}", controllers.DeleteVehicleHandler).Methods("DELETE")
	r.HandleFunc("/makePayment", controllers.PaymentHandler).Methods("POST")
	r.HandleFunc("/rcb", controllers.CallBackHandler).Methods("POST")
	r.HandleFunc("/token/{id}", controllers.FCMTokenHandler).Methods("PUT")
	r.HandleFunc("/profile/{id}", controllers.ProfileHandler).Methods("GET")
	r.HandleFunc("/userVehicles", controllers.UserVehiclesHandler).Methods("GET")
	r.HandleFunc("/userPayments", controllers.UserPaymentsHandler).Methods("GET")
	r.HandleFunc("/fetchPaidDays/{vehicleReg}", controllers.GetPaidDays).Methods("GET")
	r.HandleFunc("/verifyPayment/{vehicleReg}", controllers.VerificationHandler).Methods("GET")
	r.HandleFunc("/unpaidVehicleHistory/{vehicleReg}", controllers.UnpaidVehicleHistoryHandler).Methods("GET")
	r.HandleFunc("/clampVehicle/{vehicleReg}", controllers.ClampVehicle).Methods("GET")
	r.HandleFunc("/saveFCM", controllers.SaveAttendantsFCM).Methods("PUT")
	r.HandleFunc("/isWaitingClamp", controllers.VehiclesWaitingClamp).Methods("GET")
	r.HandleFunc("/isClamped", controllers.ClampedVehicles).Methods("GET")
	r.HandleFunc("/isVehicleClamped/{vehicleReg}", controllers.CheckVehicleClamp).Methods("GET")

	port := util.GetPort()

	util.Log("Starting app on port 👍 ✓ ⌛ :", port)

	log.Fatal(http.ListenAndServe(port, r))
}
