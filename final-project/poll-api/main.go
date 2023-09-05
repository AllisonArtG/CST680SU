package main

import (
	"flag"
	"fmt"
	"os"

	"drexel.edu/poll-api/api"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Global variables to hold the command line flags to drive the todo CLI
// application
var (
	hostFlag string
	portFlag uint
)

// processCmdLineFlags parses the command line flags for our CLI

func processCmdLineFlags() {

	//Note some networking lingo, some frameworks start the server on localhost
	//this is a local-only interface and is fine for testing but its not accessible
	//from other machines.  To make the server accessible from other machines, we
	//need to listen on an interface, that could be an IP address, but modern
	//cloud servers may have multiple network interfaces for scale.  With TCP/IP
	//the address 0.0.0.0 instructs the network stack to listen on all interfaces
	//We set this up as a flag so that we can overwrite it on the command line if
	//needed
	flag.StringVar(&hostFlag, "h", "0.0.0.0", "Listen on all interfaces")
	flag.UintVar(&portFlag, "p", 2080, "Default Port")

	flag.Parse()
}

// main is the entry point for our todo API application.  It processes
// the command line flags and then uses the db package to perform the
// requested operation
func main() {
	processCmdLineFlags()
	r := gin.Default()
	r.Use(cors.Default())

	apiHandler, err := api.NewPollApi()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	r.GET("/polls", apiHandler.GetAllPolls)

	r.GET("/polls/:id", apiHandler.GetPoll)
	r.POST("/polls/:id", apiHandler.AddPoll)

	r.GET("/polls/:id/options", apiHandler.GetPollOptions)

	r.GET("/polls/:id/options/:optionid", apiHandler.GetPollOption)
	r.POST("/polls/:id/options/:optionid", apiHandler.AddPollOption)

	r.GET("/polls/health", apiHandler.GetHealth)

	// Extra Credit Handlers

	r.DELETE("/polls/:id", apiHandler.DeletePoll)
	r.DELETE("/polls/:id/options/:optionid", apiHandler.DeletePollOption)

	serverPath := fmt.Sprintf("%s:%d", hostFlag, portFlag)
	r.Run(serverPath)
}
