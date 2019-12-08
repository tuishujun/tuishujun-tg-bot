package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"net/http"
	"os"
	"time"
)

var client http.Client
var TuiShuJunBotToken string
var TuiShuJunApiToken string

type Book struct {
	Id                int     `json:"id"`
	Name              string  `json:"book_name"`
	Status            string  `json:"book_status"`
	Author            string  `json:"book_author_name"`
	Words             int     `json:"book_words"`
	Chapters          int     `json:"book_chapters"`
	Category          string  `json:"book_category_name"`
	Synopisis         string  `json:"book_synopsis"`
	UpdatedAt         string  `json:"book_updated_at"`
	ImageUrl          string  `json:"book_img_url"`
	AuthorSource      string  `json:"book_author_source"`
	Star              float64 `json:"book_star"`
	StarNumber        int     `json:"book_star_number"`
	LatestChapterName string  `json:"book_latest_chapter_name"`
	IsBanned          bool    `json:"is_banned"`
	SourceID          string  `json:"book_source_id"`
	SourceURL         string  `json:"book_source_url"`
	SourcePcURL       string  `json:"book_source_pc_url"`
	Tags              []struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	} `json:"tags"`
}

// Function getBookInfo fetch info from tuishujun.com using a book id,
// it returns an image url and a caption of the book to be displayed in telegram.
func getBookInfo(bookId string) (imageUrl string, caption string, err error) {
	req, err := http.NewRequest("GET", "https://api.tuishujun.com/v1/books/"+bookId, nil)
	if err != nil {
		log.Fatal("Error reading request. ", err)
	}
	req.Header.Set("x-auth-token", TuiShuJunApiToken)

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var book Book

	err = decoder.Decode(&book)
	if err != nil {
		return "", "", err
	}

	imageUrl = book.ImageUrl
	caption += book.Synopisis + "\n"
	caption += fmt.Sprintf("<pre>作者: %s  状态: %s 字数: %d 万</pre>\n", book.Author, book.Status, book.Words/10000)
	if len(book.Tags) != 0 {
		tag_str := ""
		for _, tag := range book.Tags {
			tag_str += " " + tag.Name
		}
		caption += fmt.Sprintf("<pre>标签：%s </pre>\n", tag_str)
	}
	return imageUrl, caption, err
}

func main() {

	// check tg bot token is set
	if val, ok := os.LookupEnv("TUISHUJUN_BOT_TOKEN"); ok {
		TuiShuJunBotToken = val
	} else {
		log.Panic("TUISHUJUN_BOT_TOKEN not set")
	}

	// check api auth token is set
	if val, ok := os.LookupEnv("TUISHUJUN_API_TOKEN"); ok {
		TuiShuJunApiToken = val
	} else {
		log.Panic("TUISHUJUN_API_TOKEN not set")
	}

	// init net client
	client = http.Client{
		Timeout: time.Second * 10,
	}

	// init bot
	bot, err := tgbotapi.NewBotAPI(TuiShuJunBotToken)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Panic(err)
	}

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "help":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
				msg.Text += "I can help you to get book info from tuishujun.com.\n"
				msg.Text += "You can control me by sending these commands:\n"
				bot.Send(msg)
			case "book":
				bookId := update.Message.CommandArguments()
				url, caption, err := getBookInfo(bookId)
				if err != nil {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Unable to find book")
					bot.Send(msg)
				} else {
					msg := tgbotapi.NewPhotoShare(update.Message.Chat.ID, url)
					msg.Caption = caption
					msg.ParseMode = "HTML"
					bot.Send(msg)
				}
			}
		}
	}
}
