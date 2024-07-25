package server

import (
	"embed"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

//go:embed web/dist
var html embed.FS

const userkey = "user"

var secret = []byte("secret")

func isDebug() bool {
	return os.Getenv("LSHA_DEBUG") == "true"
}
func Serve() error {
	r := engine()
	if isDebug() {
		r.Use(static.Serve("/", static.LocalFile("cmd/internal/server/web/dist", true)))
	} else {
		r.Use(static.Serve("/", static.EmbedFolder(html, "web/dist")))
	}
	return r.Run()
}
func engine() *gin.Engine {
	gin.SetMode(gin.DebugMode)
	r := gin.New()

	// Setup the cookie store for session management
	r.Use(sessions.Sessions("mysession", cookie.NewStore(secret)))
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:*"},
		AllowWildcard:    true,
		AllowHeaders:     []string{"Content-Type", "Cookie"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	api := r.Group("/api")
	// Login and logout routes
	api.POST("/login", login)
	api.POST("/logout", logout)

	authAPI := api.Group("")
	authAPI.Use(AuthRequired)
	authAPI.GET("/me", me)
	authAPI.GET("/status", status)

	return r
}

// AuthRequired is a simple middleware to check the session.
func AuthRequired(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(userkey)
	if user == nil {
		// Abort the request with the appropriate error code
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	// Continue down the chain to handler etc
	c.Next()
}

// login is a handler that parses a form and checks for specific data.
func login(c *gin.Context) {
	session := sessions.Default(c)
	type Form struct {
		Username string `form:"username"`
		Password string `form:"password"`
	}
	var f Form
	err := c.BindJSON(&f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse form"})
		return
	}
	session.Set(userkey, strings.TrimSpace(f.Username))
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Successfully authenticated user"})
}

// logout is the handler called for the user to log out.
func logout(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(userkey)
	if user == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session token"})
		return
	}
	session.Delete(userkey)
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

// me is the handler that will return the user information stored in the
// session.
func me(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(userkey)
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// status is the handler that will tell the user whether it is logged in or not.
func status(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "You are logged in"})
}
