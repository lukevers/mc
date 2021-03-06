package main

import (
	"bufio"
	"fmt"
	"github.com/kierdavis/mc/mcclient"
	"io"
	"os"
)

func main() {
	/*
		defer func() {
			v := recover()
			if v != nil {
				err, ok := v.(error)
				if ok {
					fmt.Printf("Error: %s\n", err.Error())
				}
			}
		}()
	*/

	if len(os.Args) < 2 {
		fmt.Printf("Not enough arguments\n\nusage: %s <server address>\n\nThis program expects the MC_USER and MC_PASSWD environment variables to be set. Otherwise, the user is logged in with an offline account.\n", os.Args[0])
		os.Exit(2)
	}

	fmt.Printf("*** Welcome to mcchat!\n")

	addr := os.Args[1]
	username := os.Getenv("MC_USER")
	password := os.Getenv("MC_PASSWD")

	var debugWriter io.Writer
	debugWriter = os.Stdout

	fmt.Printf("*** Logging in...\n")

	var err error
	var client *mcclient.Client

	if password == "" {
		if username == "" {
			username = "Player"
		}

		client = mcclient.LoginOffline(username, debugWriter)

	} else {
		client, err = mcclient.Login(username, password, debugWriter)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			os.Exit(1)
		}
	}

	go func() {
		err := <-client.ErrChan
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			client.Leave()
			client.Logout()
		}
	}()

	client.HandleMessage = func(msg string) {
		fmt.Printf("\r%s\n>", mcclient.ANSIEscapes(msg))
	}

	fmt.Printf("*** Connecting to %s...\n", addr)

	err = client.Join(addr)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		client.Logout()
		os.Exit(1)
	}

	fmt.Printf("*** Connected!\n*** Type & press enter to send messages!\n*** Press Ctrl+D to exit\n\n")

	go func() {
		stdinReader := bufio.NewReader(os.Stdin)

		fmt.Printf(">")

		for {
			msg, err := stdinReader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					fmt.Printf("Error: %s\n", err.Error())
				}

				client.Leave()
				client.Logout()
				return
			}

			fmt.Printf("\x1b[1T>")
			client.Chat(msg[:len(msg)-1])
		}
	}()

	kickmsg := client.Run()
	if kickmsg != "" {
		fmt.Printf("\nKicked: %s\n", kickmsg)
	}
}
