package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"heat-meter-read-client/kamstrup"
	"heat-meter-read-client/server"
)

func main() {
	portPtr := flag.String("port", "", "The port of the IR transceiver")

	flag.Parse()

	if *portPtr == "" {
		fmt.Println("No port provided (using --port option), stopping.")

		os.Exit(1)

		return
	}

	kampstrupClient, err := kamstrup.CreateKamstrupClient(*portPtr, time.Millisecond*500)

	if err != nil {
		fmt.Printf("Unable to create Kampstrup client (full error below), stopping.\n\n%s\n", err)

		os.Exit(1)

		return
	}

	defer kampstrupClient.ClosePort()

	if err := server.CreateAndRunWebServer(kampstrupClient); err != nil {
		fmt.Printf("Unable to run web server (full error below), stopping.\n\n%s\n", err)

		os.Exit(1)

		return
	}
}
