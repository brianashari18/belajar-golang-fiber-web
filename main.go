package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"time"
)

func main() {
	app := fiber.New(fiber.Config{
		IdleTimeout:  time.Minute * 5,
		ReadTimeout:  time.Minute * 5,
		WriteTimeout: time.Minute * 5,
		Prefork:      true,
	})

	app.Use("/api", func(ctx *fiber.Ctx) error {
		fmt.Println("I'm a middleware before process")
		err := ctx.Next()
		fmt.Println("I'm a middleware after process")
		return err
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello World")
	})

	if fiber.IsChild() {
		fmt.Println("I'm child process")
	} else {
		fmt.Println("I'm parent process")
	}

	err := app.Listen("localhost:8080")
	if err != nil {
		panic(err)
	}
}
