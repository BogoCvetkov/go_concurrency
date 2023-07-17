package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/bogo/go-concurrency/data"
	database "github.com/bogo/go-concurrency/db"
	"github.com/bogo/go-concurrency/session"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fmt.Println(os.Getenv("DB_CONFIG"))
}

func main() {
	//  connect to DB
	db, err := database.InitDB()

	if err != nil {
		fmt.Println("Failed to conenct to the DB. Terminating App..")
		os.Exit(1)
		return
	}

	//create session
	sessions := session.InitializeSessionManager()

	// create channels

	//create waitgroup
	var shutDownWG sync.WaitGroup

	// initiate Mailer
	mailer := Mailer{
		Host:     "localhost",
		Port:     1025,
		Sender:   "go_concurrency@gmail.com",
		MailChan: make(chan MessageData, 10),
		DoneChan: make(chan string, 10),
		ErrChan:  make(chan error, 10),
	}

	// set up the application config
	errLog := log.New(os.Stdout, "ERROR--> ", log.Ldate|log.Ltime)
	infoLog := log.New(os.Stdout, "INFO--> ", log.Ldate|log.Ltime)
	app := AppConfig{
		DB:         db,
		ErrLog:     errLog,
		InfoLog:    infoLog,
		Session:    sessions,
		ShutDownWG: &shutDownWG,
		Models:     data.New(db),
		Mailer:     mailer.initiateDialer(),
	}
	router := app.initRoutes()

	// listen for mails
	go app.listenForEmails()

	// gracefull shutdown
	go shutGracefully(&app)

	// listen for web connections
	app.InfoLog.Println("Starting server and listenning on port 3000")
	err = http.ListenAndServe(":3000", router)

	if err != nil {
		log.Panic(err)
	}

}

func shutGracefully(app *AppConfig) {
	// Listen for the interrupt signal
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Wait for the interrupt signal
	<-sigint

	app.InfoLog.Println("Waiting for any running goroutines to finish")
	app.ShutDownWG.Wait()
	app.InfoLog.Println("Shutting server down")

	close(app.Mailer.MailChan)
	close(app.Mailer.DoneChan)
	close(app.Mailer.ErrChan)

	os.Exit(0)
}
