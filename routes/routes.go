package routes

import (
	"github.com/yimm/rfid-api/handlers"
	"github.com/yimm/rfid-api/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	api := r.Group("/api")

	api.GET("/health", handlers.Health)

	// Public routes with rate limiting
	public := api.Group("")
	public.Use(middleware.RateLimitLogin())
	{
		public.POST("/login", handlers.Login)
	}

	// Protected routes with Sanctum auth and rate limiting
	protected := api.Group("")
	protected.Use(middleware.SanctumAuth())
	protected.Use(middleware.RateLimitProtected())
	{
		protected.POST("/logout", handlers.Logout)
		protected.GET("/master-data", handlers.MasterData)
		protected.GET("/master-rfid", handlers.MasterRfid)
		protected.GET("/transactions/history", handlers.History)
		protected.POST("/transactions/submit", handlers.Submit)
	}
}
