package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

type ConfigType int

const (
	Integer ConfigType = iota
	Boolean
	String
)

type FilterConfig int

const (
	BedsMin FilterConfig = iota
	BedsMax
	PriceMax
	DateMin
)

var filterConfigName = map[FilterConfig]string{
	BedsMin:  "bedsMin",
	BedsMax:  "bedsMax",
	PriceMax: "priceMax",
	DateMin:  "dateMax",
}

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

type SearchFilter struct {
	BedsMin  int `json:"bedsMin"`
	BedsMax  int `json:"bedsMax"`
	PriceMax int `json:"priceMax"`
	DateMin  int `json:"dateMin"`
}

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
	rows, err := db.Query("SELECT * FROM Listings")
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

func getSearchFilter(c *gin.Context) {
	rows, err := db.Query("SELECT * FROM Config")
	if err != nil {
		fmt.Println("Failed to query DB")
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	defer rows.Close()

	searchFilter := SearchFilter{BedsMin: 0, BedsMax: 0, PriceMax: 0, DateMin: 0}

	for rows.Next() {
		var name, valueStr string
		var configType ConfigType
		if err := rows.Scan(&name, &valueStr, &configType); err != nil {
			fmt.Println("Failed to scan next row in Config table")
			c.AbortWithError(http.StatusInternalServerError, err)
		}

		// technically don't need to parse if it doesn't match any of the filter configs below
		// but the Config table shouldn't ever grow that large, anyway
		var value int64
		switch configType {
		case Integer:
			value, err = strconv.ParseInt(valueStr, 0, 64)
			if err != nil {
				fmt.Println("Failed to parse " + valueStr)
				c.AbortWithError(http.StatusInternalServerError, err)
			}
		default:
			// all of the filter configs should be Integers
			value = 0
		}

		// this probably could be more "dynamic" with a map but I don't forsee this growing past a certain low-ish number of items
		switch name {
		case filterConfigName[BedsMin]:
			searchFilter.BedsMin = int(value)
		case filterConfigName[BedsMax]:
			searchFilter.BedsMax = int(value)
		case filterConfigName[PriceMax]:
			searchFilter.PriceMax = int(value)
		case filterConfigName[DateMin]:
			searchFilter.DateMin = int(value)
		}
	}

	c.JSON(http.StatusOK, searchFilter)
}

func updateFavorite(c *gin.Context) {
	var body FavoriteIntent
	c.BindJSON(&body)

	_, err := db.Exec("UPDATE Listings SET favorite = ? WHERE addr = ?", body.IsFavorite, body.Address)
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

	createTableQuery := `CREATE TABLE IF NOT EXISTS Listings (
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

	createTableQuery = `CREATE TABLE IF NOT EXISTS Config (
		name TEXT PRIMARY KEY,
		value TEXT,
		type INTEGER
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
		v1.GET("/searchFilter", getSearchFilter)
		v1.PATCH("/favorite", updateFavorite)
	}

	router.Run(fmt.Sprintf("%s:%s", domain, port))
}
