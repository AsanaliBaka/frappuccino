package cmd

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"os"
	"strconv"

	"frappuccino/config"
	"frappuccino/internal/handler"
	internal "frappuccino/internal/repo"
	"frappuccino/internal/service"
	"frappuccino/internal/slog"

	_ "github.com/lib/pq"
)

var (
	port  int
	dbURL string
)

func Run() {
	config.LoadEnv()
	flag.IntVar(&port, "port", 8080, "Port number")
	flag.StringVar(&dbURL, "db", os.Getenv("DATABASE_URL"), "Database connection URL")
	flag.Parse()

	if port < 0 || port > 65535 {
		log.Fatal("Invalid port number")
	}

	slog.Init()

	ctx := context.Background()

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err = db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	container := internal.New(db)

	invService := service.NewInventoryService(container)
	menuService := service.NewMenuService(container)
	orderService := service.NewOrderService(container)
	statsService := service.NewStatsService(container)

	h := handler.NewHandler(invService, menuService, orderService, statsService)

	srv := handler.NewServer(strconv.Itoa(port), h)
	srv.Start()
}
