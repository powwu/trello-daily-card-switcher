package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/adlio/trello"
)

func main() {
	apiKey := os.Getenv("TRELLO_APIKEY")
	token := os.Getenv("TRELLO_TOKEN")
	boardID := os.Getenv("TRELLO_BOARD_ID")

	if apiKey == "" || token == "" || boardID == "" {
		panic("Set your TRELLO_APIKEY, TRELLO_TOKEN, and TRELLO_BOARD_ID environment variables")
	}

	client := trello.NewClient(apiKey, token)

	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	board, err := client.GetBoard(boardID, trello.Defaults())
	if err != nil {
		panic(err)
	}

	lists, err := board.GetLists(trello.Defaults())
	if err != nil {
		panic(err)
	}

	var yesterdayList, todayList *trello.List
	for _, list := range lists {
		if list.Name == yesterday {
			yesterdayList = list
		}
		if list.Name == today {
			todayList = list
		}
	}

	if todayList == nil {
		todayList, err = board.CreateList(today, trello.Defaults())
		if err != nil {
			panic(err)
		}
	}

	if yesterdayList == nil || todayList == nil {
		panic("Could not find the required lists")
	}

	cards, err := yesterdayList.GetCards(trello.Defaults())
	if err != nil {
		panic(err)
	}

	for _, card := range cards {
		linkComplete := false
		match, err := regexp.Match(`^https:\/\/trello\.com\/c\/.*`, []byte(card.Name))
		if err != nil {
			panic(err)
		}
		if match {
			sourceShortLink, _, _ := strings.Cut(strings.TrimPrefix(card.Name, "https://trello.com/c/"), "/")
			sourceCard, err := client.GetCard(sourceShortLink)
			if err != nil {
				fmt.Printf("Failed to get source card for card '%s': %v\n", card.Name, err)
			}

			linkComplete = sourceCard.DueComplete
		}
		if !card.DueComplete && !linkComplete {
			_, err := card.CopyToList(todayList.ID, trello.Defaults())
			if err != nil {
				fmt.Printf("Failed to copy card '%s': %v\n", card.Name, err)
			} else {
				fmt.Printf("Copied card: '%s'\n", card.Name)
			}
		}
	}

	fmt.Println("Completed card switching objective.")
}
