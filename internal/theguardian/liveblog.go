package theguardian

import (
	"fmt"
	"log"
	"main/internal/breaking"
	"main/internal/openai"
	"time"
)

const (
	liveblogType = "liveblog"
)

func GetLiveblogUpdates(since time.Time) error {
	// get all the latest news
	query := Query{
		OrderBy:  OrderByNewest,
		PageSize: 20,
	}

	content, err := GetContent(query)
	if err != nil {
		log.Printf("Error retrieving content: %s", err)
		return err
	}

	// we should now have the 20 most recent stories on the Guardian website, newest first
	log.Printf("Queried %d stories, %v", content.PageSize, content.OrderBy)

	// iterate over results
	for _, result := range content.Results {

		// skip anything that's not a liveblog
		if result.Type != liveblogType {
			continue
		}

		// skip sports and austalia news
		if result.SectionId == "sport" || result.SectionId == "australia-news" {
			log.Printf("Skipping: %v (%v)", result.SectionId, result.WebTitle)
			continue
		}

		// parse publish date
		published, err := time.Parse(time.RFC3339, result.WebPublicationDate)
		if err != nil {
			log.Printf("Couldn't parse date: %s", err)
			continue
		}
		// skip older stories
		if published.Before(since) {
			continue
		}

		log.Printf("Found liveblog updated %v - %v\n", published.Format("15:04"), result.WebTitle)

		// check if we've seen this story before
		story, err := breaking.StoryExists(result.Id)
		if err != nil {
			log.Printf("Story does not exist (%v), creating new one", err)
			story, err = breaking.CreateStory(result.Id)
			if err != nil {
				log.Printf("Aborting story update: %v", err)
				continue
			}
		}

		// now get the contents of the story
		item, err := GetSingleItem(Query{
			Id:         result.Id,
			ShowFields: ShowFieldsBody,
		})
		if err != nil {
			log.Printf("Could not retrieve item %v: %v", result.Id, err)
			continue
		}
		// get the most recent update
		update, err := ParseLiveBlogBody(item.Fields.Body)
		if err != nil {
			log.Printf("Could not parse body: %v", err)
			continue
		}

		// add context for OpenAI
		update += "NOTE--this update is part of a developing story, " + result.WebTitle

		// check if duplicate
		if breaking.UpdateIsDuplicate(update) {
			log.Printf("Duplicate, skipping (%v)", item.Id)
			continue
		}
		// generate a headline (liveblog updates typically do not have one)
		headline, err := openai.HeadlineForText(update)
		if err != nil {
			log.Printf("Unable to generate headline: %v", err)
			return err
		}
		// add update to story
		err = story.AddUpdate(headline, update, item.WebUrl)
		if err != nil {
			log.Println(err)
		}
		fmt.Printf("Latest update: %v\n", headline)
	}
	return nil
}
