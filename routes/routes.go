package routes

import (
	"chat-app/db"
	"chat-app/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"github.com/gin-contrib/cors"
	"time"

)

func UserRoutes(r *gin.Engine) {
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true, // This allows all origins
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:   []string{"Content-Length", "Content-Type"},
		AllowCredentials: true, // Optional: only set to true if you need to send cookies
		MaxAge:          12 * time.Hour,
	}))
	r.OPTIONS("/*path", func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", c.Request.Header.Get("Origin"))
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Status(http.StatusNoContent)
	})
	
	// Register routes for posts
	r.GET("/posts", getPosts)
	r.POST("/posts", createPost)
	r.DELETE("/posts/:id", deletePost) // New DELETE route
	r.PUT("/posts/:id", updatePost)  // Route for updating a post by ID

}

func deletePost(c *gin.Context) {
	id := c.Param("id") // Get the ID from the URL parameter

	// Delete the post with the given ID from the database
	if err := db.DB.Delete(&models.Post{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}

func getPosts(c *gin.Context) {
	var posts []models.Post

	// Retrieve all posts from the database using db.DB
	if err := db.DB.Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, posts)
}

func createPost(c *gin.Context) {
	var post models.Post

	// Bind JSON data from the request to the Post struct
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Insert the post into the database using db.DB
	if err := db.DB.Create(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
		return
	}

	// Respond with the created post
	c.JSON(http.StatusCreated, gin.H{"message": "Post created successfully", "post": post})
}


func updatePost(c *gin.Context) {
	var post models.Post
	id := c.Param("id") // Get the ID from the URL parameter

	// Find the post in the database by ID
	if err := db.DB.First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	// Bind the new data from the request to the post struct
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update the post in the database
	if err := db.DB.Save(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post"})
		return
	}

	// Respond with the updated post
	c.JSON(http.StatusOK, gin.H{"message": "Post updated successfully", "post": post})
}