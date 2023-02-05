package main

import (
	"github.com/mixedmachine/user-auth-server/cmd/v1/api"
)

// @title User Auth API
// @description This is a user auth api server.
// @version 1.0.0
// @BasePath /
// @schemes http
// @host localhost:9090
// @contact.name MixedMachine
// @contact.url mixedmachine.dev
// @contact.email michael.martinez.dev@gmail.com
func main() {
	api.Init()
	api.RunUserAuthApiServer()
}
