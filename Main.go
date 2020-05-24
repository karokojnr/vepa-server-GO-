package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"vepa/controller"
)

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/register", controller.RegisterHandler).Methods("POST")
	r.HandleFunc("/login", controller.LoginHandler).Methods("POST")
	r.HandleFunc("/addVehicle", controller.AddVehicleHandler).Methods("POST")
	r.HandleFunc("/profile", controller.ProfileHandler).Methods("GET")
	r.HandleFunc("/userVehicles", controller.UserVehiclesHandler).Methods("GET")
	log.Fatal(http.ListenAndServe(":4000", r))
}
