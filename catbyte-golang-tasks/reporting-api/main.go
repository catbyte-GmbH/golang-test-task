package main

import (
	"errors"
	"net/http"
	"sort"
	"strings"
	jsonhandling "twitch_chat_analysis/jsonHandling"
	db "twitch_chat_analysis/radis-db"

	"github.com/gin-gonic/gin"
)

type timeSlice []jsonhandling.Message

func (p timeSlice) Len() int {
	return len(p)
}

func (p timeSlice) Less(i, j int) bool {
	return p[i].CreatedAt.After(p[j].CreatedAt)
}

func (p timeSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func chronologicalDescendingOrder(current []jsonhandling.Message) []jsonhandling.Message {
	dateSortedReviews := make(timeSlice, 0, len(current))

	for _, d := range current {
		dateSortedReviews = append(dateSortedReviews, d)
	}

	sort.Sort(dateSortedReviews)

	return dateSortedReviews
}

func redisConn() []jsonhandling.Message {
	database, err := db.NewClient("localhost:6379")
	db.Must(err)

	var msgs []jsonhandling.Message
	data := db.Get(database, "messages")
	current := jsonhandling.RadisJsonUnmarshal([]byte(data), msgs)

	return chronologicalDescendingOrder(current)
}

func getHttpRequestJson(context *gin.Context) {
	context.IndentedJSON(http.StatusOK, redisConn())
}

func getMessageByParameters(sender, receiver string) (timeSlice, error) {
	counter := 0
	var founded []jsonhandling.Message

	for _, obj := range redisConn() {
		if strings.EqualFold(sender, obj.Sender) && strings.EqualFold(receiver, obj.Receiver) {
			founded = append(founded, obj)
			counter++
		}
	}

	if counter == 0 {
		return nil, errors.New("message not found")
	} else {
		return chronologicalDescendingOrder(founded), nil
	}
}

func getHttpRequestJsonByParameters(context *gin.Context) {
	sender := context.Param("sender")
	receiver := context.Param("receiver")
	message, err := getMessageByParameters(sender, receiver)
	if err != nil {
		context.IndentedJSON(http.StatusNotFound, gin.H{"response": "message not found"})
		return
	}

	context.IndentedJSON(http.StatusOK, message)
}

func main() {
	router := gin.Default()

	router.GET("/message/list", getHttpRequestJson)
	router.GET("/message/list/:sender/:receiver", getHttpRequestJsonByParameters)

	err := router.Run("localhost:8081")
	if err != nil {
		return
	}
}
