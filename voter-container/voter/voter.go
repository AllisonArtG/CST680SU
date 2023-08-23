package voter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/nitishm/go-rejson/v4"
)

//STRUCTS

type voterPoll struct {
	PollID   uint
	VoteDate time.Time
}

type Voter struct {
	VoterID     uint
	FirstName   string `json:",omitempty"`
	LastName    string `json:",omitempty"`
	VoteHistory []voterPoll
}

// type VoterList struct {
// 	Voters map[uint]Voter //A map of VoterIDs as keys and Voter structs as values
// }

func NewVoter(voterID uint, first string, last string) (*Voter, error) {
	voteHistory := make([]voterPoll, 0)
	return &Voter{VoterID: voterID, FirstName: first, LastName: last, VoteHistory: voteHistory}, nil
}

const (
	RedisNilError        = "redis: nil"
	RedisDefaultLocation = "0.0.0.0:6379"
	RedisKeyPrefix       = "todo:"
)

type cache struct {
	cacheClient *redis.Client
	jsonHelper  *rejson.Handler
	context     context.Context
}

type VoterList struct {
	//more things would be included in a real implementation

	//Redis cache connections
	cache
}

// New is a constructor function that returns a pointer to a new
// ToDo struct.  If this is called it uses the default Redis URL
// with the companion constructor NewWithCacheInstance.
func New() (*VoterList, error) {
	//We will use an override if the REDIS_URL is provided as an environment
	//variable, which is the preferred way to wire up a docker container
	redisUrl := os.Getenv("REDIS_URL")
	//This handles the default condition
	if redisUrl == "" {
		redisUrl = RedisDefaultLocation
	}
	return NewWithCacheInstance(redisUrl)
}

// NewWithCacheInstance is a constructor function that returns a pointer to a new
// ToDo struct.  It accepts a string that represents the location of the redis
// cache.
func NewWithCacheInstance(location string) (*VoterList, error) {

	//Connect to redis.  Other options can be provided, but the
	//defaults are OK
	client := redis.NewClient(&redis.Options{
		Addr: location,
	})

	//We use this context to coordinate betwen our go code and
	//the redis operaitons
	ctx := context.Background()

	//This is the reccomended way to ensure that our redis connection
	//is working
	err := client.Ping(ctx).Err()
	if err != nil {
		log.Println("Error connecting to redis" + err.Error())
		return nil, err
	}

	//By default, redis manages keys and values, where the values
	//are either strings, sets, maps, etc.  Redis has an extension
	//module called ReJSON that allows us to store JSON objects
	//however, we need a companion library in order to work with it
	//Below we create an instance of the JSON helper and associate
	//it with our redis connnection
	jsonHelper := rejson.NewReJSONHandler()
	jsonHelper.SetGoRedisClientWithContext(ctx, client)

	//Return a pointer to a new ToDo struct
	return &VoterList{
		cache: cache{
			cacheClient: client,
			jsonHelper:  jsonHelper,
			context:     ctx,
		},
	}, nil
}

//------------------------------------------------------------
// REDIS HELPERS
//------------------------------------------------------------

// We will use this later, you can ignore for now
func isRedisNilError(err error) bool {
	return errors.Is(err, redis.Nil) || err.Error() == RedisNilError
}

// In redis, our keys will be strings, they will look like
// voter:<number>.  This function will take an unsigned integer and
// return a string that can be used as a key in redis
func redisKeyFromId(id uint) string {
	return fmt.Sprintf("%s%d", RedisKeyPrefix, id)
}

// Helper to return a ToDoItem from redis provided a key
func (vl *VoterList) getItemFromRedis(key string, voter *Voter) error {

	//Lets query redis for the item, note we can return parts of the
	//json structure, the second parameter "." means return the entire
	//json structure
	itemObject, err := vl.jsonHelper.JSONGet(key, ".")
	if err != nil {
		return err
	}

	//JSONGet returns an "any" object, or empty interface,
	//we need to convert it to a byte array, which is the
	//underlying type of the object, then we can unmarshal
	//it into our Voter struct
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
	var voter Voter

	//Lets query redis for all of the items
	key := RedisKeyPrefix + "*"
	ks, _ := vl.cacheClient.Keys(vl.context, key).Result()
	for _, key := range ks {
		err := vl.getItemFromRedis(key, &voter)
		if err != nil {
			return nil, err
		}
		voters = append(voters, voter)
	}

	return voters, nil
}

// returns the Voter with the VoterID voterID
func (vl *VoterList) GetVoter(voterID uint) (Voter, error) {

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
func (vl *VoterList) GetVoterPoll(voterID, pollID uint) (voterPoll, error) {
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
		return voterPoll{}, errors.New(fmt.Sprintf("There is an error with the internal state. Voter %v was allowed to vote more than once in poll %v.", voterID, pollID))
	} else {
		return relevantPolls[0], nil
	}

}

// AddVoterPoll accepts the voterID and a new Voter and adds the voterPoll to the Voter's VoteHistory
func (vl *VoterList) AddVoterPoll(voterID uint, newVoter Voter) error {

	key := redisKeyFromId(voterID)
	var existingVoter Voter
	if err := vl.getItemFromRedis(key, &existingVoter); err != nil {
		return errors.New(fmt.Sprintf("Voter with ID %v does not exist.", voterID))
	}

	if len(newVoter.VoteHistory) > 1 || len(newVoter.VoteHistory) == 0 {
		return errors.New(fmt.Sprintf("Only allowed to add one new voterPoll at a time, and %v given.", len(newVoter.VoteHistory)))
	}

	poll := newVoter.VoteHistory[0]

	if len(existingVoter.VoteHistory) != 0 {
		for i := 0; i < len(existingVoter.VoteHistory); i++ {
			currPoll := existingVoter.VoteHistory[i]
			if currPoll.PollID == poll.PollID {
				return errors.New(fmt.Sprintf("Poll with ID %v already exists in Voter %v's VoteHistory. Voters are only allowed to vote once per poll.", poll.PollID, voterID))
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
func (vl *VoterList) DeleteVoter(voterID uint) error {

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
func (vl *VoterList) DeleteVoterPoll(voterID, pollID uint) error {

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

// updates an existing Voter with the newVoter's fields (FirstName and LastName)
func (vl *VoterList) UpdateVoter(newVoter Voter) error {

	key := redisKeyFromId(newVoter.VoterID)
	var existingVoter Voter
	if err := vl.getItemFromRedis(key, &existingVoter); err != nil {
		return errors.New(fmt.Sprintf("The voter to be updated Voter %v, does not exist.", newVoter.VoterID))
	}

	// Only updates the included Voter fields, and leaves not included ones unchanged
	if newVoter.FirstName != "" {
		existingVoter.FirstName = newVoter.FirstName
	}
	if newVoter.LastName != "" {
		existingVoter.LastName = newVoter.LastName
	}

	//Add item to database with JSON Set.  Note there is no update
	//functionality, so we just overwrite the existing item
	if _, err := vl.jsonHelper.JSONSet(key, ".", existingVoter); err != nil {
		return err
	}

	//If everything is ok, return nil for the error
	return nil
}

// updates an existing voterPoll in Voter voterID's VoteHistory
func (vl *VoterList) UpdatePollData(voterID uint, newVoter Voter) error {

	key := redisKeyFromId(voterID)
	var voter Voter
	if err := vl.getItemFromRedis(key, &voter); err != nil {
		return errors.New(fmt.Sprintf("Voter with ID %v does not exist.", voterID))
	}

	if len(newVoter.VoteHistory) > 1 || len(newVoter.VoteHistory) == 0 {
		return errors.New(fmt.Sprintf("Only allowed to update one voterPoll at a time, and %v given.", len(newVoter.VoteHistory)))
	}

	newPoll := newVoter.VoteHistory[0]

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

// Extra Functions

// PrintItem accepts a Voter and prints it to the console
// in a JSON pretty format.
func (vl *VoterList) PrintVoter(voter Voter) {
	jsonBytes, _ := json.MarshalIndent(voter, "", "  ")
	fmt.Println(string(jsonBytes))
}

// PrintAllItems accepts a slice of Voter and prints them to the console
// in a JSON pretty format.
func (vl *VoterList) PrintAllItems(voters []Voter) {
	for _, voter := range voters {
		vl.PrintVoter(voter)
	}
}
