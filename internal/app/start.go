package app

import (
	"database/sql"
	"github.com/GTedya/gophermart/config"
	"github.com/GTedya/gophermart/database"
	"github.com/GTedya/gophermart/internal/logger"
	"github.com/GTedya/gophermart/internal/repository"
	"net/http"
)

func Run(conf config.Config) {
	log := logger.GetLogger()

	db, err := repository.NewDB(conf.DatabaseUri)
	if err != nil {
		log.Fatal(err)
	}

	migrator, err := database.MustGetNewMigrator()
	if err != nil {
		log.Error(err)
	}

	err = migrator.ApplyMigrations(db)
	if err != nil {
		log.Error(err)
	}

	defer func(db *sql.DB) {
		er := db.Close()
		if er != nil {
			log.Fatal("got error when closing the DB connection: ", err)
		}
	}(db)

	e := NewRouter(log, db, []byte(conf.SecretKey))

	log.Fatal(http.ListenAndServe(conf.RunAddress, e))
}
