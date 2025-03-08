package main

import (
	"context"
	"fmt"

	"github.com/rodatboat/go-vocab/application"
)

func main() {
	// args := os.Args
	Ja3 := "123"
	listId := 2444808
	runner := application.New(application.RunParams{
		ListId:     listId,
		Ja3:        Ja3,
		AlbCookie:  "123",
		JSessionId: "123",
		Guid:       "123",
	})

	isLoggedIn := runner.IsLoggedIn()
	if !isLoggedIn {
		fmt.Println("User not logged in, exiting...")
		return
	}
	runner.Practice()
	defer runner.Conn.Close(context.Background())

}
