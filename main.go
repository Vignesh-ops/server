package main

import (


    "net/http"

    "github.com/gin-gonic/gin"

    "chat-app/routes"
     "chat-app/db"

)



// Main function
func main() {
    r := gin.Default()

    // Connect to the database
	db.ConnectDatabase()
	routes.UserRoutes(r)

    // Routes
    r.GET("/ping", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"message": "pong"})
    })

    // Start the server
    r.Run(":8080")
}
