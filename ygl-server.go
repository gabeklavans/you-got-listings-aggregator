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

// enums
type ConfigCategory string

const (
	SETTING      ConfigCategory = "SETTING"
	NOTIFICATION                = "NOTIFICATION"
	FILTER                      = "FILTER"
)

type ConfigType string

const (
	INTEGER ConfigType = "INTEGER"
	BOOLEAN            = "BOOLEAN"
	STRING             = "STRING"
	DATE               = "DATE"
)

// structs
type Broker struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Config struct {
	Name     string         `json:"name"`
	Value    string         `json:"value"`
	Type     ConfigType     `json:"type"`
	Category ConfigCategory `json:"category"`
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

func getBrokers(c *gin.Context) {
	rows, err := db.Query("SELECT * FROM Brokers")
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

func getListings(c *gin.Context) {
	rows, err := db.Query("SELECT * FROM Listings")
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

func updateFavorite(c *gin.Context) {
	var body FavoriteIntent
	err := c.BindJSON(&body)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	_, err = db.Exec("UPDATE Listings SET favorite = ? WHERE addr = ?", body.IsFavorite, body.Address)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	c.Status(http.StatusOK)
}

func getConfig(c *gin.Context) {
	var configs []Config

	rows, err := db.Query("SELECT * FROM Config")
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	defer rows.Close()

	for rows.Next() {
		var config Config
		if err := rows.Scan(&config.Name, &config.Value, &config.Type, &config.Category); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
		}

		configs = append(configs, config)
	}

	c.JSON(http.StatusOK, configs)
}

func updateConfig(c *gin.Context) {
	var config []Config
	err := c.BindJSON(&config)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	err = updateConfigDB(config)
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

	scraperCmdArgs := []string{"--db", "./ygl.db"}
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
	_, err := db.Exec(`DROP TABLE IF EXISTS Brokers`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE TABLE Brokers (
		url TEXT PRIMARY KEY,
		name TEXT
	)`)
	if err != nil {
		return err
	}

	for _, broker := range brokers {
		_, err = db.Exec(`INSERT INTO Brokers VALUES(?, ?)`, broker.URL, broker.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

func updateConfigDB(configs []Config) error {
	// reset the current config every time. simple.
	_, err := db.Exec(`DROP TABLE IF EXISTS Config`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE TABLE Config (
		name TEXT PRIMARY KEY,
		value TEXT,
		type TEXT,
		category TEXT
	)`)
	if err != nil {
		return err
	}

	for _, config := range configs {
		_, err = db.Exec(`INSERT INTO Config VALUES(?, ?, ?, ?)`,
			config.Name, config.Value, config.Type, config.Category)
		if err != nil {
			return err
		}
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
	)`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatal(err)
	}

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
		v1.GET("/config", getConfig)
		v1.PATCH("/config", updateConfig)
		v1.GET("/brokers", getBrokers)
		v1.PATCH("/brokers", updateBrokers)
		v1.GET("/listings", getListings)
		v1.PATCH("/favorite", updateFavorite)
	}

	go startScraperRoutine()

	router.Run(fmt.Sprintf("%s:%s", domain, port))
}
