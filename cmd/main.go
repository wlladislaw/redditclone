package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"redditclone/pkg"
	"redditclone/pkg/middleware"
	"redditclone/pkg/session"
	"time"

	"database/sql"

	"github.com/redis/go-redis/v9"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	commentDeliveryPkg "redditclone/pkg/comment/delivery"
	postDeliveryPkg "redditclone/pkg/post/delivery"
	postsRepoPkg "redditclone/pkg/post/repo"
	userDeliveryPkg "redditclone/pkg/user/delivery"
	userRepoPkg "redditclone/pkg/user/repo"
)

const (
	PG_DB_NAME = "redditclone"
	PG_DB_USER = "gouser"
	PG_DB_PASS = "GOUser1"
)

func mongoDB(url string) *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	client, err := mongo.Connect(options.Client().ApplyURI(url))
	if err != nil {
		log.Fatalln("cant connect to mongodb", err)
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatalln("cant ping mongodb", err)
	}
	return client
}

func main() {
	pgConn := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", PG_DB_USER, PG_DB_PASS, PG_DB_NAME)
	dbPg, err := sql.Open("postgres", pgConn)
	if err != nil {
		log.Fatalln("cant parse config:", err)
	}
	errPing := dbPg.Ping()
	if err != nil {
		log.Fatalln(errPing)
	}
	defer dbPg.Close()

	mongoClient := mongoDB("mongodb://localhost:27017")
	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			log.Fatalln(err)
		}
	}()
	mongoDB := mongoClient.Database("redditclone")
	collectionPosts := &postsRepoPkg.MongoCollection{
		Сoll: mongoDB.Collection("posts"),
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	port := ":8080"
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("cant setup zap logger %v", err)
	}
	defer logger.Sync()
	logger.Info("server starts",
		zap.String("type", "Start"),
		zap.String("port", port),
	)

	var dir string
	flag.StringVar(&dir, "dir", "./template/static", "")
	flag.Parse()

	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(dir))))

	userRepo := userRepoPkg.NewUserRepo(dbPg)
	postsRepo := postsRepoPkg.NewRepo(collectionPosts)
	sm := session.NewSessionManager(rdb)

	userHandler := &userDeliveryPkg.UserHandler{
		UserRepo: userRepo,
		Session:  sm,
	}

	postHandler := &postDeliveryPkg.PostsHandler{
		PostsRepo: postsRepo,
	}
	commentHandler := &commentDeliveryPkg.CommentHandler{
	  PostsRepo: postsRepo,
		Logger:    logger,
	}

	auth := middleware.AuthMiddleware(sm)
	r.Use(middleware.SetupLogger(logger))
	r.Use(middleware.LogMiddleware(logger))

	r.HandleFunc("/", pkg.Index)
	r.HandleFunc("/api/register", userHandler.SignUp).Methods("POST")
	r.HandleFunc("/api/login", userHandler.Login).Methods("POST")
	r.HandleFunc("/api/posts/", postHandler.List).Methods("GET")
	r.Handle("/api/posts", auth(http.HandlerFunc(postHandler.AddPost))).Methods("POST")
	r.HandleFunc("/api/posts/{CATEGORY_NAME}", postHandler.ListByCategory).Methods("GET")
	r.HandleFunc("/api/post/{POST_ID}", postHandler.GetByID).Methods("GET")
	r.Handle("/api/post/{POST_ID}", auth(http.HandlerFunc(postHandler.Delete))).Methods("DELETE")
	r.HandleFunc("/api/user/{USER_LOGIN}", postHandler.ListByLogin).Methods("GET")
	r.Handle("/api/post/{POST_ID}", auth(http.HandlerFunc(commentHandler.CreateComment))).Methods("POST")
	r.Handle("/api/post/{POST_ID}/{COMMENT_ID}", auth(http.HandlerFunc(commentHandler.DeleteComment))).Methods("DELETE")
	r.Handle("/api/post/{POST_ID}/upvote", auth(http.HandlerFunc(postHandler.Upvote))).Methods("GET")
	r.Handle("/api/post/{POST_ID}/downvote", auth(http.HandlerFunc(postHandler.Downvote))).Methods("GET")
	r.Handle("/api/post/{POST_ID}/unvote", auth(http.HandlerFunc(postHandler.Unvote))).Methods("GET")

	log.Fatal(http.ListenAndServe(port, r))
}
