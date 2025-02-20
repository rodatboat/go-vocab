package main

import (
	"context"
	"fmt"

	"github.com/rodatboat/go-vocab/application"
)

func main() {
	// args := os.Args
	Ja3 := "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,27-45-65037-16-65281-23-11-18-17513-13-43-51-5-0-10-35,4588-29-23-24,0"
	listId := 2444808
	runner := application.New(application.RunParams{
		ListId:     listId,
		Ja3:        Ja3,
		AlbCookie:  "Npqqeq3zvWYvd8aDEmMy4wL5ysnMUhgiXrQTy5kRxCWPRpf92wL3LlKrgJcOkyUKBC9SncOmzjhvGSdNrBa2b/wjFdFxUurxKAYkRMtfHIx4Qa0Gj0IURY8nqbIB",
		JSessionId: "918A4CF505CC5B23C543E54CCFAAF875",
		Guid:       "58dd986658adeece01b2d60475772ac5",
	})

	isLoggedIn := runner.IsLoggedIn()
	if !isLoggedIn {
		fmt.Println("User not logged in, exiting...")
		return
	}
	runner.Practice()
	defer runner.Conn.Close(context.Background())

}
