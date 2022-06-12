package main

import (
	controller "dumbways/controller"
	middlewares "dumbways/middleware"

	"github.com/gofiber/fiber"
	"github.com/gofiber/fiber/middleware"
)

func main() {
	app := fiber.New()
	app.Use(middleware.Logger())

	app.Post("/login", controller.Login)

	app.Get("/consume", controller.ConsumeWallet)
	app.Get("/consume-topup", controller.ConsumeTopUp)
	// Using Auth
	app.Post("/topup", middlewares.AuthRequired(), controller.Topup)
	app.Post("/create",middlewares.AuthRequired(), controller.Subscribe)
	app.Get("/wallet", middlewares.AuthRequired(), controller.Wallet)


	err := app.Listen(3000)
	if err != nil {
		panic(err)
	}
}
