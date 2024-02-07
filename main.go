package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var host string
var email string
var apiKey string

var DBFileName = "zulip.db"
var db *gorm.DB

func main() {
	if len(os.Args) != 5 {
		fmt.Printf("Usage: %s <data or files> <host> <email> <api_key>", os.Args[0])
		os.Exit(1)
	}

	cmd := os.Args[1]
	host = os.Args[2]
	email = os.Args[3]
	apiKey = os.Args[4]

	var err error

	// Open database
	db, err = gorm.Open(sqlite.Open(DBFileName), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		c, err := db.DB()
		if err == nil {
			c.Close()
		}
	}()

	db.AutoMigrate(&Stream{})
	db.AutoMigrate(&File{})
	db.AutoMigrate(&Message{})

	switch cmd {
	case "data":
		archive()

	case "files":
		avatars()
		files()

	default:
		log.Fatal("unknown command")
	}
}

func files() {
	regFile := regexp.MustCompile(`(?:href|src)="(\/[^"]+)`)

	// Process messages in batches of ten

	offset := 0
	limit := 10

	var messages []Message
	for {
		err := db.Offset(offset).Limit(10).Find(&messages).Error
		if err != nil {
			log.Fatal(err)
		}

		for _, msg := range messages {
			matches := regFile.FindAllStringSubmatch(msg.Content, -1)

			for _, match := range matches {
				getFile(match[1])
			}
		}
		if len(messages) < limit {
			break
		}
		messages = nil
		offset += limit
	}
}

func avatars() {
	var avatars []string

	err := db.Raw("SELECT DISTINCT avatar_url FROM messages").Scan(&avatars).Error
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Found %d Avatars", len(avatars))
	for _, a := range avatars {
		getFile(a)
	}
}

func getFile(path string) {
	var file File
	var err error

	db.Select("path").Where("path", path).First(&file)
	if file.Path != "" {
		log.Println("Skipping existing:", path)
		return
	}

	file.Path = path
	file.Data, file.ContentType, err = zulipGetFile(path)
	file.Size = len(file.Data)
	if err != nil {
		log.Printf("Error downloading file: %s: %s", err, path)
	} else {
		log.Printf("%12d bytes: %s", len(file.Data), path)
		err = db.Create(&file).Error
		if err != nil {
			log.Printf("Error creating file: %s: %s", err, path)
		}
		time.Sleep(time.Millisecond * 200)
	}
	file.Data = nil

}

func archive() {
	if len(os.Args) < 3 {
		log.Fatalf("Expected at least three args: %s download <host> <email> [<password>]\nPassword can optionally be set in local environment as API_KEY", os.Args[0])
	}
	host = os.Args[2]
	email = os.Args[3]

	if len(os.Args) > 4 {
		apiKey = os.Args[4]
	} else {
		apiKey = os.Getenv("API_KEY")
		if apiKey == "" {
			log.Fatalf("Missing API_KEY either as third parameter or as API_KEY environment variable")
		}
	}

	streams, err := getStreams()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Found %d streams", len(streams))

	for _, s := range streams {
		err = db.Create(&s).Error
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Processing stream %s\n", s.Name)
		topics, err := getStreamTopics(s.StreamID)
		if err != nil {
			log.Fatal(err)
		}
		for _, t := range topics {
			// err = db.Create(&t).Error
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Getting messages for '%s: %s':\n", s.Name, t.Name)
			err := GetStreamTopicMessagesCB(s.Name, t.Name, func(messages []Message) error {
				for _, message := range messages {
					err = db.Create(&message).Error
					if err != nil {
						return err
					}
				}
				log.Printf(" - fetched %d messages", len(messages))
				return nil
			})
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func getStreams() ([]Stream, error) {
	result, err := zulipGet("/api/v1/streams")
	if err != nil {
		return nil, err
	}
	return result.Streams, nil
}

func getStreamTopics(StreamID uint) ([]Topic, error) {
	result, err := zulipGet(fmt.Sprintf("/api/v1/users/me/%d/topics", StreamID))
	if err != nil {
		return nil, err
	}
	return result.Topics, nil
}

func GetStreamTopicMessagesCB(Stream string, Topic string, CallBack func([]Message) error) error {
	anchor := uint(0)
	numBefore := 0
	numAfter := 500

	path := "/api/v1/messages?"

	// Repeatedly gets messages for given stream/topic until none left to fetch
	// We could actually just get all messages for all streams/topic...
	for {
		q := url.Values{}
		q.Add("anchor", fmt.Sprintf("%d", anchor))
		q.Add("num_before", fmt.Sprintf("%d", numBefore))
		q.Add("num_after", fmt.Sprintf("%d", numAfter))
		q.Add("narrow", fmt.Sprintf(`[{"negated":false,"operator":"stream","operand":"%s"},{"negated":false,"operator":"topic","operand":"%s"}]`,
			Stream, Topic))

		url := path + q.Encode()
		// log.Println(url)
		result, err := zulipGet(url)
		if err != nil {
			return err
		}
		if len(result.Messages) == 0 {
			break
		}
		anchor = result.Messages[len(result.Messages)-1].MessageID + 1
		err = CallBack(result.Messages)
		if err != nil {
			return err
		}
		time.Sleep(time.Second)
	}

	return nil
}

type Result struct {
	Result   string    `json:"result"`
	Streams  []Stream  `json:"streams"`
	Topics   []Topic   `json:"topics"`
	Messages []Message `json:"messages"`
}

func zulipGet(path string) (result *Result, err error) {

	url := "https://" + host + path
	ctx, _ := context.WithTimeout(context.Background(), time.Minute)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return
	}
	req.SetBasicAuth(email, apiKey)

	ret, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	defer ret.Body.Close()

	if ret.StatusCode != 200 {
		return nil, errors.New(ret.Status)
	}

	dec := json.NewDecoder(ret.Body)

	err = dec.Decode(&result)
	if err != nil {
		return
	}

	return
}

func zulipGetFile(path string) (b []byte, contentType string, err error) {
	url := "https://" + host + path
	ctx, _ := context.WithTimeout(context.Background(), time.Minute)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return
	}
	req.SetBasicAuth(email, apiKey)

	ret, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	defer ret.Body.Close()

	if ret.StatusCode != 200 {
		err = errors.New(ret.Status)
	}

	contentType = ret.Header.Get("Content-Type")

	b, err = io.ReadAll(ret.Body)

	return
}
