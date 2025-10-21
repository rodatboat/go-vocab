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
		AlbCookie:  "GTJXq1l+e+m/o+oN/lzWvWAOmfXToOQMkgWeUQFKV7L7r/Tx6JsBOsIkrVQEwdbEkCBa+DJEAMo88ZL1DUbI4v3f9A5hBvA4zW5sjVbRJTbwBHEeLUaOcgbeWNH6",
		JSessionId: "BEBFF76A71CB483F7F2AB869188E25E5",
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
