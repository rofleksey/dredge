package routes

import (
	"dredge/app/controller"
	"dredge/pkg/middleware"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func WSRoutes(app *fiber.App, wsController *controller.WS) {
	app.Get("/ws", middleware.WebSocketUpgrade(), websocket.New(wsController.Handle))
}
