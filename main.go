package main

import (
	"fmt"
	"log"
	"main/internal/openai"
	"main/internal/theguardian"
  "main/internal/breaking"
	"os"
	"time"
)

//var apiKey string = ""

func retrieve() {
	apiKey := os.Getenv("GUARDIAN_API")
	if apiKey == "" {
		log.Panic("No API key provided")
	}
	theguardian.ApiKey = apiKey
	openai.ApiKey = os.Getenv("OPENAI_API")

	//log.Printf("Found API key: %s\n", apiKey)

	query := theguardian.Query{
		OrderBy: theguardian.OrderByNewest,
	}
	content, err := theguardian.GetContent(query)
	if err != nil {
		log.Printf("Error retrieving content: %s", err)
		return
	}
	//fmt.Printf("Found %d results\n", content.Total)
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	for _, result := range content.Results {
		if result.Type != "liveblog" {
			continue
		}
		published, err := time.Parse(time.RFC3339, result.WebPublicationDate)
		if err != nil {
			log.Printf("Couldn't parse date: %s", err)
			continue
		}
		if published.After(oneHourAgo) {
      story, err := breaking.StoryExists(result.Id)
      if err != nil {
        log.Printf("Story does not exist (%v), creating new one", err)
        story, err = breaking.CreateStory(result.Id)
        if err != nil {
          log.Printf("Aborting story update: %v", err)
          continue
        }
      }
			fmt.Printf("Liveblog updated at %v - %v\n", published.Format("15:04"), result.WebTitle)
			item, err := theguardian.GetSingleItem(theguardian.Query{
				Id:         result.Id,
				ShowFields: theguardian.ShowFieldsBody,
			})
			if err != nil {
				log.Printf("Could not retrieve item %v: %v", result.Id, err)
				continue
			}
			//log.Printf("parsing body: %v", item.Fields.Body)
			update, err := theguardian.ParseLiveBlogBody(item.Fields.Body)
			if err != nil {
				log.Printf("Could not parse body: %v", err)
				continue
			}
      update += "NOTE--this update is part of a developing story, " + result.WebTitle
			headline, err := openai.HeadlineForText(update)
			if err != nil {
				log.Printf("Unable to generate headline: %v", err)
				return
			}
      err = story.AddUpdate(headline, update, item.WebUrl)
      if err != nil {
        log.Println(err)
      }
			fmt.Printf("Latest update: %v\n", headline)
      log.Printf("%+v",story)
		}

	}
}

func dumpDB() {
  stories, err := breaking.AllStories()
  if err != nil {
    log.Println(err)
    os.Exit(-1)
  }
  for _, story := range stories {
    fmt.Printf("Slug: %s ", story.Slug)
    fmt.Printf("(last updated %s\n", story.LastUpdated.Format("15:04"))
    fmt.Println("-----------------------------------------------------")
    for _, update := range story.Updates {
      fmt.Println(update.Headline)
      fmt.Println(update.Body)
      fmt.Println(update.Url)
      fmt.Printf("Created: %s\n", update.CreatedAt.Format("15:04"))
      fmt.Println("+++")
      }
    fmt.Println("* * * * *")
  }
}

func main() {
  welcome := `
BREAKING NEWS
-------------
1. All stories in DB
2. Refresh
`
  fmt.Println(welcome)
  fmt.Printf("Make a choice: ")
  var choice int
  _, err := fmt.Scanf("%d", &choice)
  if err != nil {
    log.Println(err)
    os.Exit(-1)
  }
  log.Printf("Choice: %d", choice)
  switch choice {
    case 1: dumpDB()
    case 2: retrieve()
    default: fmt.Printf("Invalid choice: %d", choice)
  }
}