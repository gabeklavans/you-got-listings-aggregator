package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type ListingProps struct {
	Refs        []string `json:"refs"`
	Price       int      `json:"price"`
	Beds        float32  `json:"beds"`
	Baths       float32  `json:"baths"`
	Date        string   `json:"date"`
	Notes       string   `json:"notes"`
	IsFavorite  bool     `json:"isFavorite"`
	IsDismissed bool     `json:"isDismissed"`
}

func basicAuth(c *gin.Context) {
	user, password, hasAuth := c.Request.BasicAuth()
	if hasAuth && user == os.Getenv("AUTH_USER") && password == os.Getenv("AUTH_PASS") {
		c.Next()
	} else {
		c.Abort()
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
}

func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalf("Error loading .env: %s", err)
	}

	router := gin.Default()

	router.Use(cors.Default())

	router.Static("/static", "../public/")

	domain := os.Getenv("DOMAIN")
	ip := os.Getenv("IP")
	router.Run(fmt.Sprintf("%s:%s", domain, ip))
}
