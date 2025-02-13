package routes
// import and initialize static index.html file
import (
    "github.com/gin-gonic/gin"
    httpSwagger "github.com/swaggo/http-swagger"
    "github.com/maazxenon/task-api/handlers"
    "time"
    "github.com/gin-contrib/cors"
)

// TaskRouter returns a new router
func TaskRouter() *gin.Engine {
    r := gin.Default()
    config := cors.DefaultConfig()
    config.AllowAllOrigins = true
    config.AllowMethods = []string{"POST", "GET", "PUT", "OPTIONS", "DELETE"}
    config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Accept", "User-Agent", "Cache-Control", "Pragma"}
    config.ExposeHeaders = []string{"Content-Length"}
    config.AllowCredentials = true
    config.MaxAge = 12 * time.Hour

    r.Use(cors.New(config))
    r.Use(gin.Recovery()) // Add recovery middleware

    // Serve static files
    r.Static("/static", "./static")

    // at / server index.html

    r.GET("/", func(c *gin.Context) {
        c.File("./static/index.html")
    })



    // Serve Swagger UI
    r.GET("/swagger/*any", gin.WrapH(httpSwagger.WrapHandler))

    r.GET("/tasks", handlers.IndexHandler)
    r.POST("/tasks", handlers.CreateHandler)
    r.GET("/tasks/:id", handlers.GetTaskHandler)
    r.PUT("/tasks/:id", handlers.UpdateTaskHandler)
    r.DELETE("/tasks/:id", handlers.DeleteHandler)

    return r
}