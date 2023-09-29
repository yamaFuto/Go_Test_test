package main

import (
	"encoding/gob"
	"flag"
	"log"
	"net/http"
	"webApp/pkg/data"
	"webApp/pkg/repository"
	"webApp/pkg/repository/dbrepo"

	"github.com/alexedwards/scs/v2"
)

type application struct {
	DSN     string
	DB      repository.DatabaseRepo
	Session *scs.SessionManager
}

func main() {
	// sessionはstructを格納するときには型を登録しなくてはならない
	gob.Register(data.User{})
	// set up an app config
	app := application{}

	//commandLineで使用できるdsn flagを作りそれをapp.DSNに代入する
	// 主導でdbに接続するために用意
	flag.StringVar(&app.DSN, "dsn", "host=localhost port=5432 user=postgres password=postgres dbname=users sslmode=disable timezone=UTC connect_timeout=5", "Postgres connection")
	flag.Parse()

	conn, err := app.connectToDB()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	app.DB = &dbrepo.PostgresDBRepo{DB: conn}

	// get a session manager
	app.Session = getSession()

	// print out a message
	log.Println("Starting server on port 8080...")

	// start the server
	err = http.ListenAndServe(":8080", app.routes())
	if err != nil {
		log.Fatal(err)
	}
}
