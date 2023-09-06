package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"drexel.edu/voter-api/api"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	hostFlag    string
	portFlag    uint
	cacheURL    string
	votesAPIURL string
)

func processCmdLineFlags() {

	flag.StringVar(&hostFlag, "h", "0.0.0.0", "Listen on all interfaces")
	flag.StringVar(&votesAPIURL, "votesapi", "http://localhost:3080", "Default endpoint for the Votes API")
	flag.StringVar(&cacheURL, "c", "0.0.0.0:6379", "Default cache location")
	flag.UintVar(&portFlag, "p", 1080, "Default Port")

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

	cacheURL = envVarOrDefault("VOTERAPI_CACHE_URL", cacheURL)
	votesAPIURL = envVarOrDefault("VOTERAPI_VOTES_API_URL", votesAPIURL)
	hostFlag = envVarOrDefault("VOTERAPI_HOST", hostFlag)

	pfNew, err := strconv.Atoi(envVarOrDefault("VOTERAPI_PORT", fmt.Sprintf("%d", portFlag)))

	if err == nil {
		portFlag = uint(pfNew)
	}

}

func main() {

	setupParms()
	log.Println("Init/cacheURL: " + cacheURL)
	log.Println("Init/votesAPIURL: " + votesAPIURL)
	log.Println("Init/hostFlag: " + hostFlag)
	log.Printf("Init/portFlag: %d", portFlag)

	apiHandler, err := api.NewVoterApi(cacheURL, votesAPIURL)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	r := gin.Default()
	r.Use(cors.Default())

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

	r.PUT("/voters/:id/polls/:pollid", apiHandler.UpdatePollData)

	serverPath := fmt.Sprintf("%s:%d", hostFlag, portFlag)
	r.Run(serverPath)
}
