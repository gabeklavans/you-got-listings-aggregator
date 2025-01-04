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

type ListingData struct {
	Refs        string  `json:"refs"`
	Price       int     `json:"price"`
	Beds        float32 `json:"beds"`
	Baths       float32 `json:"baths"`
	Date        string  `json:"date"`
	Notes       string  `json:"notes"`
	IsFavorite  bool    `json:"isFavorite"`
	IsDismissed bool    `json:"isDismissed"`
	Timestamp   int     `json:"timestamp"`
}

type Listing map[string]ListingData

type FavoriteIntent struct {
	Address    string `json:"address"`
	IsFavorite bool   `json:"isFavorite"`
}

var db *sql.DB

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

func getSites(c *gin.Context) {
	sites := make(map[string]string)

	sitesContent, err := os.ReadFile("./sites.json")
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	err = json.Unmarshal(sitesContent, &sites)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	c.JSON(http.StatusOK, sites)
}

func getListings(c *gin.Context) {
	rows, err := db.Query("SELECT * FROM listings")
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	defer rows.Close()

	listings := make(Listing)

	for rows.Next() {
		var listingData ListingData
		var listingAddr string
		if err := rows.Scan(&listingAddr, &listingData.Refs, &listingData.Price, &listingData.Beds, &listingData.Baths, &listingData.Date, &listingData.Notes, &listingData.IsFavorite, &listingData.IsDismissed, &listingData.Timestamp); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		listings[listingAddr] = listingData
	}

	if err := rows.Err(); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	c.JSON(http.StatusOK, listings)
}

func updateFavorite(c *gin.Context) {
	var body FavoriteIntent
	c.BindJSON(&body)

	_, err := db.Exec("UPDATE listings SET favorite = ? WHERE addr = ?", body.IsFavorite, body.Address)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	c.Status(http.StatusOK)
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "ygl.db")
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	createTableQuery := `CREATE TABLE IF NOT EXISTS listings (
		addr TEXT PRIMARY KEY,
		refs TEXT,
		price INTEGER,
		beds REAL,
		baths REAL,
		date TEXT,
		notes TEXT,
		favorite INTEGER,
		dismissed INTEGER,
		timestamp INTEGER
	);`

	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatal(err)
	}

	domain, found := os.LookupEnv("DOMAIN")
	if !found {
		domain = "0.0.0.0"
	}

	port, found := os.LookupEnv("PORT")
	if !found {
		port = "8083"
	}

	router := gin.Default()

	router.Use(cors.Default())

	router.Static("/static", "./public/")
	router.LoadHTMLGlob("./templates/*")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"domain": domain,
			"port":   port,
		})
	})

	v1 := router.Group("/v1")
	{
		v1.GET("/sites", getSites)
		v1.GET("/listings", getListings)
		v1.PATCH("/favorite", updateFavorite)
	}

	router.Run(fmt.Sprintf("%s:%s", domain, port))
}
