package main_test

import (
	"fmt"
	"log"
	//	"log"

	"os"
	"sync"
	"testing"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/stretchr/testify/suite"

	"database/sql"
	"time"

	tracker "github.com/ksonny4/tracked-url-generator"

	"github.com/golang/protobuf/proto"
	pb "github.com/ksonny4/tracked-url-generator/generated"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"bou.ke/monkey"
)

type TestSuite struct {
	suite.Suite
}

func (suite *TestSuite) SetupTest() {

	db_name, _ := gonanoid.New()
	db_filename := fmt.Sprintf("./tmp/%s_test.db", db_name)
	log.Println(db_filename)

	file, err := os.Create(db_name)
	if err != nil {
		panic(err.Error())
	}
	file.Close()

	tracker.DB, _ = sql.Open("sqlite3", db_filename)
}

func (s *TestSuite) TearDownTest() {
	tracker.DB.Close()
}

func GetRowByID(id string, rows []tracker.UrlRecord) (tracker.UrlRecord, error) {
	for _, r := range rows {
		if id == r.Id {
			return r, nil
		}
	}
	return tracker.UrlRecord{}, fmt.Errorf("Could not find row by ID")
}

func TestExampleTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (suite *TestSuite) TestGenerateE2E() {
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

	for _, params := range parameters {
		// Mock GetUniqueId
		tracker.GetUniqueId = func() string {
			return params.id
		}

		// Actual test
		tracker.CreateTableIfNotExists()

		input := pb.UrlParams{Url: params.url, Email: proto.String(params.email), Username: proto.String(params.username)}
		url_result, _ := tracker.GetUrl(&input, tracker.ShortURL)

		assert.NotNil(suite.T(), url_result)
		rows := tracker.GetTableUrls()
		record, err := GetRowByID(params.id, rows)
		assert.Nil(suite.T(), err)

		assert.Equal(suite.T(), params.id, record.Id)
		assert.Equal(suite.T(), params.url, record.Url)
		assert.Equal(suite.T(), params.email, record.Email)
		assert.Equal(suite.T(), params.username, record.Username)
		assert.Equal(suite.T(), 0, record.Hits)
		assert.Equal(suite.T(), wayback, record.Created)
		assert.Equal(suite.T(), wayback, record.LastModified)
		assert.Equal(suite.T(), tracker.ShortURL, record.UrlType)
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

	for _, params := range parameters {
		tracker.CreateTableIfNotExists()

		input := pb.UrlParams{Url: params.url, Email: proto.String(""), Username: proto.String("")}
		_, err := tracker.GetUrl(&input, tracker.ShortURL)

		if params.expectedResult {
			assert.Nil(suite.T(), err)
		} else {
			assert.NotNil(suite.T(), err)
		}
	}
}


func (suite *TestSuite) TestParallelGetURL() {
	numberOfGoroutines := 1000

	assert.NotNil(suite.T(), tracker.DB)

	tracker.CreateTableIfNotExists()
	
	input := pb.UrlParams{Url: "https://www.example.com", Email: proto.String(""), Username: proto.String("")}
	var wg sync.WaitGroup
	for i := 0; i < numberOfGoroutines; i++ {		
		wg.Add(1)
		go func() {			
			tracker.GetUrl(&input, tracker.ShortURL)
			wg.Done()
		}()
	}
	wg.Wait()

	rows := tracker.GetTableUrls()

	assert.Equal(suite.T(), numberOfGoroutines, len(rows))

}
