package main

import (
	"github.com/joho/godotenv"
	"vepa/routes"
	"vepa/util"
)

func main() {
	godotenv.Load()
	util.InitLogger()
	routes.Routes()
}

