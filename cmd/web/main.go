package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/PushAndRun/bookings/internal/config"
	"github.com/PushAndRun/bookings/internal/driver"
	"github.com/PushAndRun/bookings/internal/handlers"
	"github.com/PushAndRun/bookings/internal/helpers"
	"github.com/PushAndRun/bookings/internal/models"
	"github.com/PushAndRun/bookings/internal/render"
	"github.com/alexedwards/scs/v2"
)

const portNumber = ":8020"

var app config.AppConfig
var session *scs.SessionManager

func main() {

	db, err := run()
	if err != nil {
		log.Fatal(err)
	}
	defer db.SQL.Close()

	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	log.Fatal(err)

}

func run() (*driver.DB, error) {
	//define whats going to be stored in a session
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})

	//Change this to true when in production
	app.InProduction = false

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	//connect to DB
	log.Println("Connecting to database")
	db, err := driver.ConnectSQL("host=localhost port=5432 dbname=bookings user=postgres password=")
	if err != nil {
		log.Fatal("Cannot connect to database! Dying...")
	}
	log.Println("Connected to dabase!")

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache")
		return nil, err
	}

	app.TemplateCache = tc
	app.UseCache = false

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)

	render.NewRenderer(&app)

	helpers.NewHelpers(&app)

	fmt.Println(fmt.Sprintf("Starting application on port %s", portNumber))

	//handlers without routing and mux
	//http.HandleFunc("/", handlers.Repo.Home)
	//http.HandleFunc("/about", handlers.Repo.About)
	//_ = http.ListenAndServe(portNumber, nil)

	return db, nil
}
