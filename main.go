package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/JulianWan1/rssagg/internal/database"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

type apiConfig struct {
	DB *database.Queries
}

func main() {
		godotenv.Load()

		portString := os.Getenv("PORT")
		if portString == "" {
			log.Fatal("PORT is not found in the environment")
		}

		dbURL := os.Getenv("DB_URL")
		if dbURL == "" {
			log.Fatal("DB_URL is not found in the environment")
		}

		conn, err := sql.Open("postgres", dbURL)
		if err != nil{
			log.Fatal("Can't connect to database")
		}

		db := database.New(conn)

		apiCfg := apiConfig{
			DB: db,
		}

		go startScraping(db, 10, time.Minute)

		router := chi.NewRouter()

		router.Use(cors.Handler(cors.Options{
			AllowedOrigins: 	[]string{"https://*", "http://*"},
			AllowedMethods: 	[]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: 	[]string{"*"},
			ExposedHeaders: 	[]string{"Link"},
			AllowCredentials: false,
			MaxAge: 					300,
		}))

		v1Router := chi.NewRouter()
		v1Router.Get("/healthz", handlerReadiness)
		v1Router.Get("/err", handlerErr)
		v1Router.Post("/users", apiCfg.handlerCreateUser)
		v1Router.Get("/users", apiCfg.middleWareAuth(apiCfg.handlerGetUser))

		v1Router.Post("/feeds", apiCfg.middleWareAuth(apiCfg.handlerCreateFeed))
		v1Router.Get("/feeds", apiCfg.handlerGetFeeds)

		v1Router.Get("/posts", apiCfg.middleWareAuth(apiCfg.handlerGetPostsForUser))

		v1Router.Post("/feed_follows", apiCfg.middleWareAuth(apiCfg.handlerCreateFeedFollow))
		v1Router.Get("/feed_follows", apiCfg.middleWareAuth(apiCfg.handlerGetFeedFollows))
		v1Router.Delete("/feed_follows/{feedFollowID}", apiCfg.middleWareAuth(apiCfg.handlerDeleteFeedFollow))

		router.Mount("/v1", v1Router)

		srv := &http.Server{
			Handler: router,
			Addr:		 ":" + portString,
		}

		log.Printf("Server starting on port %v", portString)
		err = srv.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
}