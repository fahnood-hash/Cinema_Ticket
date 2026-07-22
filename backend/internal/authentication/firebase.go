package authentication

import (
	"context"
	"fmt"
	"strings"

	firebase "firebase.google.com/go/v4"
	firebaseauth "firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
)

type FirebaseAuthenticator struct {
	client     *firebaseauth.Client
	adminEmail string
}

func NewFirebaseAuthenticator(
	credentialsPath string,
	adminEmail string,
) (*FirebaseAuthenticator, error) {
	app, err := firebase.NewApp(
		context.Background(),
		nil,
		option.WithCredentialsFile(credentialsPath),
	)
	if err != nil {
		return nil, fmt.Errorf("initialize Firebase: %w", err)
	}

	client, err := app.Auth(context.Background())
	if err != nil {
		return nil, fmt.Errorf("create Firebase auth client: %w", err)
	}

	return &FirebaseAuthenticator{
		client:     client,
		adminEmail: adminEmail,
	}, nil
}

func (a *FirebaseAuthenticator) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")

		if !strings.HasPrefix(header, "Bearer ") {
			c.JSON(401, gin.H{
				"error": "missing Firebase bearer token",
			})
			c.Abort()
			return
		}

		idToken := strings.TrimPrefix(header, "Bearer ")

		token, err := a.client.VerifyIDToken(c.Request.Context(), idToken)
		if err != nil {
			c.JSON(401, gin.H{
				"error": "invalid Firebase token",
			})
			c.Abort()
			return
		}
		email, _ := token.Claims["email"].(string)
		fmt.Println("Firebase email:", email)
		fmt.Println("Admin email:", a.adminEmail)

		role := "USER"
		if strings.EqualFold(email, a.adminEmail) {
			role = "ADMIN"
		}

		c.Set("userID", token.UID)
		c.Set("email", email)
		c.Set("role", role)
		c.Next()
	}
}

func UserID(c *gin.Context) string {
	userID, _ := c.Get("userID")
	value, _ := userID.(string)

	return value
}

func (a *FirebaseAuthenticator) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString("role") != "ADMIN" {
			c.JSON(403, gin.H{
				"error": "admin access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
