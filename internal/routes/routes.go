package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/sojborg/go-todo/internal/controllers/authController"
	"github.com/sojborg/go-todo/internal/controllers/todoController"
)

func RegisterRoutes(r *chi.Mux) {
	r.Get("/api/todos", todoController.GetTodos)
	r.Post("/api/todos", todoController.CreateTodo)
	r.Patch("/api/todos/{id}", todoController.UpdateTodo)
	r.Delete("/api/todos/{id}", todoController.DeleteTodo)

	r.Get("/auth/{provider}", authController.Login)
	r.Get("/auth/{provider}/callback", authController.GetAuthCallbackFunction)
	r.Get("/auth/userinfo", authController.GetUserInfo)
}
