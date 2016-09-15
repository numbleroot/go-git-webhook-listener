// go-git-webhook-listener provides a very small
// tool listening on a port for incoming webhooks
// sent by a git server. Defined actions will
// be triggered on such an event.
package main

import (
	"fmt"
	"log"
	"os"

	"io/ioutil"
	"net/http"

	"github.com/joho/godotenv"
)

// Structs

type Service struct {
	IP   string
	Port string
}

// Functions

func main() {

	// Load config from environment.
	err := godotenv.Load()
	if err != nil {
		log.Fatal("[git webhook listener] Failed to load .env file. Terminating.")
	}

	service := new(Service)

	// Transfer specified config from file to struct.
	service.IP = os.Getenv("GIT_WEBHOOK_SERVICE_LISTEN_IP")
	service.Port = os.Getenv("GIT_WEBHOOK_SERVICE_LISTEN_PORT")

	// Define actions to execute on received webhooks.
	http.HandleFunc("/trigger", func(w http.ResponseWriter, req *http.Request) {

		b, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Fatalf("[git webhook listener] Reading content from HTTP request body failed with: %s\n", err.Error())
		}

		log.Printf("[git webhook listener] Received incoming git webhook:\n---\n%s\n---\n\n", b)
	})

	// Open up port to listen on for webhooks.
	log.Printf("[git webhook listener] Listening for incoming HTTP POST requests on %s:%s\n", service.IP, service.Port)
	if err := http.ListenAndServe(fmt.Sprintf("%s:%s", service.IP, service.Port), nil); err != nil {
		log.Fatalf("[git webhook listener] Web server failed with: %s\n", err.Error())
	}
}
