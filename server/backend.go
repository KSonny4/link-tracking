package server

import (
	"database/sql"

	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
	_ "github.com/mattn/go-sqlite3"

	pb "github.com/ksonny4/link-tracking/proto"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedTrackerServer
}

const (
	port = ":50051"
)

var DB *sql.DB

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

func CheckIfIdExists(id, table string) bool {
	if debug {
		log.Println("In CheckIfIdExists method")
	}

	tx, err := DB.Begin()
	defer tx.Rollback()
	if err != nil {
		panic(err)
	}

	var stmt *sql.Stmt
	if table == "urls" {
		stmt, err = tx.Prepare("SELECT id FROM urls WHERE id = ?")
	} else if table == "pixels" {
		stmt, err = tx.Prepare("SELECT id FROM pixels WHERE id = ?")
	} else {
		panic("NOT IMPLEMENTED")
	}

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

func SaveUrlToDB(url string, id string, input *pb.URLGenerateRequest) {

	if debug {
		log.Printf("Inserting url record %+v", input)
	}

	urlParams := input.GetUrlParams()
	pixelParams := input.GetPixelParams()
	var insertSQLQuery string

	if urlParams != nil {
		insertSQLQuery = `INSERT INTO urls(id, url, email, username, hits, created, last_modified, url_type) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	} else if pixelParams != nil {
		insertSQLQuery = `INSERT INTO pixels(id, url, email, username, hits, created, last_modified, note) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	} else {
		panic("NOT IMPLEMENTED")
	}

	statement, err := DB.Prepare(insertSQLQuery)
	if err != nil {
		log.Fatalln(err.Error())
	}

	date := time.Now().Format(time.RFC3339)

	if urlParams != nil {
		_, err = statement.Exec(id, urlParams.Url, urlParams.Email, urlParams.Username, 0, date, date, urlParams.UrlType.String())
	} else if pixelParams != nil {
		_, err = statement.Exec(id, pixelParams.Url, pixelParams.Email, pixelParams.Username, 0, date, date, pixelParams.Note)
	} else {
		panic("NOT IMPLEMENTED")
	}

	if err != nil {
		log.Fatalln(err.Error())
	}
}

var GetUniqueId = func() string {
	// this is syntacticly differently written func so we can mock it
	id, err := gonanoid.New(8)
	if err != nil {
		log.Fatal("Could not generate id")
	}
	return id
}

func GenerateID(input *pb.URLGenerateRequest) string {
	// space for improvement
	if debug {
		log.Println("Generating info.")
	}
	var id string
	urlParams := input.GetUrlParams()
	pixelParams := input.GetPixelParams()
	if urlParams != nil {
		if urlParams.UrlType == pb.URLType_URL_SHORT {
			id = GetUniqueId()
			used := CheckIfIdExists(id, "urls")
			for used {
				id = GetUniqueId()
				used = CheckIfIdExists(id, "urls")
			}
		} else if urlParams.UrlType == pb.URLType_URL_LONG {
			id = urlParams.Url
		}
	} else if pixelParams != nil {
		id = GetUniqueId()
		used := CheckIfIdExists(id, "pixels")
		for used {
			id = GetUniqueId()
			used = CheckIfIdExists(id, "pixels")
		}
	} else {
		panic("NOT IMPLEMENTED")
	}

	return id
}

func GenerateURL(input *pb.URLGenerateRequest, ID string) string {
	urlParams := input.GetUrlParams()
	pixelParams := input.GetPixelParams()

	if urlParams != nil {
		return fmt.Sprintf("https://links.pkubelka.cz/l/%s", ID)
	} else if pixelParams != nil {
		return fmt.Sprintf("https://links.pkubelka.cz/p/%s", ID)
	} else {
		panic("NOT IMPLEMENTED")
	}
}

func GetUrl(input *pb.URLGenerateRequest) (*pb.Url, error) {
	// UrlParams request will use URL for later redirect, validate it.
	// Other requests just use it for informational purposes.
	if input.GetUrlParams() != nil {
		if _, err := url.ParseRequestURI(input.GetUrlParams().Url); err != nil {
			return &pb.Url{}, fmt.Errorf("invalid URL. Input data: %s", input)
		}
	} else if input.GetPixelParams() != nil {
		if input.GetPixelParams().Note == "" {
			return &pb.Url{}, fmt.Errorf("Missing note in request: %s", input)
		}
	} else {
		panic("NOT IMPLEMENTED")
	}

	generatedID := GenerateID(input)
	generatedURL := GenerateURL(input, generatedID)

	if debug {
		log.Printf("Generated %s for %s", generatedURL, input)
	}

	SaveUrlToDB(generatedURL, generatedID, input)
	return &pb.Url{Url: generatedURL}, nil
}

func GetUrlsForEmails(emails []string) []string {
	// TODO Make A/b testing long a short links here when making campain
	log.Println(emails)
	panic("Unimplemented")
}

func GetTableUrls() *[]UrlRecord {
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
			urlType      string
		)
		row.Scan(&id, &url, &email, &username, &hits, &created, &lastModified, &urlType)

		createdTime, err := time.Parse(time.RFC3339, created)
		if err != nil {
			log.Fatal(fmt.Sprintf("There was error when parsing created date from DB %s", created))
		}
		lastModifiedTime, err := time.Parse(time.RFC3339, lastModified)
		if err != nil {
			log.Fatal(fmt.Sprintf("There was error when parsing lastmodified date from DB %s", created))
		}

		urlRecord = UrlRecord{
			Id:           id,
			Url:          url,
			Email:        email,
			Username:     username,
			Hits:         hits,
			Created:      createdTime,
			LastModified: lastModifiedTime,
			UrlType:      urlType,
		}
		urlRecords = append(urlRecords, urlRecord)
	}
	return &urlRecords
}

func GetTablePixels() *[]PixelRecord {
	row, err := DB.Query("SELECT * FROM pixels")
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()

	var pixelRecords []PixelRecord

	for row.Next() {
		var (
			pixelRecord PixelRecord
			// SQL columns
			id           string
			url          string
			email        string
			username     string
			hits         int
			created      string
			lastModified string
			note         string
		)
		// TOOD check this
		row.Scan(&id, &url, &email, &username, &hits, &created, &lastModified, &note)

		createdTime, err := time.Parse(time.RFC3339, created)
		if err != nil {
			log.Fatal(fmt.Sprintf("There was error when parsing created date from DB %s", created))
		}
		lastModifiedTime, err := time.Parse(time.RFC3339, lastModified)
		if err != nil {
			log.Fatal(fmt.Sprintf("There was error when parsing lastmodified date from DB %s", created))
		}

		pixelRecord = PixelRecord{
			Id:           id,
			Url:          url,
			Email:        email,
			Username:     username,
			Hits:         hits,
			Created:      createdTime,
			LastModified: lastModifiedTime,
			Note:         note,
		}
		pixelRecords = append(pixelRecords, pixelRecord)
	}
	return &pixelRecords
}

func CreateTableIfNotExists() {
	// TODO Use nev variable here

	log.Println("Create tables...")
	for _, table := range []string{"./sql/url-tables-links.sql", "./sql/url-tables-pixels.sql"} {
		tableSQL, err := os.ReadFile(table)
		if debug {
			log.Println(string(tableSQL))
		}

		if err != nil {
			log.Fatal("Could not find SQL tables definition.")
		}

		if DB == nil {
			panic("DB IS NIL")
		}
		statement, err := DB.Prepare(string(tableSQL))
		if err != nil {
			log.Fatal(err.Error())
		}
		defer statement.Close()
		statement.Exec()
	}
	log.Println("Tables created")
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func (s *server) GetUrl(ctx context.Context, in *pb.URLGenerateRequest) (*pb.Url, error) {
	// TODO authorization
	// TODO ip block
	return GetUrl(in)
}

func init_main() {

	// TODO later dont create new db, just check if file exists
	// TODO use env variable to setup db name+path
	// move this to initialize method

	db_name := "sqlite-database.db"

	if !fileExists(db_name) {
		file, err := os.Create(db_name)
		if err != nil {
			log.Fatal(err.Error())
		}
		file.Close()
		log.Printf("Could not find %s, created.", db_name)
	}

	var err error
	DB, err = sql.Open("sqlite3", db_name)
	if err != nil || DB == nil {
		log.Fatal("Couldn't open DB.")
	}
	DB.SetMaxOpenConns(1) //https://github.com/mattn/go-sqlite3/issues/274	

	CreateTableIfNotExists()

}
