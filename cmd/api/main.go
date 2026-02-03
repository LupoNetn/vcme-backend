package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/luponetn/vcme/internal/auth"
	"github.com/luponetn/vcme/internal/config"
	"github.com/luponetn/vcme/internal/db"
)

type app struct {
	config *config.Config
	db     *pgxpool.Pool
}

// create db connection
func ConnectDb(cfg *config.Config) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("an error occurred here")
		return nil, err
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnIdleTime = time.Hour

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatal("could not create db connection")
		return nil, err
	}

	if err := db.Ping(ctx); err != nil {
		log.Fatal("Unable to ping db")
		return nil, err
	}

	return db, nil
}

func main() {
	//configure app
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("error configuring app")
	}

	//connect with the db
	dbConn, err := ConnectDb(cfg)
	if err != nil {
		log.Fatal("error starting up the db", err)
	}

	app := app{
		config: cfg,
		db:     dbConn,
	}

	queries := db.New(app.db)

	//setting up the services
	authsvc := auth.NewSvc(queries)
	authHandler := auth.NewHandler(authsvc, app.config)

	router := gin.Default()

	auth.RegisterAuthRoutes(router, authHandler)

	server := &http.Server{
		Addr:         ":" + app.config.Port,
		Handler:      router,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 30,
		IdleTimeout:  time.Second * 30,
	}

	log.Printf("starting up server on port: %s", app.config.Port)

	log.Fatal(server.ListenAndServe())

}
