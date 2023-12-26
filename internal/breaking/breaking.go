// Breaking news
package breaking

import (
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// A "Developing Story" that is being updated frequently
type DevelopingStory struct {
	gorm.Model
	Slug        string `gorm:"uniqueIndex"`
	Updates     []Update
	LastUpdated time.Time
}

// An update to a developing story
type Update struct {
	gorm.Model
	Headline string
	Body     string
	Url      string
  DevelopingStoryID uint
}

var db *gorm.DB

func init() {
	var err error
	db, err = gorm.Open(sqlite.Open("updates.db"), &gorm.Config{})
	if err != nil {
		log.Panic("failed to connect database")
	}
  db.AutoMigrate(&DevelopingStory{})
  db.AutoMigrate(&Update{})
  
}

func CreateStory(slug string) (*DevelopingStory, error) {
	story := DevelopingStory{
		Slug:        slug,
		Updates:     make([]Update, 0, 10),
		LastUpdated: time.Time{},
	}
  result := db.Create(&story)
  if result.Error != nil {
    return nil, result.Error
  }
	return &story, nil
}

func StoryExists(slug string) (*DevelopingStory, error) {
  story := &DevelopingStory{}
  result := db.Where(&DevelopingStory{Slug:slug}).First(story)
  if result.RowsAffected < 1 || result.Error != nil {
    return nil, result.Error
  }
  return story, nil
}

func (story *DevelopingStory) AddUpdate(headline string, body string, url string) error {
  update := Update{
    Headline: headline,
    Body: body,
    Url: url,
    DevelopingStoryID: story.ID,
  }
  result := db.Create(&update)
  if result.Error != nil {
    log.Printf("Error creating update: %v",result.Error)
    return result.Error
  }
  story.Updates = append(story.Updates, update)
  story.LastUpdated = time.Now()
  result = db.Save(story)
  if result.Error != nil {
    log.Printf("Error updating story: %v", result.Error)
    return result.Error
  }
  return nil
}

func AllStories() ([]DevelopingStory, error) {
  var stories []DevelopingStory
  result := db.Preload("Updates").Find(&stories)
  if result.Error != nil {
    return nil, result.Error
  }
  return stories, nil
}
