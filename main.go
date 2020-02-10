package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func handler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	name := query.Get("name")

	if name == "" {
		name = "Guest"
	}
	log.Printf("Received request for %s\n", name)
	w.Write([]byte(fmt.Sprintf("Hello, %s", name)))
}

func healtHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func waitForShutdown(srv *http.Server) {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-interruptChan

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	srv.Shutdown(ctx)

	log.Println("Shutting Down")
	os.Exit(0)
}

type dbConn struct {
	db *gorm.DB
}

type Product struct {
	gorm.Model
	Code  string
	Price uint
}

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func String(length int) string {
	return StringWithCharset(length, charset)
}

func (db *dbConn) getProductsHandler(w http.ResponseWriter, r *http.Request) {
	products := []Product{}
	db.db.Find(&products)

	js, err := json.Marshal(products)
	if err != nil {
		log.Fatal(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
func (db *dbConn) createProductHandler(w http.ResponseWriter, r *http.Request) {
	min := 10000
	max := 50000
	randPrice := rand.Intn(max-min) + min
	productName := String(10)
	product := Product{
		Code:  "cd-" + productName,
		Price: uint(randPrice),
	}
	db.db.Create(&product)

	js, err := json.Marshal(product)
	if err != nil {
		log.Fatal(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func main() {
	// var regexPlusLeadingNumber = regexp.MustCompile(`^\+*`)
	// var regexZeroLeadingNumber = regexp.MustCompile(`^\+{1}0+`)

	// destination := "+6281382171273"
	// destination := "081382171273"
	// fmt.Println(destination)
	// destination = regexPlusLeadingNumber.ReplaceAllString(destination, "+")
	// destination = regexZeroLeadingNumber.ReplaceAllString(destination, "+62")
	// fmt.Println(destination)

	dbDriver := os.Getenv("go_kubernetes_db_driver")
	dbHost := os.Getenv("go_kubernetes_db_host")
	dbPort := os.Getenv("go_kubernetes_db_port")
	dbName := os.Getenv("go_kubernetes_db_name")
	dbUser := os.Getenv("go_kubernetes_db_user")
	dbPassword := os.Getenv("go_kubernetes_db_password")
	connString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := gorm.Open(dbDriver, connString)
	if err != nil {
		log.Fatalf("Failed to start, error connect to DB MySQL | %v", err)
	}
	defer db.Close()

	// Migrate the schema
	// db.AutoMigrate(&Product{})

	conn := &dbConn{
		db: db,
	}

	r := mux.NewRouter()

	r.HandleFunc("/", handler)
	r.HandleFunc("/healt-check", healtHandler)
	r.HandleFunc("/readiness", readinessHandler)
	r.HandleFunc("/products", conn.getProductsHandler)
	r.HandleFunc("/product/create", conn.createProductHandler)

	srv := &http.Server{
		Handler:      r,
		Addr:         ":8000",
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}

	// Configure Logging
	LOG_FILE_LOCATION := os.Getenv("LOG_FILE_LOCATION")
	if LOG_FILE_LOCATION != "" {
		log.SetOutput(&lumberjack.Logger{
			Filename:   LOG_FILE_LOCATION,
			MaxSize:    500, // megabytes
			MaxBackups: 3,
			MaxAge:     28,   //days
			Compress:   true, // disabled by default
		})
	}

	go func() {
		log.Println("Server running on localhost:8000")
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	waitForShutdown(srv)
}
