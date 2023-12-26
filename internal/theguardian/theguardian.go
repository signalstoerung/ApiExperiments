package theguardian

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"reflect"
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

var ApiKey string

/*
status	The status of the response. It refers to the state of the API. Successful calls will receive an "ok" even if your query did not return any results	String
total	The number of results available for your search overall	Integer
pageSize	The number of items returned in this call	Integer
currentPage	The number of the page you are browsing	Integer
pages	The total amount of pages that are in this call	Integer
orderBy	The sort order used	String
id	The path to content	String
sectionId	The id of the section	String
sectionName	The name of the section	String
webPublicationDate	The combined date and time of publication	Datetime
webUrl	The URL of the html content	String
apiUrl	The URL of the raw content	String
*/

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
	Status  string `json:"status"`
	Total   int    `json:"total"`
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
			params.Add(fieldTag, field.String())
		} else {
			//log.Printf("Field %s has no tag or is zero", fieldType.Name)
		}
	}
	params.Add("api-key", ApiKey)
	//log.Printf("params: %v", params)
	return params
}

func GetContent(query Query) (Content, error) {
	params := query.ToUrlParams()
	url := contentUrl + "?" + params.Encode()
	//log.Printf("Retrieving content from %s\n", url)
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
