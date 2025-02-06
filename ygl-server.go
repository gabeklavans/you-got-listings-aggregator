package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

// structs
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

type Broker struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Filter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Notification struct {
	URL string `json:"url"`
}

// NOTE: key is the address
type Listings map[string]ListingData

type FavoriteIntent struct {
	Address    string `json:"address"`
	IsFavorite bool   `json:"isFavorite"`
}

var db *sql.DB
var scraperMutex sync.Mutex

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

func getListings(c *gin.Context) {
	rows, err := db.Query("SELECT * FROM Listing")
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	defer rows.Close()

	listings := make(Listings)

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

func getBrokers(c *gin.Context) {
	rows, err := db.Query("SELECT * FROM Broker")
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	defer rows.Close()

	var brokers []Broker

	for rows.Next() {
		var broker Broker
		if err := rows.Scan(&broker.URL, &broker.Name); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
		}

		brokers = append(brokers, broker)
	}

	c.JSON(http.StatusOK, brokers)
}

func updateBrokers(c *gin.Context) {
	var brokers []Broker
	err := c.BindJSON(&brokers)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	err = updateBrokersDB(brokers)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	c.Status(http.StatusOK)
}

func getFilters(c *gin.Context) {
	rows, err := db.Query("SELECT * FROM Filter")
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	defer rows.Close()

	var filters []Filter

	for rows.Next() {
		var filter Filter
		if err := rows.Scan(&filter.Name, &filter.Value); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
		}

		filters = append(filters, filter)
	}

	c.JSON(http.StatusOK, filters)
}

func updateFilters(c *gin.Context) {
	var filters []Filter
	err := c.BindJSON(&filters)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	err = updateFiltersDB(filters)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	c.Status(http.StatusOK)
}

func getNotifications(c *gin.Context) {
	rows, err := db.Query("SELECT * FROM Notification")
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	defer rows.Close()

	var notifications []Notification

	for rows.Next() {
		var notification Notification
		if err := rows.Scan(&notification.URL); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
		}

		notifications = append(notifications, notification)
	}

	c.JSON(http.StatusOK, notifications)
}

func updateNotifications(c *gin.Context) {
	var notifications []Notification
	err := c.BindJSON(&notifications)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	err = updateNotificationsDB(notifications)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	c.Status(http.StatusOK)
}

func updateFavorite(c *gin.Context) {
	var body FavoriteIntent
	err := c.BindJSON(&body)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	_, err = db.Exec("UPDATE Listing SET favorite = ? WHERE addr = ?", body.IsFavorite, body.Address)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	c.Status(http.StatusOK)
}

func runScraper(notify bool) {
	// allows us to run scraper from any goroutine without race conditions
	scraperMutex.Lock()
	defer scraperMutex.Unlock()
	fmt.Println("Starting scraper...")

	scraperCmdArgs := []string{"--db", "./data/ygl.db"}
	if notify {
		scraperCmdArgs = append(scraperCmdArgs, "--notify")
	}
	scraperCmd := exec.Command("./scraper/main.py", scraperCmdArgs...)
	scraperCmd.Env = append(os.Environ(),
		fmt.Sprintf("TG_KEY=%s", os.Getenv("TG_KEY")),
		fmt.Sprintf("CHAT_ID=%s", os.Getenv("CHAT_ID")),
	)

	scraperOut, err := scraperCmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(scraperOut))
		panic(err)
	}

	fmt.Printf("Scraper done with output:\n%s\n", string(scraperOut))
}

func updateBrokersDB(brokers []Broker) error {
	_, err := db.Exec(`DELETE FROM Broker`)
	if err != nil {
		return err
	}

	for _, broker := range brokers {
		_, err = db.Exec(`INSERT INTO Broker VALUES(?, ?)`, broker.URL, broker.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

func updateFiltersDB(filters []Filter) error {
	_, err := db.Exec(`DELETE FROM Filter`)
	if err != nil {
		return err
	}

	for _, filter := range filters {
		_, err = db.Exec(`INSERT INTO Filter VALUES(?, ?)`, filter.Name, filter.Value)
		if err != nil {
			return err
		}
	}

	return nil
}

func updateNotificationsDB(notifs []Notification) error {
	_, err := db.Exec(`DELETE FROM Notification`)
	if err != nil {
		return err
	}

	for _, notif := range notifs {
		_, err = db.Exec(`INSERT INTO Notification VALUES(?)`, notif.URL)
		if err != nil {
			return err
		}
	}

	return nil
}

func initDB() error {
	var err error

	db, err = sql.Open("sqlite3", "./data/ygl.db")
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS Listing (
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
	)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS Broker (
		url TEXT PRIMARY KEY,
		name TEXT
	)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS Filter (
		name TEXT PRIMARY KEY,
		value TEXT
	)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS Notification (
		url TEXT PRIMARY KEY
	)`)
	if err != nil {
		return err
	}

	return nil
}

func startScraperRoutine() {
	// NOTE: might want to use a Timer with some added variation on each tick
	// to be more rate-friendly to the YGL sites
	ticker := time.NewTicker(time.Hour)

	for {
		<-ticker.C
		runScraper(true)
	}
}

func main() {
	// set up env
	err := godotenv.Load()
	if err != nil {
		log.Println("Problem loading .env file; running without one")
	}

	// set up DB
	err = initDB()
	if err != nil {
		panic(err)
	}

	// start periodically scraping
	go startScraperRoutine()

	// set up web server
	domain := os.Getenv("DOMAIN")
	if domain == "" {
		domain = "0.0.0.0"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	router := gin.Default()

	router.Use(cors.Default())

	router.Static("/static", "./frontend/static/")
	router.LoadHTMLGlob("./frontend/*.html")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	v1 := router.Group("/v1")
	{
		v1.GET("/brokers", getBrokers)
		v1.PATCH("/brokers", updateBrokers)
		v1.GET("/filters", getFilters)
		v1.PATCH("/filters", updateFilters)
		v1.GET("/notifications", getNotifications)
		v1.PATCH("/notifications", updateNotifications)
		v1.GET("/listings", getListings)
		v1.PATCH("/favorite", updateFavorite)
	}

	router.Run(fmt.Sprintf("%s:%s", domain, port))
}
