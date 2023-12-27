package main

import (
	"fmt"
	"log"
	"main/internal/breaking"
	"main/internal/openai"
	"main/internal/theguardian"
	"os"
	"time"
)

func retrieve() {
	theguardian.ApiKey = os.Getenv("GUARDIAN_API")
	if theguardian.ApiKey == "" {
		log.Panic("Guardian API key missing")
	}
	if openai.ApiKey = os.Getenv("OPENAI_API"); openai.ApiKey == "" {
		log.Panic("OpenAI API key missing")
	}

	err := theguardian.GetLiveblogUpdates(time.Now().Add(-1 * time.Hour))
	if err != nil {
		log.Println(err)
	}
}

func printStories(stories []breaking.DevelopingStory) {
	for _, story := range stories {
		fmt.Printf("Slug: %s ", story.Slug)
		fmt.Printf("(last updated %s)\n", story.LastUpdated.Format("15:04"))
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

func dumpDB(since time.Time) {
	var stories []breaking.DevelopingStory
	var err error
	if since.IsZero() {
		stories, err = breaking.AllStories()
	} else {
		stories, err = breaking.StoriesSince(time.Now().Add(-1 * time.Hour))
	}
	if err != nil {
		log.Println(err)
		os.Exit(-1)
	}
	printStories(stories)
}

func main() {
	welcome := `
BREAKING NEWS
-------------
1. All stories in DB
2. Last Hour
3. Refresh
`
	fmt.Println(welcome)
	fmt.Printf("Make a choice: ")
	var choice int
	_, err := fmt.Scanf("%d", &choice)
	if err != nil {
		log.Println(err)
		os.Exit(-1)
	}
	//log.Printf("Choice: %d", choice)
	switch choice {
	case 1:
		dumpDB(time.Time{})
	case 2:
		dumpDB(time.Now().Add(-1 * time.Hour))
	case 3:
		retrieve()
	default:
		fmt.Printf("Invalid choice: %d", choice)
	}
}
