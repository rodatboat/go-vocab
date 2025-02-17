package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/rodatboat/go-vocab/application"
	"github.com/rodatboat/go-vocab/utils"
)

func main() {
	// args := os.Args
	Ja3 := "123"
	listId := 1
	runner := application.New(application.RunParams{
		ListId: listId,
		Ja3:    Ja3,
		// AlbCookie:  "123",
		// JSessionId: "123",
		// Guid:       "123",
	})

	// isLoggedIn := runner.IsLoggedIn()
	// if !isLoggedIn {
	// 	fmt.Println("User not logged in, exiting...")
	// 	return
	// }
	// runner.Start(listId)
	defer runner.Conn.Close(context.Background())

	// ------------- LOAD FROM FILE -----------

	byteValue, err := os.ReadFile("./example/example.start.json")
	if err != nil {
		fmt.Println(err)
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(byteValue, &result); err != nil {
		fmt.Println(err)
		return
	}

	question, _, _ := utils.ExtractQuestion(result)
	runner.SaveQuestionToDB(*question)
	answer := runner.Ask(*question)
	fmt.Println(answer)
	// utils.PrettyPrint(question)

}
