package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var db *mgo.Database

const (
	hostname       string = "localhost:27017"
	dbName         string = "demo_instagram"
	collectionName string = "instagram"
	port           string = ":9000"
)

type (
	postModel struct {
		ID        string    `bson:"id"`
		UserId    string    `bson:"userid"`
		Caption   string    `bson:"caption"`
		Imageurl  string    `bson:"imageurl"`
		CreatedAt time.Time `bson:"createdAt"`
	}

	post struct {
		ID        string    `json:"id"`
		UserId    string    `json:"userid"`
		Caption   string    `json:"title"`
		Imageurl  string    `json:"imageurl"`
		CreatedAt time.Time `json:"created_at"`
	}
)

func init() {
	sess, err := mgo.Dial(hostname)
	checkErr(err)
	sess.SetMode(mgo.Monotonic, true)
	db = sess.DB(dbName)
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err) //respond with error page or message
	}
}

func main() {
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Mount("/", todoHandlers())

	srv := &http.Server{
		Addr:         port,
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Println("Listening on port ", port)
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("listen: %s\n", err)
		}
	}()

	<-stopChan
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	srv.Shutdown(ctx)
	defer cancel()
	log.Println("Server gracefully stopped!")

}

func createPost(w http.ResponseWriter, r *http.Request) {
	var p post

	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		println(err)
		return
	}

	// simple validation
	// if p.ID == "" {
	// 	println("Null Id")
	// 	return
	// }

	// if input is okay, create a post
	pm := postModel{
		ID:        p.ID,
		UserId:    p.UserId,
		Caption:   p.Caption,
		Imageurl:  p.Imageurl,
		CreatedAt: time.Now(),
	}
	// if err := db.C(collectionName).Insert(&pm); err != nil {
	// 	panic("error while adding to db")
	// 	return
	// }
	db.C(collectionName).Insert(&pm)

}

func fetchPost(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "id"))

	// if !bson.IsObjectIdHex(id) {
	// 	rnd.JSON(w, http.StatusBadRequest, renderer.M{
	// 		"message": "The id is invalid",
	// 	})
	// 	return
	// }

	var p post

	// if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
	// 	panic("error")
	// 	return
	// }
	json.NewDecoder(r.Body).Decode(&p)

	// simple validation
	// if p.ID == "" {
	// 	panic("null id")
	// 	return
	// }

	// if input is okay, update a todo
	// if err := db.C(collectionName).
	// 	Find(
	// 		bson.M{"id": id},
	// 	); err != nil {
	// 	panic("failed to fetch")
	// 	return
	// }
	db.C(collectionName).
		Find(
			bson.M{"id": id},
		)
}

func todoHandlers() http.Handler {
	rg := chi.NewRouter()
	rg.Group(func(r chi.Router) {
		r.Get("/posts/{id}", fetchPost)
		r.Post("/posts", createPost)
		// r.Put("/{id}", updateTodo)
		// r.Delete("/{id}", deleteTodo)
	})
	return rg
}
