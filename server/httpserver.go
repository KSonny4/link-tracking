package server

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"log"

	"github.com/gin-gonic/gin"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

func (server *Server) StartHTTPServer() {
	router := gin.Default()

	if PASSWORD == "" {
		PASSWORD, _ = gonanoid.New(8)
		fmt.Println("Password was not set up, using ", PASSWORD)
	}

	router.GET("/l/GetTableUrlHits", func(c *gin.Context) {

		pass := c.GetHeader("Authorization")
		if pass != PASSWORD {
			c.Data(http.StatusUnauthorized, "text/html; charset=utf-8", []byte("<html><body><h1>401 UNAUTHORIZED</h1></body></html>"))
			str := fmt.Sprintf("Bad pasword for showhits %s", pass)
			c.Error(errors.New(str))
			return
		}

		table := server.GetTableUrlHitsFromDB()
		var str string
		for _, i := range *table {
			str += fmt.Sprintf("%+v\n", i)
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(str))
	})

	router.GET("/l/GetTableUrls", func(c *gin.Context) {

		pass := c.GetHeader("Authorization")
		if pass != PASSWORD {
			c.Data(http.StatusUnauthorized, "text/html; charset=utf-8", []byte("<html><body><h1>401 UNAUTHORIZED</h1></body></html>"))
			str := fmt.Sprintf("Bad pasword for showhits %s", pass)
			c.Error(errors.New(str))
			return
		}

		table := server.GetTableUrlsFromDB()
		var str string
		for _, i := range *table {
			str += fmt.Sprintf("%+v\n", i)
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(str))
	})

	router.GET("/l/:id", func(c *gin.Context) {
		userIP := c.ClientIP()
		Path := c.Request.URL.Path
		if debug {
			log.Printf("User IP: %s, Path: %s", userIP, Path)
		}

		shortenedID := c.Param("id")
		println(shortenedID)

		// Get Url for id

		db_result := server.GetUrlForIDFromDB(shortenedID)
		if (db_result == UrlRecord{}) {
			if debug {
				log.Println(server.GetTableUrlsFromDB())
			}
			c.Data(http.StatusNotFound, "text/html; charset=utf-8", []byte("<html><body><h1>404 NOT FOUND</h1></body></html>"))
			str := fmt.Sprintf("Failed to found shortenedID in DB %s", shortenedID)
			c.Error(errors.New(str))
			return
		}

		server.IncrementHitDB(shortenedID, "urls")
		server.SaveUrlHitToDB(db_result, shortenedID, userIP)

		if debug {
			log.Printf("DB Result: %v", db_result)
		}

		UrlFromDB := db_result.Url
		println(UrlFromDB)
		if debug {
			log.Printf("UrlFromDB: %s", UrlFromDB)
		}

		location := url.URL{Path: UrlFromDB}

		c.Redirect(http.StatusMovedPermanently, location.RequestURI())
	})
	
	//router.GET("/p/:id", PixelHit)

	router.Run(":3000")
}
