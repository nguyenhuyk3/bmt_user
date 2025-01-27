package initializations

import (
	"log"
	"user_service/internal/injectors"
	"user_service/internal/routers"

	"github.com/gin-gonic/gin"
)

func initRouter() *gin.Engine {
	r := gin.Default()
	// Routers
	authRouter := routers.UserServiceRouterGroup.Auth
	customerRouter := routers.UserServiceRouterGroup.Customer
	// Middlewares
	authMiddleware, err := injectors.InitAuthMiddleware()
	if err != nil {
		log.Fatalf("An error occurred while initializing auth middleware: %v", err)
	}

	mainGroup := r.Group("/v1")
	{
		authRouter.InitAuthRouter(mainGroup)
		customerRouter.InitCustomerRouter(mainGroup, authMiddleware)
	}

	return r
}
