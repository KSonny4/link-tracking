package server

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"time"

	pb "github.com/ksonny4/link-tracking/proto"
)

func (server *Server) SaveUrlToDB(url string, id string, input *pb.URLGenerateRequest) {

	if debug {
		log.Printf("Inserting url record %+v", input)
	}

	urlParams := input.GetUrlParams()
	pixelParams := input.GetPixelParams()
	var insertSQLQuery string

	if urlParams != nil {
		insertSQLQuery = `INSERT INTO urls(id, url, email, username, hits, created, last_modified, url_type, note) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	} else if pixelParams != nil {
		insertSQLQuery = `INSERT INTO pixels(id, url, email, username, hits, created, last_modified, note) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	} else {
		panic("NOT IMPLEMENTED")
	}

	statement, err := server.DB.Prepare(insertSQLQuery)
	if err != nil {
		panic(err.Error())
	}

	date := time.Now().Format(time.RFC3339)
	fmt.Println(date)

	if urlParams != nil {
		_, err = statement.Exec(id, urlParams.Url, urlParams.Email, urlParams.Username, 0, date, date, urlParams.UrlType.String(), urlParams.Note)
	} else if pixelParams != nil {
		_, err = statement.Exec(id, pixelParams.Url, pixelParams.Email, pixelParams.Username, 0, date, date, pixelParams.Note)
	} else {
		panic("NOT IMPLEMENTED")
	}

	if err != nil {
		panic(err.Error())
	}
}

func (server *Server) SaveUrlHitToDB(url UrlRecord, id, ipAddr string) {

	if debug {
		log.Printf("Inserting url hit %+v", url)
	}

	insertSQLQuery := `INSERT INTO url_hits(id, url, email, username, hit_date, url_type, ip_address, note) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	statement, err := server.DB.Prepare(insertSQLQuery)
	if err != nil {
		panic(err.Error())
	}

	date := time.Now().Format(time.RFC3339)

	_, err = statement.Exec(id, url.Url, url.Email, url.Username, date, url.UrlType, ipAddr, url.Note)

	if err != nil {
		panic(err.Error())
	}
}

func (server *Server) IncrementHitDB(id, table string) {

	if debug {
		log.Printf("Incrementing hit for %s", id)
	}

	var insertSQLQuery string

	if table == "urls" {
		insertSQLQuery = `UPDATE urls	SET hits = hits + 1, last_modified = ? WHERE id = ?;`
	} else if table == "pixels" {
		insertSQLQuery = `UPDATE pixels	SET hits = hits + 1, last_modified = ? WHERE id = ?;`
	} else {
		panic("NOT IMPLEMENTED")
	}

	statement, err := server.DB.Prepare(insertSQLQuery)
	if err != nil {
		panic(err.Error())
	}

	date := time.Now().Format(time.RFC3339)
	_, err = statement.Exec(date, id)

	if err != nil {
		panic(err.Error())
	}
}

func (server *Server) CheckIfIdExistsInDB(id, table string) bool {
	if debug {
		log.Println("In CheckIfIdExistsInDB method")
	}

	tx, err := server.DB.Begin()
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
		panic(err)
	}
	defer stmt.Close()

	row, err := stmt.Query(id)
	if err != nil {
		panic(err)
	}
	defer row.Close()

	if err := tx.Commit(); err != nil {
		panic(err)
	}

	return row.Next()
}

func UnwrapNullString(str sql.NullString) string {
	s := ""
	if str.Valid {
		s = str.String
	}
	return s
}

func (server *Server) GetUrlForIDFromDB(requestId string) UrlRecord {
	
	if debug {
		log.Println("In GetUrlForIDFromDB method, requesting ", requestId)
	}

	var stmt *sql.Stmt
	stmt, err := server.DB.Prepare("SELECT * FROM urls WHERE id = ?")

	if err != nil {
		panic(err)
	}

	var (
		// SQL columns
		id           string
		url          string
		email        sql.NullString
		username     sql.NullString
		hits         int
		created      string
		lastModified string
		urlType      string
		note         sql.NullString
	)

	err = stmt.QueryRow(requestId).Scan(&id, &url, &email, &username, &hits, &created, &lastModified, &urlType, &note)
	if err != nil {
		return UrlRecord{}
	}

	createdTime, err := time.Parse(time.RFC3339, created)
	if err != nil {
		panic(fmt.Sprintf("There was error when parsing created date from DB, date: '%s'", created))
	}
	lastModifiedTime, err := time.Parse(time.RFC3339, lastModified)
	if err != nil {
		panic(fmt.Sprintf("There was error when parsing lastmodified date from DB, date: '%s'", lastModified))
	}

	return UrlRecord{
		Id:           id,
		Url:          url,
		Email:        UnwrapNullString(email),
		Username:     UnwrapNullString(username),
		Hits:         hits,
		Created:      createdTime,
		LastModified: lastModifiedTime,
		UrlType:      urlType,
		Note:         UnwrapNullString(note),
	}

}

func (server *Server) GetTableUrlsFromDB() *[]UrlRecord {
	row, err := server.DB.Query("SELECT * FROM urls ORDER BY hits DESC LIMIT 1000")
	if err != nil {
		panic(err)
	}
	defer row.Close()

	var urlRecords []UrlRecord

	for row.Next() {
		var (
			urlRecord UrlRecord
			// SQL columns
			id           string
			url          string
			email        sql.NullString
			username     sql.NullString
			hits         int
			created      string
			lastModified string
			urlType      string
			note         sql.NullString
		)
		row.Scan(&id, &url, &email, &username, &hits, &created, &lastModified, &urlType, &note)

		createdTime, err := time.Parse(time.RFC3339, created)
		if err != nil {
			panic(fmt.Sprintf("There was error when parsing created date from DB, date: '%s'", created))
		}
		lastModifiedTime, err := time.Parse(time.RFC3339, lastModified)
		if err != nil {
			panic(fmt.Sprintf("There was error when parsing lastmodified date from DB, date: '%s'", lastModified))
		}

		urlRecord = UrlRecord{
			Id:           id,
			Url:          url,
			Email:        UnwrapNullString(email),
			Username:     UnwrapNullString(username),
			Hits:         hits,
			Created:      createdTime,
			LastModified: lastModifiedTime,
			UrlType:      urlType,
			Note:         UnwrapNullString(note),
		}
		urlRecords = append(urlRecords, urlRecord)
	}
	return &urlRecords
}

func (server *Server) GetTableUrlHitsFromDB() *[]HitRecord {
	row, err := server.DB.Query("SELECT * FROM url_hits ORDER BY hit_date DESC LIMIT 1000")
	if err != nil {
		panic(err)
	}
	defer row.Close()

	var hitRecords []HitRecord

	for row.Next() {
		var (
			hitRecord HitRecord
			// SQL columns
			id              string
			url             string
			email           sql.NullString
			username        sql.NullString
			hitDate         string
			urlType         string
			IPAddressString string
			note            sql.NullString
		)
		row.Scan(&id, &url, &email, &username, &hitDate, &urlType, &IPAddressString, &note)

		HitDateConverted, err := time.Parse(time.RFC3339, hitDate)
		if err != nil {
			panic(fmt.Sprintf("There was error when parsing created date from table url_hits %s", hitDate))
		}

		hitRecord = HitRecord{
			Id:        id,
			Url:       url,
			Email:     UnwrapNullString(email),
			Username:  UnwrapNullString(username),
			HitDate:   HitDateConverted,
			UrlType:   urlType,
			IPAddress: IPAddressString,
			Note:      UnwrapNullString(note),
		}
		hitRecords = append(hitRecords, hitRecord)
	}
	return &hitRecords
}

func (server *Server) CreateTableIfNotExists() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}

	exPath := path.Join(path.Dir(filename), "../")

	log.Println("Create tables...")
	for _, table := range []string{path.Join(exPath, "sql/url-tables-links.sql"), path.Join(exPath, "sql/table-url-hits.sql"), path.Join(exPath, "sql/url-tables-pixels.sql")} {
		tableSQL, err := os.ReadFile(table)
		if debug {
			log.Println(string(tableSQL))
		}

		if err != nil {
			panic(fmt.Sprintf("Could not find SQL tables definition. %s", table))
		}

		if server.DB == nil {
			panic("DB IS NIL")
		}
		statement, err := server.DB.Prepare(string(tableSQL))
		if err != nil {
			panic(err.Error())
		}
		defer statement.Close()
		statement.Exec()
	}
	log.Println("Tables created")
}

func (server *Server) GetTablePixels() *[]PixelRecord {
	row, err := server.DB.Query("SELECT * FROM pixels")
	if err != nil {
		panic(err)
	}
	defer row.Close()

	var pixelRecords []PixelRecord

	for row.Next() {
		var (
			pixelRecord PixelRecord
			// SQL columns
			id           string
			url          string
			email        sql.NullString
			username     sql.NullString
			hits         int
			created      string
			lastModified string
			note         sql.NullString
		)
		row.Scan(&id, &url, &email, &username, &hits, &created, &lastModified, &note)

		createdTime, err := time.Parse(time.RFC3339, created)
		if err != nil {
			panic(fmt.Sprintf("There was error when parsing created date from DB %s", created))
		}
		lastModifiedTime, err := time.Parse(time.RFC3339, lastModified)
		if err != nil {
			panic(fmt.Sprintf("There was error when parsing lastmodified date from DB %s", created))
		}


		
		pixelRecord = PixelRecord{
			Id:           id,
			Url:          url,
			Email:        UnwrapNullString(email),
			Username:     UnwrapNullString(username),
			Hits:         hits,
			Created:      createdTime,
			LastModified: lastModifiedTime,
			Note:         UnwrapNullString(note),
		}
		pixelRecords = append(pixelRecords, pixelRecord)
	}
	return &pixelRecords
}

func GetUrlsForEmails(emails []string) []string {
	log.Println(emails)
	panic("Unimplemented")
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
