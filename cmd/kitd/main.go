package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dunkinfrunkin/kit/internal/server"
	"github.com/dunkinfrunkin/kit/internal/store"
)

func main() {
	secret := os.Getenv("KIT_SECRET")
	if secret == "" {
		log.Fatal("KIT_SECRET is required")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8430"
	}

	st, err := store.New(databaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	if err := st.Migrate(); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: server.New(st, secret),
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("kitd listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-done
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
}
