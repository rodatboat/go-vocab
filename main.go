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
		AlbCookie:  "T9t25HSzk0QDly14smAk8ylw89D93AaPsdLS8aSrf/fHM+ZdofL7q4DC0oSzTjT06pInbDZHc/l47CrSJAQiUbmuUA+OEkMFd7HxCi9GvO4VqydLtNgcADsgMxJx",
		JSessionId: "CB67DA933F0F886376138C8A97F0EA06",
		Guid:       "99f79e3e51d78d8770571821802e9a11",
	})

	isLoggedIn := runner.IsLoggedIn()
	if !isLoggedIn {
		fmt.Println("User not logged in, exiting...")
		return
	}
	runner.Practice()
	defer runner.Conn.Close(context.Background())

}
