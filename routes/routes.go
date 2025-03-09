package routes

import (
	"chat-app/db"
	"chat-app/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"github.com/gin-contrib/cors"
	"time"
	"chat-app/middleware"
	"golang.org/x/crypto/bcrypt"
	"github.com/gorilla/websocket"
	"sync"
	"fmt"
	"strconv"



)
var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	clients   = make(map[*websocket.Conn]int) // Map WebSocket connection to user ID
	broadcast = make(chan Message)
	mu        sync.Mutex
)

// Message struct
type Message struct {
	UserID  int    `json:"user_id"`
	Content string `json:"content"`
    Fromid  int `json:"from_id" gorm:"column:from_id"`
}

func UserRoutes(r *gin.Engine) {

allowedOrigins := []string{"https://v-cart-one.vercel.app"} // Production URL

	// Check if running in local environment
	if gin.Mode() == gin.DebugMode {
		allowedOrigins = append(allowedOrigins, "http://localhost:3000") // Add localhost for development
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins, // Set specific allowed origin
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true, // Allowed because a specific origin is set
		MaxAge:           12 * time.Hour,
	}))

	// OPTIONS request handler
r.OPTIONS("/*path", func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "http://localhost:3000" || origin == "https://v-cart-one.vercel.app" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin) // Match frontend origin
		}
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Status(http.StatusNoContent)
	})


	
	// Register routes for posts
	r.GET("/posts", getPosts)
	r.POST("/posts", createPost)
	r.POST("/register", Register)
	r.POST("/login", Login)
	r.POST("/logout", Logout)
	r.GET("/users", getUsers)
	r.GET("/messages", getMessages)


	r.DELETE("/posts/:id", deletePost) // New DELETE route
	r.PUT("/posts/:id", updatePost)  // Route for updating a post by ID
	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware()) // Apply Auth Middleware
	{
		protected.GET("/", dashboard)
		protected.GET("/layout", dashboard) // Redirected here after login
	}
	r.GET("/ws", WsHandler)

	go handleMessages()

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


func Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	user.Password = string(hashedPassword)

	// Save user to DB
	if err := db.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

// Login user
func Login(c *gin.Context) {
	var user models.User
	var input models.User

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by email
	if err := db.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Compare hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	userIDStr := strconv.Itoa(int(user.ID))

	// Store user ID in cookie
	c.SetCookie("user_id", userIDStr, 3600, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{"message": "Login successful","userid":user.ID, "username":user.Username})
}

// Logout user
func Logout(c *gin.Context) {
	c.SetCookie("user_id", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

func dashboard(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Welcome to the dashboard!"})
}



func WsHandler(c *gin.Context) {
	userID, err := strconv.Atoi(c.Query("user_id")) // Pass user ID via query
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	// Register client
	mu.Lock()
	clients[conn] = userID
	mu.Unlock()


	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			mu.Lock()
			delete(clients, conn)
			mu.Unlock()
			break
		}
		fromIDStr := c.Query("from_id") // Get the query parameter as a string
		if fromIDStr != "" {            // Check if it is not empty
			fromID, err := strconv.Atoi(fromIDStr) // Convert it to an integer
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid FROM ID"})
				return
			}
			msg.Fromid = fromID
			// Use fromID (which is an int)
		}
		msg.UserID = userID
		
		broadcast <- msg
	}
}



// Function to handle WebSocket messages
func handleMessages() {
	for {
		msg := <-broadcast

		mu.Lock()
		for client, userID := range clients {
			if userID != msg.UserID { // Send to all except sender
				err := client.WriteJSON(msg)
				if err != nil {
					client.Close()
					delete(clients, client)
				}
			}
		}
		mu.Unlock()

		// Save message to DB
		db.DB.Create(&models.Message{
			UserID:  msg.UserID,
			Content: msg.Content,
			Fromid :msg.Fromid,
		})
	}
}

// func getMessages(c *gin.Context) {
// 	userID := c.Query("user_id")
// 	fromID := c.Query("from_id")

// 	fmt.Println("userID:", userID, "fromID:", fromID)

// 	var messages []Message
// 	query := db.DB.Order("created_at ASC") // Default query ordering

// 	// Fetch only if both parameters are provided
// 	if userID != "" && fromID != "" {
// 		query = query.Where("user_id = ? AND from_id = ? OR user_id = ? AND from_id = ?", userID, fromID,fromID, userID)
// 	} else {
// 		// Return an error if one parameter is missing
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id and from_id are required"})
// 		return
// 	}

// 	// Execute the query
// 	if err := query.Find(&messages).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, messages)
// }


func getMessages(c *gin.Context) {
	userID := c.Query("user_id")
	fromID := c.Query("from_id")

	fmt.Println("userID:", userID, "fromID:", fromID)

	var messages []Message
	query := db.DB.Order("created_at ASC") // Default query ordering

	// Fetch only if both parameters are provided
	if userID != "" && fromID != "" {
		query = query.Where("(user_id = ? AND from_id = ?) OR (user_id = ? AND from_id = ?)", userID, fromID, fromID, userID)
	} else {
		// Return an error if one parameter is missing
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id and from_id are required"})
		return
	}

	// Execute the query
	if err := query.Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}

	c.JSON(http.StatusOK, messages)
}




func getUsers(c *gin.Context) {
	// Retrieve user_id from the cookie
	// userID := c.Query("user_id") // Only assign 1 variable
	// .Where("id != ?", userID)

	// Convert user_id back to int

	// Fetch users excluding logged-in user
	var users []models.User
	if err := db.DB.Select("id, username, email").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

