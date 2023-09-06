package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"drexel.edu/votes-api/api"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	hostFlag    string
	portFlag    uint
	cacheURL    string
	voterAPIURL string
	pollAPIURL  string
)

func processCmdLineFlags() {

	flag.StringVar(&hostFlag, "h", "0.0.0.0", "Listen on all interfaces")
	flag.StringVar(&voterAPIURL, "voterapi", "http://localhost:1080", "Default endpoint for the Voter API")
	flag.StringVar(&pollAPIURL, "pollapi", "http://localhost:2080", "Default endpoint for the Poll API")
	flag.StringVar(&cacheURL, "c", "0.0.0.0:6379", "Default cache location")
	flag.UintVar(&portFlag, "p", 3080, "Default Port")

	flag.Parse()
}

func envVarOrDefault(envVar string, defaultVal string) string {
	envVal := os.Getenv(envVar)
	if envVal != "" {
		return envVal
	}
	return defaultVal
}

func setupParms() {
	processCmdLineFlags()

	cacheURL = envVarOrDefault("VOTESAPI_CACHE_URL", cacheURL)
	voterAPIURL = envVarOrDefault("VOTESAPI_VOTER_API_URL", voterAPIURL)
	pollAPIURL = envVarOrDefault("VOTESAPI_POLL_API_URL", pollAPIURL)
	hostFlag = envVarOrDefault("VOTESAPI_HOST", hostFlag)

	pfNew, err := strconv.Atoi(envVarOrDefault("VOTESAPI_PORT", fmt.Sprintf("%d", portFlag)))

	if err == nil {
		portFlag = uint(pfNew)
	}

}

func main() {
	setupParms()
	log.Println("Init/cacheURL: " + cacheURL)
	log.Println("Init/voterAPIURL: " + voterAPIURL)
	log.Println("Init/pollAPIURL: " + pollAPIURL)
	log.Println("Init/hostFlag: " + hostFlag)
	log.Printf("Init/portFlag: %d", portFlag)

	apiHandler, err := api.NewVotesAPI(cacheURL, voterAPIURL, pollAPIURL)

	if err != nil {
		panic(err)
	}

	r := gin.Default()
	r.Use(cors.Default())

	r.GET("/votes", apiHandler.GetAllVotes)
	r.GET("/votes/:voteid", apiHandler.GetVote)
	r.POST("/votes/:voteid", apiHandler.AddVote)
	r.DELETE("/votes/:voteid", apiHandler.DeleteVote)
	r.PUT("/votes/:voteid", apiHandler.UpdateVote)

	r.GET("/votes/voters", apiHandler.GetAllVoters)
	r.GET("/votes/voters/:voterid", apiHandler.GetVoter)
	r.GET("/votes/voters/:voterid/polls", apiHandler.GetVoterPolls)
	r.GET("/votes/voters/:voterid/polls/:pollid", apiHandler.GetVoterPoll)

	r.GET("/votes/polls", apiHandler.GetAllPolls)
	r.GET("/votes/polls/:pollid", apiHandler.GetPoll)
	r.GET("/votes/polls/:pollid/options", apiHandler.GetPollOptions)
	r.GET("/votes/polls/:pollid/options/:optionid", apiHandler.GetPollOption)

	// EXTRA CREDIT

	// r.DELETE("/voters/:id/polls/:pollid", apiHandler.DeletePollData)

	// r.PUT("/voters/:id", apiHandler.UpdateVoter)
	// r.PUT("/voters/:id/polls/:pollid", apiHandler.UpdatePollData)

	// // LEFTOVERS (from todo-api)

	// r.GET("/crash", apiHandler.CrashSim)

	serverPath := fmt.Sprintf("%s:%d", hostFlag, portFlag)
	r.Run(serverPath)
}
