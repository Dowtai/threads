package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"threads/internal/broker"
	graph2 "threads/internal/graphql/graph"
	"threads/internal/graphql/loader"
	"threads/internal/repo"
	"threads/internal/repo/inmemory"
	"threads/internal/repo/postgres"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vektah/gqlparser/v2/ast"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	var storage repo.Repo
	storageType := os.Getenv("STORAGE_TYPE")

	if storageType == "postgres" {
		dbURL := os.Getenv("DATABASE_URL")

		log.Println("Starting migrations")

		m, err := migrate.New(
			"file://migrations",
			dbURL,
		)
		if err != nil {
			log.Fatalf("Could not create migrate instance: %v", err)
		}

		err = m.Up()

		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatalf("Could not run migrations: %v", err)
		}

		log.Println("Migrations successfully applied")

		if dbURL == "" {
			log.Fatal("DATABASE_URL environment variable not set")
		}

		pool, err := pgxpool.New(context.Background(), dbURL)
		if err != nil {
			log.Fatalf("Unable to create connection pool: %v", err)
		}
		defer pool.Close()

		if err := pool.Ping(context.Background()); err != nil {
			log.Fatalf("Unable to connect to database: %v", err)
		}

		log.Println("Storage: PostgreSQL")
		storage = postgres.New(pool)
	} else {
		log.Println("Storage: In-Memory")
		storage = inmemory.New()
	}

	commentBroker := broker.NewInMemoryCommentBroker()
	resolver := graph2.New(storage, commentBroker)
	srv := handler.New(graph2.NewExecutableSchema(graph2.Config{Resolvers: resolver}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(&transport.Websocket{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		KeepAlivePingInterval: 10 * time.Second,
	})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	handlerWithLoaders := loader.Middleware(storage, srv)
	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", handlerWithLoaders)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
