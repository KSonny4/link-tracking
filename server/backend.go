package server

import (
	"database/sql"

	"context"
	"log"
	"os"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"

	pb "github.com/ksonny4/link-tracking/proto"
)

// server is used to implement helloworld.GreeterServer.
type Server struct {
	pb.UnimplementedTrackerServer
	DB *sql.DB
	Mu *sync.RWMutex
}

type URLType int

type UrlRecord struct {
	Id           string
	Url          string
	Email        string
	Username     string
	Hits         int
	Created      time.Time
	LastModified time.Time
	UrlType      string
	Note         string
}

type HitRecord struct {
	Id        string
	Url       string
	Email     string
	Username  string
	HitDate   time.Time
	UrlType   string
	IPAddress string
	Note      string
}

type PixelRecord struct {
	Id           string
	Url          string
	Email        string
	Username     string
	Hits         int
	Created      time.Time
	LastModified time.Time
	Note         string
}

var debug bool = true
var PASSWORD string = os.Getenv("AUTH_PASSWORD")

func (server *Server) GetUrl(ctx context.Context, input *pb.URLGenerateRequest) (*pb.Url, error) {	
	server.Mu.RLock()
	defer server.Mu.RUnlock()
	return server.GetUrlCallback(input)
}

// New initializes a new Backend struct.
func New() *Server {
	db_name := os.Getenv("DB_PATH")
	if db_name == "" {
		db_name = "sqlite-database.db"
	}

	if !fileExists(db_name) {
		file, err := os.Create(db_name)
		if err != nil {
			panic(err.Error())
		}
		file.Close()
		log.Printf("Could not find %s, created.", db_name)
	}

	var err error
	var DB *sql.DB
	DB, err = sql.Open("sqlite3", db_name)
	if err != nil || DB == nil {
		log.Fatal("Couldn't open DB.")
	}
	DB.SetMaxOpenConns(1) //https://github.com/mattn/go-sqlite3/issues/274

	s := &Server{
		DB: DB,
		Mu: &sync.RWMutex{},
	}

	s.CreateTableIfNotExists()

	return s
}
