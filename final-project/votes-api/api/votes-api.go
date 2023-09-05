package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/nitishm/go-rejson/v4"

	"drexel.edu/votes-api/schema"
	"github.com/go-resty/resty/v2"
)

const (
	RedisNilError        = "redis: nil"
	RedisDefaultLocation = "0.0.0.0:6379"
	RedisKeyPrefix       = "vote:"
)

type cache struct {
	client  *redis.Client
	helper  *rejson.Handler
	context context.Context
}

type VotesAPI struct {
	cache
	voterAPIURL string
	pollAPIURL  string
	apiClient   *resty.Client
}

func NewVotesAPI(location string, voterAPIurl string, pollAPIurl string) (*VotesAPI, error) {

	apiClient := resty.New()

	client := redis.NewClient(&redis.Options{
		Addr: location,
	})

	ctx := context.Background()

	err := client.Ping(ctx).Err()
	if err != nil {
		log.Println("Error connecting to redis" + err.Error())
		return nil, err
	}

	jsonHelper := rejson.NewReJSONHandler()
	jsonHelper.SetGoRedisClientWithContext(ctx, client)

	return &VotesAPI{
		cache: cache{
			client:  client,
			helper:  jsonHelper,
			context: ctx,
		},
		voterAPIURL: voterAPIurl,
		pollAPIURL:  pollAPIurl,
		apiClient:   apiClient,
	}, nil
}

func (v *VotesAPI) GetAllVotes(c *gin.Context) {

	var votes []schema.Vote

	key := RedisKeyPrefix + "*"
	ks, _ := v.cache.client.Keys(v.context, key).Result()
	for _, key := range ks {
		var vote schema.Vote
		err := v.getVoteFromRedis(key, &vote)
		if err != nil {
			log.Println(fmt.Sprintf("An error occurred getting Vote %v from Redis.", key), err)
			// TODO: Error type?
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		votes = append(votes, vote)
	}

	if votes == nil {
		votes = make([]schema.Vote, 0)
	}

	c.JSON(http.StatusOK, votes)
}

func (v *VotesAPI) GetVote(c *gin.Context) {

	voteid := c.Request.URL.String()

	var vote schema.Vote
	key := redisKeyFromId(voteid)
	err := v.getVoteFromRedis(key, &vote)
	if err != nil {
		log.Println(fmt.Sprintf("Vote %v does not exist.", voteid), err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, vote)
}

func (v *VotesAPI) AddVote(c *gin.Context) {

	voteid := c.Request.URL.String()

	var vote schema.Vote
	if err := c.ShouldBindJSON(&vote); err != nil {
		log.Println("Error binding JSON: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	vote.VoteID = voteid

	var existingVote schema.Vote
	key := redisKeyFromId(vote.VoteID)
	if err := v.getVoteFromRedis(key, &existingVote); err == nil {
		log.Println(fmt.Sprintf("Vote %v already exists, to update a vote use the PUT method.", vote.VoterID), err)
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	vote.VoteID = c.Request.URL.String()

	voterURL := v.voterAPIURL + vote.VoterID
	var voter schema.Voter

	// checks if the Voter with VoterID exists
	resp, err := v.apiClient.R().SetResult(&voter).Get(voterURL)
	if err != nil || resp.Status() != "200 OK" {
		log.Println(fmt.Sprintf("Could not get Voter %v from Voter API", vote.VoterID))
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	// checks if the Poll with PollID exists
	pollURL := v.pollAPIURL + vote.PollID
	var poll schema.Poll
	resp, err = v.apiClient.R().SetResult(&poll).Get(pollURL)
	if err != nil || resp.Status() != "200 OK" {
		log.Println(fmt.Sprintf("Could not get Poll %v from Poll API", vote.PollID))
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	// checks if the PollOption with PollOptionID (Vote.VoteValue) exists
	pollOptionURL := v.pollAPIURL + vote.VoteValue
	var option schema.PollOption
	_, err = v.apiClient.R().SetResult(&option).Get(pollOptionURL)
	if err != nil || resp.Status() != "200 OK" {
		log.Println(fmt.Sprintf("Could not get PollOption %v from Poll API", vote.VoteValue))
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	// Add the Vote to Redis
	if _, err := v.cache.helper.JSONSet(key, ".", vote); err != nil {
		log.Println(fmt.Sprintf("An error occurred when trying to add Vote %v to Redis.", vote.VoteID) + err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// TODO: Add the voterPoll to Voter.VoteHistory

	c.Status(http.StatusOK)

}

// Delete Vote
// /votes/:voteid
func (v *VotesAPI) DeleteVote(c *gin.Context) {

	voteid := c.Request.URL.String()

	key := redisKeyFromId(voteid)
	numDeleted, err := v.client.Del(v.context, key).Result()
	if err != nil {
		log.Println(fmt.Sprintf("An error occurred deleting Vote %v from Redis.", voteid) + err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if numDeleted == 0 {
		log.Println(fmt.Sprintf("Vote %v does not exist, thus it can't be deleted from Redis.", voteid))
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
}

// Update Vote
// /votes/:voteid
// The only field that is allowed to be updated is VoteValue (PollOptionID)
// All other fields (VoterID, PollID) that are provided are ignored
func (v *VotesAPI) UpdateVote(c *gin.Context) {

	voteid := c.Request.URL.String()

	var vote schema.Vote
	if err := c.ShouldBindJSON(&vote); err != nil {
		log.Println("Error binding JSON: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	vote.VoteID = voteid

	key := redisKeyFromId(vote.VoteID)
	var existingVote schema.Vote
	if err := v.getVoteFromRedis(key, &existingVote); err != nil {
		errors.New(fmt.Sprintf("The vote to be updated Vote %v, does not exist.", vote.VoterID) + err.Error())
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	// Default value, did not provide new VoteValue to update
	if vote.VoteValue == "" {
		log.Println(fmt.Sprintf("Did not provide a VoteValue to update Vote %v.", vote.VoteID))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// checks if the PollOption with PollOptionID (Vote.VoteValue) exists
	pollOptionURL := v.pollAPIURL + vote.VoteValue
	var option schema.PollOption
	resp, err := v.apiClient.R().SetResult(&option).Get(pollOptionURL)
	if err != nil || resp.Status() != "200 OK" {
		log.Println(fmt.Sprintf("Could not get PollOption %v from Poll API", vote.VoteValue))
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	// oldVoteValue := existingVote.VoteValue

	existingVote.VoteValue = vote.VoteValue

	// Finally update Vote
	if _, err := v.helper.JSONSet(key, ".", existingVote); err != nil {
		log.Println(fmt.Sprintf("An error occurred while updating Vote %v", vote.VoteValue) + err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

}

// /votes/polls
func (v *VotesAPI) GetAllPolls(c *gin.Context) {

	url := c.Request.URL.String()
	re := regexp.MustCompile(`/polls$`)
	pollsS := string(re.Find([]byte(url)))

	pollURL := v.pollAPIURL + pollsS
	polls := []schema.Poll{}

	resp, err := v.apiClient.R().SetResult(&polls).Get(pollURL)
	if err != nil || resp.Status() != "200 OK" {
		log.Println(fmt.Sprintf("Could not get Polls from Poll API", pollsS) + err.Error())
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, polls)

}

// /votes/polls/:pollid
func (v *VotesAPI) GetPoll(c *gin.Context) {

	url := c.Request.URL.String()
	re := regexp.MustCompile(`/polls/\d+`)
	pollidS := string(re.Find([]byte(url)))

	pollURL := v.pollAPIURL + pollidS
	var poll schema.Poll
	resp, err := v.apiClient.R().SetResult(&poll).Get(pollURL)
	if err != nil || resp.Status() != "200 OK" {
		log.Println(fmt.Sprintf("Could not get Poll %v from Poll API", pollidS) + err.Error())
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, poll)

}

// /votes/polls/:pollid/options/:optionid
func (v *VotesAPI) GetPollOption(c *gin.Context) {

	url := c.Request.URL.String()
	re := regexp.MustCompile(`/polls/\d+/options\d+$`)
	optionidS := string(re.Find([]byte(url)))

	pollURL := v.pollAPIURL + optionidS
	var pollOption schema.PollOption
	resp, err := v.apiClient.R().SetResult(&pollOption).Get(pollURL)
	if err != nil || resp.Status() != "200 OK" {
		log.Println(fmt.Sprintf("Could not get pollOption %v from Poll API", optionidS) + err.Error())
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, pollOption)

}

// /votes/voters
func (v *VotesAPI) GetAllVoters(c *gin.Context) {

	url := c.Request.URL.String()
	re := regexp.MustCompile(`/voters$`)
	votersS := string(re.Find([]byte(url)))

	voterURL := v.voterAPIURL + votersS
	voters := []schema.Voter{}

	resp, err := v.apiClient.R().SetResult(&voters).Get(voterURL)
	if err != nil || resp.Status() != "200 OK" {
		log.Println(fmt.Sprintf("Could not get Voters from Voter API", votersS) + err.Error())
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, voters)

}

// /votes/voters/:voterid
func (v *VotesAPI) GetVoter(c *gin.Context) {

	url := c.Request.URL.String()
	re := regexp.MustCompile(`/voters/\d+`)
	voteridS := string(re.Find([]byte(url)))

	voterURL := v.voterAPIURL + voteridS
	var voter schema.Voter
	resp, err := v.apiClient.R().SetResult(&voter).Get(voterURL)
	if err != nil || resp.Status() != "200 OK" {
		log.Println(fmt.Sprintf("Could not get Voter %v from Poll API", &voter) + err.Error())
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, voter)

}

// /votes/voters/:voterid/polls/:pollid
func (v *VotesAPI) GetVoterPoll(c *gin.Context) {

	url := c.Request.URL.String()
	re := regexp.MustCompile(`/voters/\d+/pollss\d+$`)
	voterpollidS := string(re.Find([]byte(url)))

	voterURL := v.voterAPIURL + voterpollidS
	var voterPoll schema.VoterPoll
	resp, err := v.apiClient.R().SetResult(&voterPoll).Get(voterURL)
	if err != nil || resp.Status() != "200 OK" {
		log.Println(fmt.Sprintf("Could not get voterPoll %v from Voter API", voterpollidS) + err.Error())
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, voterPoll)

}

// func (r *ReadingListAPI) RedirectWithPublication(c *gin.Context) {
// 	rlId := c.Param("id")
// 	if rlId == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "No publication ID provided"})
// 		return
// 	}

// 	rlIdxKey := c.Param("idx")
// 	if rlIdxKey == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "No publication incdex provided"})
// 		return
// 	}

// 	cacheKey := "publist:" + rlId
// 	var rl schema.ReadingList
// 	err := r.getItemFromRedis(cacheKey, &rl)
// 	if err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "Could not find reading list in cache with id=" + cacheKey})
// 		return
// 	}

// 	pubItemLocation, ok := rl.Items[rlIdxKey]
// 	if !ok {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "Could not find publication in reading list with id=" + rlIdxKey})
// 		return
// 	}

// 	pubURL := r.pubAPIURL + pubItemLocation
// 	var pub schema.Publication

// 	_, err = r.apiClient.R().SetResult(&pub).Get(pubURL)
// 	if err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "Could not get publication from API"})
// 		return
// 	}

// 	if pub.Link == "" {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "Publication does not have a link"})
// 		return
// 	}

// 	c.Redirect(http.StatusMovedPermanently, pub.Link)
// }

//------------------------------------------------------------
// REDIS HELPERS
//------------------------------------------------------------

// We will use this later, you can ignore for now
func isRedisNilError(err error) bool {
	return errors.Is(err, redis.Nil) || err.Error() == RedisNilError
}

// In redis, our keys will be strings, they will look like
// poll:<number>.  This function will take an unsigned integer and
// return a string that can be used as a key in redis
func redisKeyFromId(id string) string {
	return fmt.Sprintf("%s%s", RedisKeyPrefix, id)
}

// Helper to return a Vote from redis provided a key
func (v *VotesAPI) getVoteFromRedis(key string, vote *schema.Vote) error {

	//Lets query redis for the item, note we can return parts of the
	//json structure, the second parameter "." means return the entire
	//json structure
	itemObject, err := v.cache.helper.JSONGet(key, ".")
	if err != nil {
		return err
	}

	//JSONGet returns an "any" object, or empty interface,
	//we need to convert it to a byte array, which is the
	//underlying type of the object, then we can unmarshal
	//it into our ToDoItem struct
	err = json.Unmarshal(itemObject.([]byte), vote)
	if err != nil {
		return err
	}

	return nil
}
