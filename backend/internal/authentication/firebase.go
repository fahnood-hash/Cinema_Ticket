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
	client *firebaseauth.Client
}

func NewFirebaseAuthenticator(
	credentialsPath string,
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
		client: client,
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

		c.Set("userID", token.UID)
		c.Set("email", token.Claims["email"])
		c.Next()
	}
}

func UserID(c *gin.Context) string {
	userID, _ := c.Get("userID")
	value, _ := userID.(string)

	return value
}
