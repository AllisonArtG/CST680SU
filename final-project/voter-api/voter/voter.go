package voter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-resty/resty/v2"
	"github.com/nitishm/go-rejson/v4"
)

// VOTER STRUCTS

type voterPoll struct {
	PollID   string
	VoteDate time.Time
}

type Voter struct {
	VoterID     string
	FirstName   string `json:",omitempty"`
	LastName    string `json:",omitempty"`
	VoteHistory []voterPoll
}

func NewVoter(voterID string, first string, last string) (*Voter, error) {
	voteHistory := make([]voterPoll, 0)
	return &Voter{VoterID: voterID, FirstName: first, LastName: last, VoteHistory: voteHistory}, nil
}

// Additional STRUCTS

type Poll struct {
	PollID       string
	PollTitle    string `json:",omitempty"`
	PollQuestion string `json:",omitempty"`
	PollOptions  []pollOption
}

type pollOption struct {
	PollOptionID   string
	PollOptionText string
}

type Vote struct {
	VoteID    string
	VoterID   string
	PollID    string
	VoteValue string
}

const (
	RedisNilError        = "redis: nil"
	RedisDefaultLocation = "0.0.0.0:6379"
	RedisKeyPrefix       = "voter:"
)

type cache struct {
	cacheClient *redis.Client
	jsonHelper  *rejson.Handler
	context     context.Context
}

type VoterList struct {
	cache
	votesAPIURL string
	apiClient   *resty.Client
}

func New(location string, votesAPIurl string) (*VoterList, error) {
	return NewWithCacheInstance(location, votesAPIurl)
}

func NewWithCacheInstance(location string, votesAPIurl string) (*VoterList, error) {

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

	//Return a pointer to a new PollList struct
	return &VoterList{
		cache: cache{
			cacheClient: client,
			jsonHelper:  jsonHelper,
			context:     ctx,
		},
		votesAPIURL: votesAPIurl,
		apiClient:   apiClient,
	}, nil
}

//------------------------------------------------------------
// REDIS HELPERS
//------------------------------------------------------------

func isRedisNilError(err error) bool {
	return errors.Is(err, redis.Nil) || err.Error() == RedisNilError
}

func redisKeyFromId(id string) string {
	return fmt.Sprintf("%s%s", RedisKeyPrefix, id)
}

func (vl *VoterList) getItemFromRedis(key string, voter *Voter) error {

	itemObject, err := vl.jsonHelper.JSONGet(key, ".")
	if err != nil {
		return err
	}

	err = json.Unmarshal(itemObject.([]byte), voter)
	if err != nil {
		return err
	}

	return nil
}

//------------------------------------------------------------
// VOTER APP FUNCTIONS
//------------------------------------------------------------

// returns all Voters (as a Slice)
func (vl *VoterList) GetAllVoters() ([]Voter, error) {

	var voters []Voter

	//Lets query redis for all of the items
	key := RedisKeyPrefix + "*"
	ks, _ := vl.cacheClient.Keys(vl.context, key).Result()
	for _, key := range ks {
		var voter Voter
		err := vl.getItemFromRedis(key, &voter)
		if err != nil {
			return nil, err
		}
		voters = append(voters, voter)
	}

	return voters, nil
}

// returns the Voter with the VoterID voterID
func (vl *VoterList) GetVoter(voterID string) (Voter, error) {

	var voter Voter
	key := redisKeyFromId(voterID)
	err := vl.getItemFromRedis(key, &voter)
	if err != nil {
		return Voter{}, err
	}

	return voter, nil
}

// AddVoter accepts a Voter and adds it to Voters.
// its VoteHistory is always initialized to an empty slice
func (vl *VoterList) AddVoter(voter Voter) error {

	key := redisKeyFromId(voter.VoterID)
	var existingVoter Voter
	if err := vl.getItemFromRedis(key, &existingVoter); err == nil {
		return errors.New(fmt.Sprintf("A Voter with the ID %v already exists.", voter.VoterID))
	}

	if voter.VoteHistory == nil || len(voter.VoteHistory) > 0 {
		voter.VoteHistory = make([]voterPoll, 0)
	}

	if _, err := vl.jsonHelper.JSONSet(key, ".", voter); err != nil {
		return err
	}

	return nil
}

// returns the Voter's voterPoll where the PollID matches pollID
func (vl *VoterList) GetVoterPoll(voterID, pollID string) (voterPoll, error) {
	key := redisKeyFromId(voterID)
	var voter Voter
	if err := vl.getItemFromRedis(key, &voter); err != nil {
		return voterPoll{}, errors.New(fmt.Sprintf("Voter with ID %v does not exist.", voterID))
	}

	if len(voter.VoteHistory) == 0 {
		return voterPoll{}, errors.New(fmt.Sprintf("Poll with ID %v not found in voter %v's history.", pollID, voterID))
	}

	relevantPolls := make([]voterPoll, 0)
	for i := 0; i < len(voter.VoteHistory); i++ {
		poll := voter.VoteHistory[i]
		if poll.PollID == pollID {
			relevantPolls = append(relevantPolls, poll)
		}
	}
	if len(relevantPolls) == 0 {
		return voterPoll{}, errors.New(fmt.Sprintf("Poll with ID %v not found in voter %v's history.", pollID, voterID))
	} else if len(relevantPolls) > 1 {
		return voterPoll{}, errors.New(fmt.Sprintf("There is an error with the internal state. Multiple instances of voterPoll with ID %v in Voter %v's VoteHistory.", pollID, voterID))
	} else {
		return relevantPolls[0], nil
	}

}

// AddVoterPoll accepts the voterID, the pollID and a new Voter and adds the voterPoll to the Voter's VoteHistory
func (vl *VoterList) AddVoterPoll(voterID, pollID string, newVoter Voter) error {

	key := redisKeyFromId(voterID)
	var existingVoter Voter
	if err := vl.getItemFromRedis(key, &existingVoter); err != nil {
		return errors.New(fmt.Sprintf("Voter with ID %v does not exist.", voterID))
	}

	if len(newVoter.VoteHistory) > 1 || len(newVoter.VoteHistory) == 0 {
		return errors.New(fmt.Sprintf("Only allowed to add one new voterPoll at a time, and %v given.", len(newVoter.VoteHistory)))
	}

	poll := newVoter.VoteHistory[0]

	poll.PollID = pollID

	// Query Votes API to ensure that the Poll exists

	URL := vl.votesAPIURL + "/votes" + poll.PollID
	var pollTwo Poll

	resp, err := vl.apiClient.R().SetResult(&pollTwo).Get(URL)
	if err != nil || resp.Status() != "200 OK" {
		return errors.New(fmt.Sprintf("Could not get Poll %v from Votes API (Poll API)", poll.PollID))
	}

	// Check if the Poll

	if len(existingVoter.VoteHistory) != 0 {
		for i := 0; i < len(existingVoter.VoteHistory); i++ {
			currPoll := existingVoter.VoteHistory[i]
			if currPoll.PollID == poll.PollID {
				return errors.New(fmt.Sprintf("Poll with ID %v already exists in Voter %v's VoteHistory. Use PUT to update the voterPoll.", poll.PollID, voterID))
			}
		}
	}
	existingVoter.VoteHistory = append(existingVoter.VoteHistory, poll)
	if _, err := vl.jsonHelper.JSONSet(key, ".", existingVoter); err != nil {
		return err
	}
	return nil
}

// deletes the Voter with the VoterID voterID from Voters
func (vl *VoterList) DeleteVoter(voterID string) error {

	key := redisKeyFromId(voterID)
	numDeleted, err := vl.cacheClient.Del(vl.context, key).Result()
	if err != nil {
		return err
	}
	if numDeleted == 0 {
		return errors.New(fmt.Sprintf("An voter with the ID %v does not exist, thus they cannot be removed.", voterID))
	}

	return nil
}

// deletes the voterPoll with the PollID pollID from the Voter voterID
func (vl *VoterList) DeleteVoterPoll(voterID, pollID string) error {

	key := redisKeyFromId(voterID)
	var voter Voter
	if err := vl.getItemFromRedis(key, &voter); err != nil {
		return errors.New(fmt.Sprintf("Voter with ID %v does not exist.", voterID))
	}

	i := -1
	for index, poll := range voter.VoteHistory {
		if poll.PollID == pollID {
			i = index
			break
		}
	}

	if i == -1 {
		return errors.New(fmt.Sprintf("Poll with ID %v does not exist in Voter %v's VoteHistory", pollID, voter.VoterID))
	}

	voter.VoteHistory = append(voter.VoteHistory[:i], voter.VoteHistory[i+1:]...)

	if _, err := vl.jsonHelper.JSONSet(key, ".", voter); err != nil {
		return err
	} else {
		return nil
	}
}

// updates an existing voterPoll in Voter voterID's VoteHistory
func (vl *VoterList) UpdatePollData(voterID, pollID string, newVoter Voter) error {

	key := redisKeyFromId(voterID)
	var voter Voter
	if err := vl.getItemFromRedis(key, &voter); err != nil {
		return errors.New(fmt.Sprintf("Voter with ID %v does not exist.", voterID))
	}

	if len(newVoter.VoteHistory) > 1 || len(newVoter.VoteHistory) == 0 {
		return errors.New(fmt.Sprintf("Only allowed to update one voterPoll at a time, and %v given.", len(newVoter.VoteHistory)))
	}

	newPoll := newVoter.VoteHistory[0]

	newPoll.PollID = pollID

	for index, currPoll := range voter.VoteHistory {
		if currPoll.PollID == newPoll.PollID {
			voter.VoteHistory[index] = newPoll
			if _, err := vl.jsonHelper.JSONSet(key, ".", voter); err != nil {
				return err
			} else {
				return nil
			}
		}
	}

	return errors.New(fmt.Sprintf("Poll with ID %v does not exist in Voter %v's VoteHistory", newPoll.PollID, voterID))
}
