package main

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	config "github.com/010blue/jmeter-status/service/config"
	"github.com/PuerkitoBio/goquery"
	"github.com/jinzhu/now"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

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
var itemsLen = 50 // only 50 lastest items
var cacheDataPath = "../web/data/"
var tomlConfig *config.TomlConfig

// data container
type container struct {
	Websites []config.Website `json:"websites"`
	Default  int              `json:"default"`
}

// judge file exists
func fileExist(file string) bool {
	_, err := os.Stat(file)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

func notify(tomlConfig *config.TomlConfig) {
	notifyByPageduty(tomlConfig) // default by Pageduty
}

func notifyByPageduty(tomlConfig *config.TomlConfig) {
	data := []byte(`{
		"incident": {
		  "type": "incident",
		  "title": "` + tomlConfig.Notification.Pageduty.NotificationTitle + `",
		  "service": {
			"id": "` + tomlConfig.Notification.Pageduty.ServiceID + `",
			"type": "service_reference"
		  }
		}
	  }`)
	request, _ := http.NewRequest("POST", "https://api.pagerduty.com/incidents", bytes.NewBuffer(data))
	request.Header.Set("Content-type", "application/json; charset=utf-8")
	request.Header.Set("Accept", "application/vnd.pagerduty+json;version=2")
	request.Header.Set("Authorization", "Token token="+tomlConfig.Notification.Pageduty.AuthToken)
	request.Header.Set("From", tomlConfig.Notification.Pageduty.From)
	_, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Println(err)
	}
}

// get content with authorization or not
func requestWithAuthorization(url string, authUser string, authPassword string) (response *http.Response, err error) {
	client := &http.Client{}
	request, _ := http.NewRequest("GET", url, nil)
	if authUser != "" {
		authorization := []byte(authUser + ":" + authPassword)
		request.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString(authorization))
	}
	return client.Do(request)
}

// get day tasks from DB
func getDayTasks(date string, websiteID int, tomlConfig *config.TomlConfig) (tasks []config.Task, err error) {
	beginTime, _ := now.Parse(date)
	beginTime = beginTime.UTC()
	endTime := beginTime.AddDate(0, 0, 1)

	// get tasks from DB
	db := config.InitDB(tomlConfig)
	defer db.Close()

	rows, err := db.Query("SELECT id,website_id,file,api_count,api_error_count,api_error_rate,executed_at FROM tasks WHERE website_id=? AND executed_at>=? AND executed_at<?", websiteID, beginTime, endTime)

	if err != nil {
		return tasks, err
	}

	for rows.Next() {
		var id int
		var websiteID int
		var file string
		var apiCount int
		var apiErrorCount int
		var apiErrorRate float32
		var executedAt time.Time
		if rowErr := rows.Scan(&id, &websiteID, &file, &apiCount, &apiErrorCount, &apiErrorRate, &executedAt); rowErr != nil {
			log.Println(rowErr)
		}

		task := config.Task{
			ID:         id,
			WebsiteID:  websiteID,
			ExecutedAt: executedAt,
			Count:      apiCount,
			ErrCount:   apiErrorCount,
			ErrRate:    apiErrorRate,
			File:       file,
		}

		tasks = append(tasks, task)
	}

	return tasks, err
}

// get day's status data
func getDayStatus(date string, websiteID int, tomlConfig *config.TomlConfig) (dayStatus config.DayStatus, err error) {
	// get tasks from DB
	db := config.InitDB(tomlConfig)
	defer db.Close()

	tasks, err := getDayTasks(date, websiteID, tomlConfig)
	if err != nil {
		return dayStatus, err
	}

	count := 1
	errCount := 0
	for _, task := range tasks {
		count += task.Count
		errCount += task.ErrCount
	}
	errRate := float32(100 * errCount / count)

	dayStatus = config.DayStatus{
		Date:     date,
		Count:    count,
		ErrCount: errCount,
		ErrRate:  errRate,
	}

	return dayStatus, err
}

// generate json for today
func generateTodayStatus(dataContainer *container, tomlConfig *config.TomlConfig) (err error) {
	todayDate := time.Now().Format("2006-01-02")
	// find tasks
	for k, website := range dataContainer.Websites {
		dataContainer.Websites[k].Days = []config.DayStatus{}
		dataContainer.Websites[k].Data, err = getDayTasks(todayDate, website.ID, tomlConfig)
	}

	containerJSON, _ := json.Marshal(dataContainer)
	// save data to file
	ioutil.WriteFile(cacheDataPath+"today.json", containerJSON, 0644)

	return err
}

func generateWeekStatus(dataContainer *container, tomlConfig *config.TomlConfig) (err error) {
	beginTime := now.BeginningOfWeek()
	endTime := now.EndOfWeek()

	// find day status
	for k, website := range dataContainer.Websites {
		statuses := []config.DayStatus{}
		for i := 0; i < 7; i++ {
			dateTime := beginTime.AddDate(0, 0, i)
			if dateTime.Unix() >= endTime.Unix() {
				break
			}
			date := dateTime.Format("2006-01-02")
			status, _ := getDayStatus(date, website.ID, tomlConfig)
			statuses = append(statuses, status)
		}

		dataContainer.Websites[k].Data = []config.Task{}
		dataContainer.Websites[k].Days = statuses
	}

	containerJSON, _ := json.Marshal(dataContainer)
	// save data to file
	ioutil.WriteFile(cacheDataPath+"week.json", containerJSON, 0644)

	return err
}

func generateMonthStatus(dataContainer *container, tomlConfig *config.TomlConfig) (err error) {
	beginTime := now.BeginningOfMonth()
	endTime := now.EndOfMonth()

	// find day status
	for k, website := range dataContainer.Websites {
		statuses := []config.DayStatus{}
		for i := 0; i < 31; i++ {
			dateTime := beginTime.AddDate(0, 0, i)
			if dateTime.Unix() >= endTime.Unix() {
				break
			}
			date := dateTime.Format("2006-01-02")
			status, _ := getDayStatus(date, website.ID, tomlConfig)
			statuses = append(statuses, status)
		}

		dataContainer.Websites[k].Data = []config.Task{}
		dataContainer.Websites[k].Days = statuses
	}

	containerJSON, _ := json.Marshal(dataContainer)
	// save data to file
	ioutil.WriteFile(cacheDataPath+"month.json", containerJSON, 0644)

	return err
}

func main() {
	// config
	dataContainer := new(container)

	tomlConfig, configErr := config.InitConfig()
	if configErr != nil {
		log.Fatal(configErr)
	}

	dataContainer.Websites = tomlConfig.Websites
	dataContainer.Default = 0
	if tomlConfig.Rows > 0 {
		itemsLen = tomlConfig.Rows
	}

	if tomlConfig.Datapath != "" {
		cacheDataPath = tomlConfig.Datapath
	}

	hasNotified := false // only notify once

	// /config
	// fetch JMeter by index page
	for websiteK, website := range dataContainer.Websites {
		log.Println(website.Name + " fetching")

		rootPath := website.URL
		indexRes, indexErr := requestWithAuthorization(rootPath, website.AuthUser, website.AuthPassword)

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
		tasks := []config.Task{}

		linksLen := len(links)

		if len(links) > itemsLen {
			links = links[linksLen-itemsLen : linksLen]
		}

		for linkKey, link := range links {
			apiTask := config.Task{
				WebsiteID:  website.ID,
				ExecutedAt: time.Now().UTC(),
				Count:      0,
				ErrCount:   0,
				ErrRate:    0,
				File:       "",
			}

			apiTask.File = link
			var linkContent []byte
			var linkContentErr error
			// if file exists in cache, read cache or request remote
			if fileExist(cacheDataPath + apiTask.File) {
				linkContent, linkContentErr = ioutil.ReadFile(cacheDataPath + apiTask.File)
			}
			if linkContent == nil {
				// get content by remote
				linkRes, linkErr := requestWithAuthorization(rootPath+"/"+apiTask.File, website.AuthUser, website.AuthPassword)
				if linkErr != nil {
					continue
				}
				linkContent, linkContentErr = ioutil.ReadAll(linkRes.Body)
				if linkContentErr != nil {
					continue
				}
				// cache content
				ioutil.WriteFile(cacheDataPath+apiTask.File, linkContent, 0644)
			}

			content := string(linkContent)
			csvReader := csv.NewReader(strings.NewReader(content))

			// count error rate
			apiTask.Count = 0
			apiTask.ErrCount = 0
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
					timestamp, timestampErr := strconv.Atoi(csvRow[0])
					createdAt := time.Now().UTC()
					if timestampErr == nil {
						createdAt = time.Unix(int64(timestamp/1000), 0).UTC()
					}
					// time default from the first row's timestamp
					apiTask.ExecutedAt = createdAt
				}

				apiTask.Count++

				success := csvRow[7]
				if success != "true" {
					apiTask.ErrCount++
				}
			}

			if apiTask.Count == 0 {
				apiTask.Count = 1
			}

			apiTask.ErrRate = float32(100 * apiTask.ErrCount / apiTask.Count)

			// Store to db
			config.SyncTaskToDB(&apiTask, tomlConfig)
			tasks = append(tasks, apiTask)

			// the last task trigger notification
			if linkKey == len(links)-1 && !hasNotified && apiTask.ErrCount >= tomlConfig.Notification.ShouldNotifyErrorNum {
				hasNotified = true
				notify(tomlConfig)
			}
		}

		dataContainer.Websites[websiteK].Data = tasks
		log.Println(website.Name + " fetched")
	}

	generateTodayStatus(dataContainer, tomlConfig)
	generateWeekStatus(dataContainer, tomlConfig)
	generateMonthStatus(dataContainer, tomlConfig)
}
