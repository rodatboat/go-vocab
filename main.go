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
		AlbCookie:  "tcY60Vl0T4C8pnJ6vNB66ZtIQ+1fpsMbE+42KqUzvtrMy33XINek1MJPq6VV6mlJaHDkM9osFiOvql/lQXK8xQewZqozrKlZ4WI4cRYZYsknhO8eIo0YS2Q7pl7O",
		JSessionId: "E725DAE779AA26E241C298F2C79A4DA2",
		Guid:       "ef728ecf4d7cf82f6316111525368cc1",
	})

	isLoggedIn := runner.IsLoggedIn()
	if !isLoggedIn {
		fmt.Println("User not logged in, exiting...")
		return
	}
	runner.Practice()
	defer runner.Conn.Close(context.Background())

}
