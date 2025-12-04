package api

import (
	"net/http"
	"strconv"

	"github.com/Iknite-Space/sqlc-example-api/db/repo"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Request structs for our content body for "post"

type CreatePostRequest struct {
	UserID  int32  `json:"user_id" binding:"required"`
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type DeletePostRequest struct {
	UserID int32 `json:"user_id" binding:"required"`
}

type UpdatePostContentRequest struct {
	UserID  int32  `json:"user_id" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// Request structs for our content body for "user"

type RegisterRequest struct {
	Username string `json:"user_name" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// APIHandler is the main controller struct that holds all the queriers/dependencies
// and implements the router wiring.
type APIHandler struct {
	querier repo.Querier
}

func NewAPIHandler(querier repo.Querier) *APIHandler {
	return &APIHandler{
		querier: querier,
	}
}

// gin and routing
func (h *APIHandler) WireHttpHandler() http.Handler {

	r := gin.Default()
	r.Use(gin.CustomRecovery(func(c *gin.Context, _ any) {
		c.String(http.StatusInternalServerError, "Internal Server Error: panic")
		c.AbortWithStatus(http.StatusInternalServerError)
	}))

	// User Routes
	r.POST("/register", h.RegisterUser)
	r.POST("/login", h.LoginUser)
	r.GET("/users", h.ListAllUsers)
	r.GET("/user/:id", h.GetUserByID)

	// Post Routes
	r.POST("/posts", h.CreatePost)
	r.GET("/posts", h.ListAllPosts)
	r.GET("/posts/:id", h.GetPostByID)
	r.DELETE("/posts/:id", h.DeletePost)
	r.PUT("/posts/:id", h.UpdatePost)

	return r
}

func (h *APIHandler) CreatePost(c *gin.Context) {
	var req CreatePostRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post input (requires user_id, title, content): " + err.Error()})
		return
	}

	params := repo.CreatePostParams{
		UserID:  req.UserID,
		Title:   req.Title,
		Content: req.Content,
	}

	post, err := h.querier.CreatePost(c, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, post)
}

func (h *APIHandler) ListAllPosts(c *gin.Context) {
	posts, err := h.querier.ListAllPosts(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve posts: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, posts)
}

func (h *APIHandler) GetPostByID(c *gin.Context) {
	postIDStr := c.Param("id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID format"})
		return
	}

	post, err := h.querier.GetPostByID(c, int32(postID))
	if err != nil {
		if err.Error() == "no rows in result set" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve post: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, post)
}

func (h *APIHandler) DeletePost(c *gin.Context) {
	postIDStr := c.Param("id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID format"})
		return
	}

	var req DeletePostRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required for deletion (must be a number)"})
		return
	}

	err = h.querier.DeletePost(c, repo.DeletePostParams{
		ID:     int32(postID),
		UserID: req.UserID,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post or post not found/unauthorized: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}

func (h *APIHandler) UpdatePost(c *gin.Context) {
	postIDStr := c.Param("id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID format"})
		return
	}

	var req UpdatePostContentRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID and new content are required."})
		return
	}

	updatedPost, err := h.querier.UpdatePostContent(c, repo.UpdatePostContentParams{
		ID:      int32(postID),
		Content: req.Content,
		UserID:  req.UserID,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post (Post not found or unauthorized): " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedPost)
}

// User Handler Implementations

func (h *APIHandler) RegisterUser(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	params := repo.CreateUserParams{
		UserName:     req.Username,
		PasswordHash: string(hashedPassword),
	}

	user, err := h.querier.CreateUser(c, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Registration failed (Username may be taken)"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": user.ID, "username": user.UserName, "message": "Worker registered successfully."})
}

func (h *APIHandler) LoginUser(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	user, err := h.querier.GetUserByUsername(c, req.Username)
	if err != nil {
		if err.Error() == "no rows in result set" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password" + err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Login failed"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": user.ID, "username": user.UserName, "message": "Login successful."})
}

func (h *APIHandler) GetUserByID(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid worker ID format"})
		return
	}

	user, err := h.querier.GetUserByID(c, int32(userID))
	if err != nil {
		if err.Error() == "no rows in result set" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Worker not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve worker data"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *APIHandler) ListAllUsers(c *gin.Context) {
	users, err := h.querier.ListUsers(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}
