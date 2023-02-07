package main

import (
	"errors"

	"github.com/ZhangTao1596/neo-go-util/application"
)

const (
	defaultUrl string = "http://localhost:20332"
)

func main() {
	app := application.NewApp("neo")
	app.RegisterCommand("print", "[msg]", "show error message", func(a *application.Context) error {
		return errors.New("test error")
	})
	app.Run()
}
