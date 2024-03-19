package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/ferretcode-freelancing/blog/admin"
	"github.com/ferretcode-freelancing/blog/auth"
	"github.com/ferretcode-freelancing/blog/blog"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal(err)
	}

	dsn := os.Getenv("POSTGRES_URL")

	if dsn == "" {
		log.Fatal("the dsn environment variable was not found")
	}

	dsn = strings.ReplaceAll(dsn, "\n", "")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err != nil {
		log.Fatal(err)
	}

	err = db.AutoMigrate(&admin.Post{}, &auth.User{})

	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		err := blog.Articles(w, r, db)
		handleError(w, err)
	})

	r.Route("/auth", func(r chi.Router) {
		r.Get("/create_account", func(w http.ResponseWriter, r *http.Request) {
			err := auth.ShowCreateAccount(w, r)
			handleError(w, err)
		})

		r.Get("/sign_in", func(w http.ResponseWriter, r *http.Request) {
			err := auth.ShowSignIn(w, r)
			handleError(w, err)
		})

		r.Get("/create", func(w http.ResponseWriter, r *http.Request) {
			err := auth.CreateAccount(w, r, db)
			handleError(w, err)
		})

		r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
			err := auth.SignIn(w, r, db)
			handleError(w, err)
		})
	})

	r.Route("/admin", func(r chi.Router) {
		r.Use(auth.CheckAuth(db))

		r.Get("/new_post", func(w http.ResponseWriter, r *http.Request) {
			err := admin.NewPost(w, r, db)
			handleError(w, err)
		})

		r.Post("/new_post", func(w http.ResponseWriter, r *http.Request) {
			err := admin.CreateNewPost(w, r, db)
			handleError(w, err)
		})
	})

	r.Route("/blog", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			err := blog.Articles(w, r, db)
			handleError(w, err)
		})

		r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
			err := blog.Article(w, r, db)
			handleError(w, err)
		})
	})

	err = http.ListenAndServe(":"+os.Getenv("PORT"), r)

	if err != nil {
		log.Fatal(err)
	}
}

func handleError(w http.ResponseWriter, err error) {
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
