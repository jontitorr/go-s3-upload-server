package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type AppState struct {
	APIKey     string
	CDNBaseURL string
	S3Bucket   string
	S3Client   *s3.S3
}

type ApiResponse struct {
	Message string `json:"message"`
}

type UploadApiResponse struct {
	Error *string `json:"error,omitempty"`
	URL   *string `json:"url,omitempty"`
}

func main() {
	godotenv.Load()
	apiKey := os.Getenv("API_KEY")
	cdnBaseURL := os.Getenv("CDN_BASE_URL")
	allowedOriginsStr := os.Getenv("ALLOWED_ORIGINS")
	allowedOrigins := []string{}
	if allowedOriginsStr != "" {
		allowedOrigins = strings.Split(allowedOriginsStr, ",")
	}
	whitelistStr := os.Getenv("IP_WHITELIST")
	whitelist := make(map[string]bool)
	for _, ip := range strings.Split(whitelistStr, ",") {
		whitelist[ip] = true
	}
	s3Bucket := os.Getenv("S3_BUCKET")
	s3Key := os.Getenv("S3_KEY")
	s3Secret := os.Getenv("S3_SECRET")
	s3Endpoint := os.Getenv("S3_ENDPOINT")
	s3Region := os.Getenv("S3_REGION")

	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(s3Key, s3Secret, ""),
		Endpoint:    aws.String(s3Endpoint),
		Region:      aws.String(s3Region),
	})
	if err != nil {
		log.Fatal(err)
	}

	s3Client := s3.New(sess)

	state := &AppState{
		APIKey:     apiKey,
		CDNBaseURL: cdnBaseURL,
		S3Bucket:   s3Bucket,
		S3Client:   s3Client,
	}

	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = allowedOrigins
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "X-Requested-With", "Content-Type", "Accept", "X-API-Key"}
	config.MaxAge = 3600
	r.Use(cors.New(config))

	r.GET("/", apiRoot)
	r.GET("/status", apiStatus)

	authorized := r.Group("/")
	authorized.Use(authMiddleware(apiKey))
	authorized.Use(IPWhiteList(whitelist))
	{
		authorized.PUT("/api/upload", uploadFile(state))
	}

	r.Run(":8080")
}

func apiRoot(c *gin.Context) {
	c.JSON(http.StatusOK, ApiResponse{Message: "Why are you here?"})
}

func apiStatus(c *gin.Context) {
	c.String(http.StatusOK, "Healthy")
}

func uploadFile(state *AppState) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, header, err := c.Request.FormFile("data")
		if err != nil {
			c.JSON(http.StatusBadRequest, UploadApiResponse{Error: aws.String("Failed to get file")})
			return
		}
		defer file.Close()

		key := c.PostForm("key")
		if key == "" {
			c.JSON(http.StatusBadRequest, UploadApiResponse{Error: aws.String("Key is required")})
			return
		}

		// Read the file content
		data, err := io.ReadAll(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, UploadApiResponse{Error: aws.String("Failed to read file")})
			return
		}

		// Prepare the S3 upload input
		uploadInput := &s3.PutObjectInput{
			Bucket:        aws.String(state.S3Bucket),
			Key:           aws.String(key),
			Body:          bytes.NewReader(data),
			ContentType:   aws.String(header.Header.Get("Content-Type")),
			ContentLength: aws.Int64(int64(len(data))),
		}

		// Upload to S3
		_, err = state.S3Client.PutObject(uploadInput)
		if err != nil {
			log.Printf("Failed to upload file to S3: %v", err)
			c.JSON(http.StatusInternalServerError, UploadApiResponse{Error: aws.String("Failed to upload file")})
			return
		}

		url := fmt.Sprintf("%s/%s", state.CDNBaseURL, key)
		c.JSON(http.StatusOK, UploadApiResponse{URL: &url})
	}
}

func authMiddleware(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		providedKey := c.GetHeader("X-API-Key")
		if providedKey != apiKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		c.Next()
	}
}

func IPWhiteList(whitelist map[string]bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !whitelist[c.ClientIP()] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"status":  http.StatusForbidden,
				"message": "Permission denied",
			})
			return
		}
	}
}
