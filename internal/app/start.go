package app

import (
	"context"
	"database/sql"
	"github.com/GTedya/gophermart/config"
	"github.com/GTedya/gophermart/database"
	"github.com/GTedya/gophermart/internal/logger"
	"github.com/GTedya/gophermart/internal/repository"
	"github.com/GTedya/gophermart/internal/scheduler"
	"net/http"
)

func Run(conf config.Config) {
	log := logger.GetLogger()
	ctx := context.Background()

	db, err := repository.NewDB(conf.DatabaseURI)
	if err != nil {
		log.Error(err)
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
			log.Error("got error when closing the DB connection: ", err)
		}
	}(db)

	planner := scheduler.NewPlanner(log, db, conf)
	go planner.UpdateAccrual(ctx)

	e := NewRouter(log, db, []byte(conf.SecretKey))

	log.Fatal(http.ListenAndServe("localhost:8080", e))
}
