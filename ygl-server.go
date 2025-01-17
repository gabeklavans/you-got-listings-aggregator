package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v3"
)

type ConfigType int

const (
	Integer ConfigType = iota
	Boolean
	String
	Notification
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

var filterConfigType = map[FilterConfig]ConfigType{
	BedsMin:  Integer,
	BedsMax:  Integer,
	PriceMax: Integer,
	DateMin:  Integer,
}

type Broker struct {
	Name string `json:"name" yaml:"name"`
	URL  string `json:"url" yaml:"url"`
}

type Config struct {
	Brokers       []Broker ` yaml:"brokers"`
	Filter        Filter   `yaml:"filter"`
	Notifications []string `yaml:"notifications"`
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

type Filter struct {
	BedsMin  int `json:"bedsMin" yaml:"bedsMin"`
	BedsMax  int `json:"bedsMax" yaml:"bedsMax"`
	PriceMax int `json:"priceMax" yaml:"priceMax"`
	DateMin  int `json:"dateMin" yaml:"dateMin"`
}

type FavoriteIntent struct {
	Address    string `json:"address"`
	IsFavorite bool   `json:"isFavorite"`
}

var db *sql.DB
var scraperMutex sync.Mutex

func updateFilter(filter Filter) {
	updateCommandString := `INSERT INTO Config
		VALUES(?, ?, ?)`

	_, err := db.Exec(updateCommandString, filterConfigName[BedsMin], filter.BedsMin, filterConfigType[BedsMin])
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(updateCommandString, filterConfigName[BedsMax], filter.BedsMax, filterConfigType[BedsMax])
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(updateCommandString, filterConfigName[PriceMax], filter.PriceMax, filterConfigType[PriceMax])
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(updateCommandString, filterConfigName[DateMin], filter.DateMin, filterConfigType[DateMin])
	if err != nil {
		log.Fatal(err)
	}
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

func getFilter(c *gin.Context) {
	rows, err := db.Query("SELECT * FROM Config")
	if err != nil {
		fmt.Println("Failed to query DB")
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	defer rows.Close()

	filter := Filter{BedsMin: 0, BedsMax: 0, PriceMax: 0, DateMin: 0}

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
			filter.BedsMin = int(value)
		case filterConfigName[BedsMax]:
			filter.BedsMax = int(value)
		case filterConfigName[PriceMax]:
			filter.PriceMax = int(value)
		case filterConfigName[DateMin]:
			filter.DateMin = int(value)
		}
	}

	c.JSON(http.StatusOK, filter)
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

func processConfig(config Config) {
	var err error

	// clear the whole config every time, just read it all from config.yaml
	_, err = db.Exec(`DROP TABLE IF EXISTS Config`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`DROP TABLE IF EXISTS Brokers`)
	if err != nil {
		log.Fatal(err)
	}

	// brokers
	createTableQuery := `CREATE TABLE Brokers (
		url TEXT PRIMARY KEY,
		name TEXT
	)`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatal(err)
	}

	for _, broker := range config.Brokers {
		_, err = db.Exec(`INSERT INTO Brokers
			VALUES(?, ?)`,
			broker.URL, broker.Name)
		if err != nil {
			log.Fatal(err)
		}
	}

	// config
	createTableQuery = `CREATE TABLE Config (
		name TEXT PRIMARY KEY,
		value TEXT,
		type INTEGER
	)`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatal(err)
	}
	updateCommandString := `INSERT INTO Config
		VALUES(?, ?, ?)`
	for idx, notif := range config.Notifications {
		_, err = db.Exec(updateCommandString, idx, notif, Notification)
	}

	updateFilter(config.Filter)

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
	);`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatal(err)
	}

	// read config
	configFile, err := os.ReadFile("./config.yaml")
	if err != nil {
		log.Println("Problem loading config.yaml; running without one")
	} else {
		config := Config{}
		err = yaml.Unmarshal(configFile, &config)
		if err != nil {
			log.Fatal(err)
		}

		// NOTE: Broker info is also processed in here
		processConfig(config)
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
		v1.GET("/brokers", getBrokers)
		v1.GET("/listings", getListings)
		v1.GET("/filter", getFilter)
		v1.PATCH("/favorite", updateFavorite)
	}

	go func() {
		runScraper(false)
	}()

	go startScraperRoutine()

	router.Run(fmt.Sprintf("%s:%s", domain, port))
}
