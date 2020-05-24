package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"vepa/controller"
)

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/register", controller.RegisterHandler).Methods("POST")
	r.HandleFunc("/login", controller.LoginHandler).Methods("POST")
	r.HandleFunc("/addVehicle", controller.AddVehicleHandler).Methods("POST")
	r.HandleFunc("/makePayment", controller.PaymentHandler).Methods("POST")
	r.HandleFunc("/rcb", controller.CallBackHandler).Methods("POST")
	r.HandleFunc("/profile", controller.ProfileHandler).Methods("GET")
	r.HandleFunc("/userVehicles", controller.UserVehiclesHandler).Methods("GET")
	log.Fatal(http.ListenAndServe(GetPort(), r))
}

// GetPort from the environment so we can run on Heroku
func GetPort() string {
	var port = os.Getenv("PORT")
	// Set a default port if there is nothing in the environment
	if port == "" {
		port = "4000"
		fmt.Println("INFO: No PORT environment variable detected, defaulting to " + port)
	}
	return ":" + port
}
