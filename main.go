// go-git-webhook-listener provides a very small
// tool listening on a port for incoming webhooks
// sent by a git server. Defined actions will
// be triggered on such an event.
package main

import (
	"fmt"
	"log"
	"os"

	"net/http"
	"os/exec"

	"github.com/joho/godotenv"
)

// Structs

// HugoWebsite stores information about the project that
// is to be automatically rebuild by this tool. This
// includes the process ID of the hugo server as well as
// the path to the local repository to rebuild.
type HugoWebsite struct {
	HugoProc *os.Process
	Repo     string
}

// Functions

// StartHugo starts the hugo static site generator
// in server mode to deliver already compiled website.
func (hugoSite *HugoWebsite) StartHugo() error {

	// Define 'hugo server' command to be run from repository.
	cmdHugoServer := exec.Command("hugo", "server")
	cmdHugoServer.Dir = hugoSite.Repo

	// Start the process but immediately detach from it.
	if err := cmdHugoServer.Start(); err != nil {
		return err
	}

	// Save process to struct.
	hugoSite.HugoProc = cmdHugoServer.Process

	return nil
}

// StopHugo ends the currently running hugo static
// site generator process.
func (hugoSite *HugoWebsite) StopHugo() error {

	// Tell the process to stop immediately.
	if err := hugoSite.HugoProc.Kill(); err != nil {
		return err
	}

	// Wait for process to exit and collect
	// process status information.
	if _, err := hugoSite.HugoProc.Wait(); err != nil {
		return err
	}

	return nil
}

// RebuildHugo defines the actions to be done when
// a git webhook was received that indicated change
// in a hugo website repository. In such case, this
// function pulls the repository and lets hugo compile
// and serve the static files.
func (hugoSite *HugoWebsite) RebuildHugo() {

	log.Println()
	log.Println("[git webhook listener] Hugo rebuild was triggered.")
	log.Printf("[git webhook listener] Pulling latest commit on '%s'.\n", hugoSite.Repo)

	// Specify to run 'git pull' on repository to be rebuilt.
	cmdGitPull := exec.Command("git", "pull")
	cmdGitPull.Dir = hugoSite.Repo

	// Run the command.
	if err := cmdGitPull.Run(); err != nil {
		log.Fatalf("[git webhook listener]  => 'git pull' failed with: %s. Terminating.\n", err.Error())
	}

	log.Println("[git webhook listener]  => 'git pull' succeeded.")
	log.Println("[git webhook listener] Stopping hugo server process.")

	// Stop the currently running hugo command.
	if err := hugoSite.StopHugo(); err != nil {
		log.Fatalf("[git webhook listener]  => Stopping hugo server failed with: %s. Terminating.\n", err.Error())
	}

	log.Println("[git webhook listener]  => hugo server stopped.")
	log.Printf("[git webhook listener] Removing '%s' folder in repository.", (hugoSite.Repo + "/public"))

	// Delete the public folder to prohibit stale content.
	if err := os.RemoveAll((hugoSite.Repo + "/public")); err != nil {
		log.Fatalf("[git webhook listener]  => RemoveAll('%s') failed with: %s. Terminating.\n", (hugoSite.Repo + "/public"), err.Error())
	}

	log.Printf("[git webhook listener]  => RemoveAll('%s')  succeeded.\n", (hugoSite.Repo + "/public"))
	log.Println("[git webhook listener] Recompiling files with hugo.")

	// Execute hugo's compile command.
	cmdHugo := exec.Command("hugo")
	cmdHugo.Dir = hugoSite.Repo

	// Run the command.
	if err := cmdHugo.Run(); err != nil {
		log.Fatalf("[git webhook listener]  => 'hugo' failed with: %s. Terminating.\n", err.Error())
	}

	log.Println("[git webhook listener]  => 'hugo' succeeded.")
	log.Println("[git webhook listener] Starting hugo server again.")

	// Start hugo server again.
	if err := hugoSite.StartHugo(); err != nil {
		log.Fatalf("[git webhook listener]  => Starting hugo server failed: %s\n", err.Error())
	}

	log.Println("[git webhook listener] SUCCESS: hugo server started again. All done. Goodbye.")
}

func main() {

	// Load config from environment.
	err := godotenv.Load()
	if err != nil {
		log.Fatal("[git webhook listener] Failed to load .env file. Terminating.")
	}

	hugoSite := new(HugoWebsite)

	// Transfer specified config from file to struct.
	ip := os.Getenv("GIT_WEBHOOK_LISTENER_IP")
	port := os.Getenv("GIT_WEBHOOK_LISTENER_PORT")
	hugoSite.Repo = os.Getenv("GIT_WEBHOOK_REBUILD_REPO_PATH")

	// Start hugo.
	if err := hugoSite.StartHugo(); err != nil {
		log.Fatalf("[git webhook listener] Could not start hugo server: %s\n", err.Error())
	}

	// Define actions to execute on received webhooks.
	http.HandleFunc("/trigger", func(w http.ResponseWriter, req *http.Request) {
		hugoSite.RebuildHugo()
	})

	// Open up port to listen on for webhooks.
	log.Printf("[git webhook listener] Listening for incoming HTTP POST requests on %s:%s\n", ip, port)
	if err := http.ListenAndServe(fmt.Sprintf("%s:%s", ip, port), nil); err != nil {
		log.Fatalf("[git webhook listener] Web server failed with: %s\n", err.Error())
	}
}
