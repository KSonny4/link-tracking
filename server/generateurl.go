package server

import (
	"fmt"
	"log"
	"net/url"

	gonanoid "github.com/matoous/go-nanoid/v2"

	pb "github.com/ksonny4/link-tracking/proto"
)

var alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var baseurl = "https://links.pkubelka.cz/%s/%s"


var GetUniqueId = func() string {
	// this is syntactically differently written func so we can mock it
	id, err := gonanoid.Generate(alphabet, 8)
	if err != nil {
		log.Fatal("Could not generate id")
	}
	return id
}

func (server *Server) GenerateID(input *pb.URLGenerateRequest) string {
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
			used := server.CheckIfIdExistsInDB(id, "urls")
			for used {
				id = GetUniqueId()
				used = server.CheckIfIdExistsInDB(id, "urls")
			}
		} else if urlParams.UrlType == pb.URLType_URL_LONG {
			id = urlParams.Url
		}
	} else if pixelParams != nil {
		id = GetUniqueId()
		used := server.CheckIfIdExistsInDB(id, "pixels")
		for used {
			id = GetUniqueId()
			used = server.CheckIfIdExistsInDB(id, "pixels")
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
		return fmt.Sprintf(baseurl, "l", ID)
	} else if pixelParams != nil {
		return fmt.Sprintf(baseurl, "p", ID)
	} else {
		panic("NOT IMPLEMENTED")
	}
}

func (server *Server) GetUrlCallback(input *pb.URLGenerateRequest) (*pb.Url, error) {
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

	generatedID := server.GenerateID(input)
	generatedURL := GenerateURL(input, generatedID)

	if debug {
		log.Printf("Generated %s for %s", generatedURL, input)
	}

	server.SaveUrlToDB(generatedURL, generatedID, input)

	if debug {
		log.Printf("DB contains: %+v",server.GetTableUrlsFromDB())
	} 
	return &pb.Url{Url: generatedURL}, nil
}
