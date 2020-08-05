package routes

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"vepa/controllers"
	"vepa/util"
)

func Routes() {
	r := mux.NewRouter()
	r.HandleFunc("/register", controllers.RegisterHandler).Methods("POST")
	r.HandleFunc("/login", controllers.LoginHandler).Methods("POST")
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
	r.HandleFunc("/paymentPush", controllers.GetPushHandler).Methods("GET")

	log.Fatal(http.ListenAndServe(util.GetPort(), r))
}
