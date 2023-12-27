package theguardian

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	contentUrl         = "https://content.guardianapis.com/search"
	tagsUrl            = "https://content.guardianapis.com/tags"
	singleItemUrl      = "https://content.guardianapis.com/"
	OrderByNewest      = "newest"
	OrderByOldest      = "oldest"
	OrderByRelevance   = "relevance"
	ShowFieldsBody     = "body"
	ShowFieldsByline   = "byline"
	ShowFieldsHeadline = "headline"
)

type GuardianApiStats struct {
	gorm.Model
	ApiKey            string `gorm:"-"`
	LastCalled        time.Time
	PauseBetweenCalls time.Duration
	RequestCounter    int
}

func (s *GuardianApiStats) LogCall() {
	s.LastCalled = time.Now()
	s.RequestCounter++
	result := db.Save(s)
	if result.Error != nil {
		log.Println(result.Error)
	}
}

func (s *GuardianApiStats) PauseTimeElapsed() bool {
	return time.Now().After(s.LastCalled.Add(s.PauseBetweenCalls))
}

var Stats = GuardianApiStats{
	PauseBetweenCalls: 60 * time.Minute,
}

var db *gorm.DB

type Response struct {
	Content Content `json:"response"`
}

type Content struct {
	Status      string   `json:"status"`
	Total       int      `json:"total"`
	PageSize    int      `json:"pageSize"`
	CurrentPage int      `jsom:"currentPage"`
	Pages       int      `json:"pages"`
	OrderBy     string   `json:"orderBy"`
	Results     []Result `json:"results"`
}

type SingleItemResponse struct {
	Content SingleItemContent `json:"response"`
}

type SingleItemContent struct {
	Status string `json:"status"`
	Total  int    `json:"total"`
	Result Result `json:"content"`
}

type Result struct {
	Id                 string `json:"id"`
	WebTitle           string `json:"webTitle"`
	Type               string `json:"type"`
	SectionId          string `json:"sectionId"`
	SectionName        string `json:"sectionName"`
	WebPublicationDate string `json:"webPublicationDate"`
	WebUrl             string `json:"webUrl"`
	ApiUrl             string `json:"apiUrl"`
	Fields             Fields `json:"fields"`
}

type Fields struct {
	Headline string `json:"headline"`
	Body     string `json:"body"`
	Byline   string `json:"byline"`
}

type Query struct {
	Q           string `param:"q"`
	QueryFields string `param:"query-fields"`
	Id          string `param:"id"`
	Tags        string `param:"tags"`
	OrderBy     string `param:"order-by"`
	ShowFields  string `param:"show-fields"`
	Page        int    `param:"page"`
	PageSize    int    `param:"page-size"`
}

func init() {
	var err error
	db, err = gorm.Open(sqlite.Open("apistats.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	db.AutoMigrate(&GuardianApiStats{})
	result := db.FirstOrCreate(&Stats)
	if result.Error != nil {
		log.Println(result.Error)
	}
}

func (q Query) ToUrlParams() url.Values {
	params := url.Values{}
	// get a Value from Reflect to iterate over the fields
	v := reflect.ValueOf(q)

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)                    // field as a reflect.Value
		fieldType := v.Type().Field(i)         // field as a reflect.StructField
		fieldTag := fieldType.Tag.Get("param") // field tag
		if fieldTag != "" && !field.IsZero() {
			var valueStr string
			switch field.Kind() {
			case reflect.String:
				valueStr = field.String()
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				valueStr = fmt.Sprintf("%d", field.Int())
			case reflect.Float32, reflect.Float64:
				valueStr = fmt.Sprintf("%f", field.Float())
			case reflect.Bool:
				valueStr = fmt.Sprintf("%t", field.Bool())
			// Add cases for other types as needed
			default:
				log.Printf("ToUrlParams: %v not supported", field.Kind())
				continue
			}
			params.Add(fieldTag, valueStr)
		}
	}
	params.Add("api-key", Stats.ApiKey)
	//log.Printf("params: %v", params)
	return params
}

func GetContent(query Query) (Content, error) {
	if !Stats.PauseTimeElapsed() {
		return Content{}, fmt.Errorf("not enough time between updates - last update: %v", Stats.LastCalled)
	}
	params := query.ToUrlParams()
	url := contentUrl + "?" + params.Encode()
	log.Printf("Retrieving content from %s\n", url)
	Stats.LogCall()
	resp, err := http.Get(url)
	if err != nil {
		return Content{}, err
	}
	if resp.StatusCode != 200 {
		log.Printf("Error retrieving content: %s", resp.Status)
		return Content{}, errors.New(resp.Status)
	}
	defer resp.Body.Close()

	// Read the response body into a byte slice
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %s", err)
		return Content{}, err
	}

	// Convert the byte slice to a string
	//bodyString := string(bodyBytes)
	//log.Printf("Response body: %s", bodyString)

	//Decode the JSON from the string
	var response Response
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		log.Printf("Error decoding JSON: %s", err)
		return Content{}, err
	}

	//log.Println(response.Content)
	return response.Content, err
}

func GetSingleItem(query Query) (Result, error) {
	params := query.ToUrlParams()
	params.Del("id") // id needs to be in the URL instead
	url := singleItemUrl + query.Id + "?" + params.Encode()
	//log.Println("Getting url: ",url)
	Stats.LogCall()
	resp, err := http.Get(url)
	if err != nil {
		return Result{}, err
	}
	if resp.StatusCode != 200 {
		log.Printf("Error retrieving content: %s", resp.Status)
		return Result{}, errors.New(resp.Status)
	}
	defer resp.Body.Close()

	// Read the response body into a byte slice
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %s", err)
		return Result{}, err
	}

	// Convert the byte slice to a string
	//bodyString := string(bodyBytes)
	//log.Printf("Response body: %s", bodyString)

	//Decode the JSON from the string
	var response SingleItemResponse
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		log.Printf("Error decoding JSON: %s", err)
		return Result{}, err
	}

	//log.Println(response.Content)
	return response.Content.Result, err
}
