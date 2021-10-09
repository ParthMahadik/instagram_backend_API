package main

import (
	"context"
	"encoding/json"
	"fmt"

	"net/http"
	"time"

	// mgo "gopkg.in/mgo.v2"
	// "gopkg.in/mgo.v2/bson"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var client *mongo.Client

// declaring constants
const (
	hostname       string = "localhost:27017"
	dbName         string = "demo_instagram"
	collectionName string = "instagram"
	port           string = ":9000"
)

//struct declaring for json and bson
type (
	Post struct {
		ID        string    `json:"id" bson:"id"`
		UserId    string    `json:"userid" bson:"userid"`
		Caption   string    `json:"caption" bson:"caption"`
		Imageurl  string    `json:"imageurl" bson:"imageurl"`
		CreatedAt time.Time `json:"createdat" bson:"createdat"`
	}

	User struct {
		USERID   string `json:"userid" bson:"userid"`
		Name     string `json:"name" bson:"name"`
		Email    string `json:"email" bson:"email"`
		Password string `json:"password" bson:"password"`
	}
)

func main() {
	//MongoDB connect
	fmt.Println("Starting the application...")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ = mongo.Connect(ctx, clientOptions)
	router := mux.NewRouter() //Router instance initiate

	router.HandleFunc("/posts", CreatePostEndpoint).Methods("POST") //HandleFunc for Routing of Handler functions
	router.HandleFunc("/users", CreateUserEndpoint).Methods("POST")
	router.HandleFunc("/users/posts/{userid}", GetUserPostsEndpoint).Methods("GET")
	router.HandleFunc("/posts/{id}", GetPostEndpoint).Methods("GET")
	router.HandleFunc("/users/{userid}", GetUserEndpoint).Methods("GET")

	http.ListenAndServe(port, router)
}

//function to create post
func CreatePostEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var post Post
	// _ = json.NewDecoder(request.Body).Decode(&post)
	json.NewDecoder(request.Body).Decode(&post)
	collection := client.Database(dbName).Collection(collectionName)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result, _ := collection.InsertOne(ctx, post)
	json.NewEncoder(response).Encode(result)
}

//function to list all posts of a particular user
func GetUserPostsEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var timeline []Post
	collection := client.Database(dbName).Collection(collectionName)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	params := mux.Vars(request)
	userid := params["userid"]
	filter := bson.M{"userid": userid}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var post Post
		cursor.Decode(&post)
		timeline = append(timeline, post)
	}
	if err := cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(timeline)
}

//function to get post
func GetPostEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	params := mux.Vars(request)
	id := params["id"]
	var post Post
	collection := client.Database(dbName).Collection(collectionName)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	filter := bson.M{"id": id}
	err := collection.FindOne(ctx, filter).Decode(&post)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(post)
}

//function to create user
func CreateUserEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var user User
	_ = json.NewDecoder(request.Body).Decode(&user)
	// json.NewDecoder(request.Body).Decode(&user)
	collection := client.Database(dbName).Collection(collectionName)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	user.Password, _ = HashPassword(user.Password)
	print(user.Password)
	result, _ := collection.InsertOne(ctx, user)
	json.NewEncoder(response).Encode(result)
}

//function to get user info based on Id
func GetUserEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	// params := mux.Vars(request)
	// userid := params["userid"]
	var user User
	params := mux.Vars(request)
	userid := params["userid"]
	filter := bson.M{"userid": userid}
	collection := client.Database(dbName).Collection(collectionName)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	err := collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(user)
}

//Function for hashing
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
