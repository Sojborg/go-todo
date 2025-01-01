package routes

import (
	"github.com/gofiber/fiber/v2"
	todoControllers "github.com/sojborg/go-todo/internal/controllers"
)

func RegisterRoutes(app *fiber.App) {
	app.Get("/todos", todoControllers.GetTodos)
	app.Post("/todos", todoControllers.CreateTodo)
	app.Patch("/todos/:id", todoControllers.UpdateTodo)
	app.Delete("/todos/:id", todoControllers.DeleteTodo)
}
