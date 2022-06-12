package middlewares

import (
	"github.com/gofiber/fiber"
	jwtware "github.com/gofiber/jwt"
)

const jwtSecret = "secret"

func AuthRequired() func(ctx *fiber.Ctx) {
	return jwtware.New(jwtware.Config{
		ErrorHandler: func(ctx *fiber.Ctx, err error) {
			ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		},
		SigningKey: []byte(jwtSecret),
	})
}
