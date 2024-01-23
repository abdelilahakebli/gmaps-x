package main

import "github.com/gofiber/fiber/v2"

func FetchData(c *fiber.Ctx) error {
	return c.SendString("Hello, World 👋!")
}
