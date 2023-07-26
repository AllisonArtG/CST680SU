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
//
// TODO: This function uses the flag package to parse the command line
//		 flags.  The flag package is not very flexible and can lead to
//		 some confusing code.

//			 REQUIRED:     Study the code below, and make sure you understand
//						   how it works.  Go online and readup on how the
//						   flag package works.  Then, write a nice comment
//				  		   block to document this function that highights that
//						   you understand how it works.
//
//			 EXTRA CREDIT: The best CLI and command line processor for
//						   go is called Cobra.  Refactor this function to
//						   use it.  See github.com/spf13/cobra for information
//						   on how to use it.
//
//	 YOUR ANSWER: <GOES HERE>
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

	//r.PUT("/todo", apiHandler.UpdateToDo)
	// EXTRA CREDIT
	r.DELETE("/voters/:id", apiHandler.DeleteVoter)
	r.DELETE("/voters/:id/polls/:pollid", apiHandler.DeletePollData)

	r.UPDATE("/voters/:id", apiHandler.UpdateVoter)
	r.UPDATE("/voters/:id/polls/:pollid", apiHandler.UpdatePollData)

	r.GET("/voters/health", apiHandler.HealthCheck)

	r.GET("/crash", apiHandler.CrashSim)

	serverPath := fmt.Sprintf("%s:%d", hostFlag, portFlag)
	r.Run(serverPath)
}
