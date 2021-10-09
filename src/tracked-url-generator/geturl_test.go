package main_test

import (
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"github.com/stretchr/testify/suite"
	"database/sql"
	"time"
	tracker "github.com/ksonny4/tracked-url-generator"
	"bou.ke/monkey"
	"github.com/golang/protobuf/proto"
	pb "github.com/ksonny4/tracked-url-generator/generated"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

type TestSuite struct {
	suite.Suite
}

var db_filename string = "sqlite_test.db"

func (suite *TestSuite) SetupTest() {

	log.Println(db_filename)

	file, err := os.Create(db_filename)
	if err != nil {
		panic(err.Error())
	}
	file.Close()

	tracker.DB, _ = sql.Open("sqlite3", db_filename)
	tracker.DB.SetMaxOpenConns(1)
}

func (s *TestSuite) TearDownTest() {
	tracker.DB.Close()
	os.Remove(db_filename)
}

func GetRowByID(id string, rows []tracker.UrlRecord) (tracker.UrlRecord, error) {
	for _, r := range rows {
		if id == r.Id {
			return r, nil
		}
	}
	return tracker.UrlRecord{}, fmt.Errorf("Could not find row by ID")
}

func GetRowByIDPixels(id string, rows []tracker.PixelRecord) (tracker.PixelRecord, error) {
	for _, r := range rows {
		if id == r.Id {
			return r, nil
		}
	}
	return tracker.PixelRecord{}, fmt.Errorf("Could not find row by ID")
}

func TestExampleTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (suite *TestSuite) TestGenerateE2ELinks() {
	// Prepare for test
	wayback := time.Date(1974, time.May, 19, 1, 2, 3, 0, time.UTC)
	patch := monkey.Patch(time.Now, func() time.Time { return wayback })
	defer patch.Unpatch()

	assert.NotNil(suite.T(), tracker.DB)

	parameters := []struct {
		email    string
		username string
		url      string
		id       string
	}{
		{email: "user@email.com", username: "TestUser", url: "https://www.example.com", id: "1"},
		{email: "user@email.com", url: "https://www.example2.com", id: "2"},
		{username: "TestUser", url: "https://www.example3.com", id: "3"},
	}

	// Save GetUniqueId meant to be mocked
	GetUniqueIdOriginal := tracker.GetUniqueId

	tracker.CreateTableIfNotExists()

	for _, params := range parameters {
		// Mock GetUniqueId
		tracker.GetUniqueId = func() string {
			return params.id
		}

		// Actual test
		
		input := &pb.URLGenerateRequest{Request: &pb.URLGenerateRequest_UrlParams{
			UrlParams: &pb.UrlParams{Url: params.url, UrlType: pb.URLType_URL_SHORT, Email: proto.String(params.email), Username: proto.String(params.username)},
		}}
		
		url_result, _ := tracker.GetUrl(input)

		assert.NotNil(suite.T(), url_result)
		rows := tracker.GetTableUrls()
		record, err := GetRowByID(params.id, *rows)
		assert.Nil(suite.T(), err)

		assert.Equal(suite.T(), params.id, record.Id)
		assert.Equal(suite.T(), params.url, record.Url)
		assert.Equal(suite.T(), params.email, record.Email)
		assert.Equal(suite.T(), params.username, record.Username)
		assert.Equal(suite.T(), 0, record.Hits)
		assert.Equal(suite.T(), wayback, record.Created)
		assert.Equal(suite.T(), wayback, record.LastModified)
		assert.Equal(suite.T(), pb.URLType_URL_SHORT.String(), record.UrlType)
	}

	// Put original GetUniqueId back
	tracker.GetUniqueId = GetUniqueIdOriginal
}


func (suite *TestSuite) TestGenerateE2EPixels() {
	// Prepare for test
	wayback := time.Date(1974, time.May, 19, 1, 2, 3, 0, time.UTC)
	patch := monkey.Patch(time.Now, func() time.Time { return wayback })
	defer patch.Unpatch()

	assert.NotNil(suite.T(), tracker.DB)

	parameters := []struct {
		note     string
		email    string
		username string
		url      string
		id       string
	}{
		{note: "Note", email: "user@email.com", username: "TestUser", url: "https://www.example.com", id: "1"},
		{note: "Note",email: "user@email.com", url: "https://www.example2.com", id: "2"},
		{note: "Note",username: "TestUser", url: "https://www.example3.com", id: "3"},
	}

	// Save GetUniqueId meant to be mocked
	GetUniqueIdOriginal := tracker.GetUniqueId

	tracker.CreateTableIfNotExists()

	for _, params := range parameters {
		// Mock GetUniqueId
		tracker.GetUniqueId = func() string {
			return params.id
		}

		// Actual test
		
		input := &pb.URLGenerateRequest{Request: &pb.URLGenerateRequest_PixelParams{
			PixelParams: &pb.PixelParams{Note: params.note, Url: proto.String(params.url), Email: proto.String(params.email), Username: proto.String(params.username)},
		}}
		
		url_result, _ := tracker.GetUrl(input)

		assert.NotNil(suite.T(), url_result)
		rows := tracker.GetTablePixels()
		record, err := GetRowByIDPixels(params.id, *rows)
		assert.Nil(suite.T(), err)

		assert.Equal(suite.T(), params.id, record.Id)
		assert.Equal(suite.T(), params.url, record.Url)
		assert.Equal(suite.T(), params.email, record.Email)
		assert.Equal(suite.T(), params.username, record.Username)
		assert.Equal(suite.T(), 0, record.Hits)
		assert.Equal(suite.T(), wayback, record.Created)
		assert.Equal(suite.T(), wayback, record.LastModified)
		assert.Equal(suite.T(), params.note, record.Note)
	}

	// Put original GetUniqueId back
	tracker.GetUniqueId = GetUniqueIdOriginal
}

func (suite *TestSuite) TestURLValidation() {
	assert.NotNil(suite.T(), tracker.DB)

	parameters := []struct {
		url            string
		expectedResult bool
	}{
		{url: "", expectedResult: false},
		{url: "SELECT UserId, Name, Password FROM Users WHERE UserId = 105 or 1=1;", expectedResult: false},
		{url: "https://www.google.com", expectedResult: true},
		{url: "http://www.google.com", expectedResult: true},
	}

	tracker.CreateTableIfNotExists()

	for _, params := range parameters {

		
		input := &pb.URLGenerateRequest{Request: &pb.URLGenerateRequest_UrlParams{
			UrlParams: &pb.UrlParams{Url: params.url, UrlType: pb.URLType_URL_SHORT, Email: proto.String(""), Username: proto.String("")},
		}}
		
		_, err := tracker.GetUrl(input)

		if params.expectedResult {
			assert.Nil(suite.T(), err)
		} else {
			assert.NotNil(suite.T(), err)
		}
	}
}

func (suite *TestSuite) TestParallelGetURL() {
	numberOfGoroutines := 100

	assert.NotNil(suite.T(), tracker.DB)

	tracker.CreateTableIfNotExists()


	input := &pb.URLGenerateRequest{Request: &pb.URLGenerateRequest_UrlParams{
		UrlParams: &pb.UrlParams{Url: "https://www.example.com", UrlType: pb.URLType_URL_SHORT, Email: proto.String(""), Username: proto.String("")},
	}}
	
	var wg sync.WaitGroup
	for i := 0; i < numberOfGoroutines; i++ {
		wg.Add(1)
		go func() {
			tracker.GetUrl(input)
			wg.Done()
		}()
	}
	wg.Wait()

	rows := tracker.GetTableUrls()

	assert.Equal(suite.T(), numberOfGoroutines, len(*rows))

}
