package main

import (
	"flag"
	"fmt"

	"drexel.edu/voter-api/api"
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
	flag.UintVar(&portFlag, "p", 1080, "Default Port")

	flag.Parse()
}

// main is the entry point for our todo API application.  It processes
// the command line flags and then uses the db package to perform the
// requested operation
func main() {
	processCmdLineFlags()
	r := gin.Default()
	r.Use(cors.Default())

	apiHandler := api.NewVoterApi()

	r.GET("/voters", apiHandler.GetAllVoters)

	r.GET("/voters/:id", apiHandler.GetVoter)
	r.POST("/voters/:id", apiHandler.AddVoter)

	r.GET("/voters/:id/polls", apiHandler.GetVoteHistory)

	r.GET("/voters/:id/polls/:pollid", apiHandler.GetPollData)
	r.POST("/voters/:id/polls/:pollid", apiHandler.AddPollData)

	r.GET("/voters/health", apiHandler.GetHealth)

	// EXTRA CREDIT

	r.DELETE("/voters/:id", apiHandler.DeleteVoter)
	r.DELETE("/voters/:id/polls/:pollid", apiHandler.DeletePollData)

	r.PUT("/voters/:id", apiHandler.UpdateVoter)
	r.PUT("/voters/:id/polls/:pollid", apiHandler.UpdatePollData)

	// LEFTOVERS (from todo-api)

	r.GET("/crash", apiHandler.CrashSim)

	serverPath := fmt.Sprintf("%s:%d", hostFlag, portFlag)
	r.Run(serverPath)
}
