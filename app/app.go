package app

import (
	"github.com/joho/godotenv"
	"vepa/app/routes"
	"vepa/app/util"
)

func init() {
}

func Run() {

	godotenv.Load()
	util.InitLogger()
	routes.Routes()

}
