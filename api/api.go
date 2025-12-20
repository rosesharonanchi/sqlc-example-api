package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/jackc/pgx/v5"
	"github.com/Iknite-Space/sqlc-example-api/db/repo"
	"golang.org/x/crypto/bcrypt"
)

func HashedPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func CheckPassword(password string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

type UserClaims struct {
	ID       int32  `json:"id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func GenerateToken(userID int32, username string, secretKey string, duration time.Duration) (string, error) {
	claims := UserClaims{
		ID:       userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			// Set token expiration relative to the current time (e.g., 1 hour)
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secretKey))
}

// server structure and API handler
type Server struct {
	store     *repo.Queries
	JWTSecret string
}

func NewAPIHandler(querier *repo.Queries, jwtSecret string) *Server {
	return &Server{
		store:     querier,
		JWTSecret: jwtSecret,
	}
}

func (server *Server) WireHttpHandler() http.Handler {
	router := gin.Default()
	//user routes
	router.POST("/signup", server.signup)
	router.POST("/login", server.login)
	//post routes
	router.POST("/post", server.createPost)
	router.GET("/post/:id", server.getPost)
	router.GET("/post", server.listPosts)
	router.PUT("/posts", server.updatePost)
	router.DELETE("/posts/:id", server.deletePost)

	return router
}

type signupRequest struct {
	Username string `json:"username" binding:"required,alphanum"` // Must be present, must only contain letters/numbers.
	Email    string `json:"email" binding:"required,email"`       // Must be present, must be a valid email format.
	Password string `json:"password" binding:"required,min=6"`
}

func (server *Server) signup(c *gin.Context) {
	var req signupRequest
	// bind and validate request body
	if err := c.ShouldBindJSON(&req); err != nil { //(**important)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//step2 hash the password securely
	hashedPassword, err := HashedPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error:": "failed to hash password, username or email already taken"})
		return
	}
	//step 3: Prepare parameters for the sqlc function
	arg := repo.CreateUserParams{
		Username:       req.Username,
		Email:          req.Email,
		HashedPassword: hashedPassword,
	}

	//create the user using the generated go function (**important)
	_, err = server.store.CreateUser(c, arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "User crested successfully"})

}

// login
type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type loginResponse struct {
	AccessToken string `json:"access_token"`
	Username    string `json:"username"`
}

func (server *Server) login(c *gin.Context) {
	var req loginRequest
	//step 1 bind and vaidate request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}

	//step 2 retrieve the user
	user, err := server.store.GetUseryByEmail(c, req.Email)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	//step 3 verify password
	err = CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	//step 4 generate the jwt token
	tokenDuration := 24 * time.Hour
	token, err := GenerateToken(user.ID, user.Username, server.JWTSecret, tokenDuration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create token"})
		return
	}

	//step 5 send the response
	rsp := loginResponse{
		AccessToken: token,
		Username:    user.Username,
	}
	c.JSON(http.StatusOK, rsp)

}

// POST
type createPostRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	UserID  int32  `json:"user_id" binding:"required"`
}

func (server *Server) createPost(c *gin.Context) {
	var req createPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	arg := repo.CreatePostParams{
		Title:   req.Title,
		Content: req.Content,
		UserID:  req.UserID,
	}

	post, err := server.store.CreatePost(c, arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create post"})
		return
	}

	c.JSON(http.StatusCreated, post)
}

// get all post
type listPostsRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=20"`
}

func (server *Server) listPosts(c *gin.Context) {
	var req listPostsRequest
	// We use ShouldBindQuery for ?page_id=1&page_size=10
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Calculate OFFSET for SQL (e.g., Page 2 with size 10 starts at record 10)
	arg := repo.ListPostsParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	posts, err := server.store.ListPosts(c, arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve posts"})
		return
	}

	c.JSON(http.StatusOK, posts)
}

// get post by id
type getPostRequest struct {
	ID int32 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getPost(c *gin.Context) {
	var req getPostRequest
	// Notice we use ShouldBindUri because the ID is in the URL, not the JSON body
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	post, err := server.store.GetPost(c, req.ID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	c.JSON(http.StatusOK, post)
}

// update post
type updatePostRequest struct {
	ID      int32  `json:"id" binding:"required,min=1"`
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}

func (server *Server) updatePost(c *gin.Context) {
	var req updatePostRequest
	// We bind the JSON body which includes the ID and new data
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	arg := repo.UpdatePostParams{
		ID:      req.ID,
		Title:   req.Title,
		Content: req.Content,
	}

	post, err := server.store.UpdatePost(c, arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update post"})
		return
	}

	c.JSON(http.StatusOK, post)
}

// delete
type deletePostRequest struct {
	ID int32 `uri:"id" binding:"required,min=1"`
}

func (server *Server) deletePost(c *gin.Context) {
	var req deletePostRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := server.store.DeletePost(c, req.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}
