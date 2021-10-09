package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"
	"net/url"
	//"net"
	"sync"
	"os"
	
	//"github.com/golang/protobuf/proto"
	pb "github.com/ksonny4/tracked-url-generator/generated"
	gonanoid "github.com/matoous/go-nanoid/v2"
	_ "github.com/mattn/go-sqlite3"
	//"google.golang.org/grpc"
)

var DB *sql.DB
var mutex = &sync.Mutex{}

//TODO https://www.alexedwards.net/blog/organising-database-access
//type Clients struct {
//  DB *sql.DB
//}

type URLType int

const (
	ShortURL URLType = iota
	LongURL
)

type UrlRecord struct {
	Id           string
	Url          string
	Email        string
	Username     string
	Hits         int
	Created      time.Time
	LastModified time.Time
	UrlType      URLType
}


var debug bool = true


func CheckIfIdExists(id string) bool {
	tx, err := DB.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Fatal(err)
	}	

	stmt, err := tx.Prepare("SELECT id FROM urls WHERE id = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	

	row, err := stmt.Query(id)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}

	return row.Next()
}

func SaveUrlToDB(url string, id string, input *pb.UrlParams, urlType URLType) {
	
	if debug{
		log.Printf("Inserting url record %+v", input)
	}
	
	// TODO test when value is missing
	insertStudentSQL := `INSERT INTO urls(id, url, email, username, hits, created, last_modified, url_type) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	statement, err := DB.Prepare(insertStudentSQL)
	if err != nil {
		log.Fatalln(err.Error())
	}

	date := time.Now().Format(time.RFC3339)
	
	_, err = statement.Exec(id, input.Url, input.Email, input.Username, 0, date, date, urlType)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

var GetUniqueId = func() string {
	// this is differently written func so we can mock it
	id, err := gonanoid.New(8)
	if err != nil {
		panic("Could not generate id")
	}
	return id
}

func GenerateID(input *pb.UrlParams, urlType URLType) string {
	var id string
	if urlType == ShortURL {
		id = GetUniqueId()
		// TODO pixel implementation, Check if in DB for p(ixel) or l(ink)
		used := CheckIfIdExists(id)
		for used {
			id = GetUniqueId()
			// TODO pixel implementation, Check if in DB for p(ixel) or l(ink)
			used = CheckIfIdExists(id)
		}
	} else {
		id = input.Url
	}
	return id
}

func GetUrl(input *pb.UrlParams, urlType URLType) (*pb.Url, error) {
	/*
		* If urlType is shortURL, returns URL address with generated shortened name (e.g. https://links.pkubelka.cz/l/a1b2c3d4)
			If urlType is LongURL, returns URL address created from original link (e.g. https://links.pkubelka.cz/l/www.example.com)
	*/

	mutex.Lock() // TODO make this better, move this somewhere closer
	_, err := url.ParseRequestURI(input.Url)
	if err != nil {
		//Might as well validate if this is proper URL address
		return &pb.Url{}, fmt.Errorf("invalid URL. Input data: %s", input)
	}

	generatedID := GenerateID(input, urlType)
	generatedURL := fmt.Sprintf("https://links.pkubelka.cz/l/%s", generatedID)
	
	if debug {
		log.Printf("Generated https://links.pkubelka.cz/l/%s for %s", generatedID, input)
	}
	
	SaveUrlToDB(generatedURL, generatedID, input, urlType)
	mutex.Unlock()
	return &pb.Url{Url: generatedURL}, nil
}

func GetUrlsForEmails(emails []string) []string {
	// Make A/b testing long a short links here
	log.Println(emails)
	panic("Unimplemented")
}

func GetTableUrls() []UrlRecord {
	row, err := DB.Query("SELECT * FROM urls")
	if err != nil {
		log.Fatal(err)
	}	
	defer row.Close()

	var urlRecords []UrlRecord

	for row.Next() {
		var (
			urlRecord UrlRecord
			// SQL columns
			id           string
			url          string
			email        string
			username     string
			hits         int
			created      string
			lastModified string
			urlType      int
		)
		row.Scan(&id, &url, &email, &username, &hits, &created, &lastModified, &urlType)

		// TOOD chytat ty errory, jinak time vraci default :(
		createdTime, _ := time.Parse(time.RFC3339, created)
		lastModifiedTime, _ := time.Parse(time.RFC3339, lastModified)

		urlRecord = UrlRecord{Id: id, Url: url, Email: email, Username: username, Hits: hits, Created: createdTime, LastModified: lastModifiedTime, UrlType: URLType(urlType)}
		urlRecords = append(urlRecords, urlRecord)
	}

	// Return pointer
	return urlRecords
}

func CreateTableIfNotExists() {
	// TODO sql this to file
	createStudentTableSQL := `CREATE TABLE IF NOT EXISTS urls (
		"id" TEXT NOT NULL PRIMARY KEY,		
		"url" TEXT,
		"email" TEXT,
		"username" TEXT,
		"hits" INTEGER,
		"created" TEXT,
		"last_modified" TEXT,
		"url_type" INTEGER
	  );`

	log.Println("Create student table...")
	statement, err := DB.Prepare(createStudentTableSQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer statement.Close()
	statement.Exec()
	log.Println("student table created")
}

func main() {

	// TODO later dont create new db, just check if file exists
	// TODO use env variable to setup db name+path
	// move this to initialize method
	log.Println("Creating sqlite-database.db...")
	file, err := os.Create("sqlite-database.db")
	if err != nil {
		panic(err.Error())
	}
	file.Close()
	log.Println("sqlite-database.db created")

	DB, err = sql.Open("sqlite3", "./sqlite-database.db")
	if err != nil || DB == nil {
		panic("Couldn't open DB.")
	}
	defer DB.Close()

	CreateTableIfNotExists()

	// TODO Run HTTP and gRPC server

}
