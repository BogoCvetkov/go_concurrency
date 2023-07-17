package main

import (
	"log"
	"sync"

	"github.com/alexedwards/scs/v2"
	"github.com/bogo/go-concurrency/data"
	"github.com/jackc/pgx/v5"
)

type AppConfig struct {
	DB         *pgx.Conn
	ErrLog     *log.Logger
	InfoLog    *log.Logger
	Session    *scs.SessionManager
	ShutDownWG *sync.WaitGroup
	Models     data.Models
	Mailer     *Mailer
}
