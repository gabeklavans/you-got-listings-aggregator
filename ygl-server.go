package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
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

type FavoriteIntent struct {
	Address    string `json:"address"`
	IsFavorite bool   `json:"isFavorite"`
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

func updateFavorite(c *gin.Context) {
	var body FavoriteIntent
	c.BindJSON(&body)

	listingsData, err := os.ReadFile("../public/data/listings.json")
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	listings := make(map[string]ListingProps)
	if err := json.Unmarshal(listingsData, &listings); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	listing, ok := listings[body.Address]
	if !ok {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	listing.IsFavorite = body.IsFavorite
	listings[body.Address] = listing

	listingsData, err = json.Marshal(listings)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	os.WriteFile("../public/data/listings.json", listingsData, 0666)

	c.Status(http.StatusOK)
}

func main() {
	db, err := sql.Open("sqlite3", "ygl.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTableQuery := `CREATE TABLE IF NOT EXISTS listings (
		id INTEGER PRIMARY KEY
	);`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()

	router.Use(cors.Default())

	router.Static("/static", "./public/")
	router.PUT("/favorite", updateFavorite)

	domain, found := os.LookupEnv("DOMAIN")
	if !found {
		domain = "0.0.0.0"
	}

	ip, found := os.LookupEnv("IP")
	if !found {
		ip = "8083"
	}

	router.Run(fmt.Sprintf("%s:%s", domain, ip))
}
