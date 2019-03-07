package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/PuerkitoBio/goquery"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type status struct {
	Timestamp string `json:"timestamp"`
	Count     int    `json:"count"`
	ErrCount  int    `json:"errCount"`
	ErrRate   int    `json:"errRate"`
	File      string `json:"file"`
}

// Jtl file content
// type apirequest struct {
// 	timeStamp       string
// 	elapsed         string
// 	label           string
// 	responseCode    string
// 	responseMessage string
// 	threadName      string
// 	dataType        string
// 	success         string
// 	failureMessage  string
// 	bytes           string
// 	sentBytes       string
// 	grpThreads      string
// 	all             string
// 	Threads         string
// 	URL             string
// 	Filename        string
// 	Latency         string
// 	Encoding        string
// 	SampleCount     string
// 	ErrorCount      string
// 	Hostname        string
// 	IdleTime        string
// 	Connect         string
// }

var dataContainer container
var itemsLen = 24 // only show 2 lastest items
var configFile = "./config.toml"
var cacheDataPath = "../web/data/"
var config tomlConfig

// JMeter website
type website struct {
	Name          string   `json:"name"`
	URL           string   `json:"url"`
	Authorization string   // for 401
	Data          []status `json:"data"`
}

// toml config
type tomlConfig struct {
	Title    string
	Websites []website
	Datapath string
	Rows     int
}

// data container
type container struct {
	Websites []website `json:"websites"`
	Default  int       `json:"default"`
}

// judge file exists
func fileExist(file string) bool {
	_, err := os.Stat(file)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

// get content with authorization or not
func requestWithAuthorization(url string, authorization string) (response *http.Response, err error) {
	client := &http.Client{}
	request, _ := http.NewRequest("GET", url, nil)
	if authorization != "" {
		request.Header.Add("Authorization", authorization)
	}
	return client.Do(request)
}

func main() {
	// config
	dataContainer := new(container)

	_, configErr := toml.DecodeFile(configFile, &config)
	if configErr != nil {
		log.Fatal(configErr)
	}

	dataContainer.Websites = config.Websites
	dataContainer.Default = 0
	if config.Rows > 0 {
		itemsLen = config.Rows
	}

	if config.Datapath != "" {
		cacheDataPath = config.Datapath
	}

	// /config
	// fetch JMeter by index page
	for websiteK, website := range dataContainer.Websites {
		fmt.Println(time.Now().String() + ": " + website.Name + " fetching")

		rootPath := website.URL
		indexRes, indexErr := requestWithAuthorization(rootPath, website.Authorization)

		if indexErr != nil {
			log.Fatal(indexErr)
		}

		// Load index HTML document
		indexDoc, indexDocErr := goquery.NewDocumentFromReader(indexRes.Body)
		if indexDocErr != nil {
			log.Fatal(indexDocErr)
		}

		links := []string{}
		indexDoc.Find("a").Each(func(i int, s *goquery.Selection) {
			link, linkExists := s.Attr("href")
			if linkExists && strings.HasSuffix(link, ".jtl") {
				links = append(links, link)
			}
		})

		// Parse data
		statuses := []status{}

		linksLen := len(links)

		if len(links) > itemsLen {
			links = links[linksLen-itemsLen : linksLen]
		}

		for _, link := range links {
			linkAPIStatus := status{
				Timestamp: "",
				Count:     0,
				ErrCount:  0,
				ErrRate:   0,
				File:      "",
			}

			linkAPIStatus.File = link
			var linkContent []byte
			var linkContentErr error
			// if file exists in cache, read cache or request remote
			if fileExist("./cache/" + linkAPIStatus.File) {
				linkContent, linkContentErr = ioutil.ReadFile(cacheDataPath + linkAPIStatus.File)
			}
			if linkContent == nil {
				// get content by remote
				linkRes, linkErr := requestWithAuthorization(rootPath+"/"+linkAPIStatus.File, website.Authorization)
				if linkErr != nil {
					continue
				}
				linkContent, linkContentErr = ioutil.ReadAll(linkRes.Body)
				if linkContentErr != nil {
					continue
				}
				// cache content
				ioutil.WriteFile(cacheDataPath+linkAPIStatus.File, linkContent, 0644)
			}

			content := string(linkContent)
			csvReader := csv.NewReader(strings.NewReader(content))

			// count error rate
			linkAPIStatus.Count = 0
			linkAPIStatus.ErrCount = 0
			i := 0
			for {
				csvRow, csvRowErr := csvReader.Read()
				if csvRowErr == io.EOF {
					break
				}

				if len(csvRow) < 10 {
					// pass if not enough items
					continue
				}

				i++
				if i == 1 {
					continue
				} else if i == 2 {
					// time default from the first row's timestamp
					linkAPIStatus.Timestamp = csvRow[0]
				}

				linkAPIStatus.Count++

				success := csvRow[7]
				if success != "true" {
					linkAPIStatus.ErrCount++
				}
			}

			if linkAPIStatus.Count == 0 {
				linkAPIStatus.Count = 1
			}

			linkAPIStatus.ErrRate = 100 * linkAPIStatus.ErrCount / linkAPIStatus.Count
			statuses = append(statuses, linkAPIStatus)
		}

		dataContainer.Websites[websiteK].Data = statuses
		fmt.Println(time.Now().String() + ": " + website.Name + " fetched")
	}
	containerJSON, _ := json.Marshal(dataContainer)
	// save data to file
	ioutil.WriteFile(cacheDataPath+"default.json", containerJSON, 0644)
}
