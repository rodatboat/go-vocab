package main

import (
	"context"
	"fmt"

	"github.com/rodatboat/go-vocab/application"
)

func main() {
	// args := os.Args
	Ja3 := "772,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,65281-35-27-43-10-45-18-0-23-17513-65037-5-13-16-11-51,4588-29-23-24,0"
	listId := 2444808
	runner := application.New(application.RunParams{
		ListId:     listId,
		Ja3:        Ja3,
		AlbCookie:  "PFEG/zwtYPdlCl67Ih97VHzNXz+p5BrKl+fpzht+Mlc9p1/g0egAaSzvW92U1RbCAOuMmyh/bW5N8Fjm0jUZrF7bR0nm4RCtacq8F2Hn/MPN3/9e8PDuxKc7Yern",
		JSessionId: "7846DA8D5DC65408337FD248DA80DE68",
		Guid:       "eb728ecf4d7cf82f6316111525368cc1",
	})

	isLoggedIn := runner.IsLoggedIn()
	if !isLoggedIn {
		fmt.Println("User not logged in, exiting...")
		return
	}
	runner.Practice()
	defer runner.Conn.Close(context.Background())

}
