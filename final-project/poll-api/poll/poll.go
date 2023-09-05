package poll

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/nitishm/go-rejson/v4"
)

//STRUCTS

type pollOption struct {
	PollOptionID   string
	PollOptionText string
}

type Poll struct {
	PollID       string
	PollTitle    string `json:",omitempty"`
	PollQuestion string `json:",omitempty"`
	PollOptions  []pollOption
}

const (
	RedisNilError        = "redis: nil"
	RedisDefaultLocation = "0.0.0.0:6379"
	RedisKeyPrefix       = "poll:"
)

type cache struct {
	cacheClient *redis.Client
	jsonHelper  *rejson.Handler
	context     context.Context
}

type PollList struct {

	//Redis cache connections
	cache
}

// New is a constructor function that returns a pointer to a new
// PollList struct.  If this is called it uses the default Redis URL
// with the companion constructor NewWithCacheInstance.
func New() (*PollList, error) {
	//We will use an override if the REDIS_URL is provided as an environment
	//variable, which is the preferred way to wire up a docker container
	redisUrl := os.Getenv("POLLAPI_REDIS_URL")
	//This handles the default condition
	if redisUrl == "" {
		redisUrl = RedisDefaultLocation
	}
	return NewWithCacheInstance(redisUrl)
}

// NewWithCacheInstance is a constructor function that returns a pointer to a new
// PollList struct.  It accepts a string that represents the location of the redis
// cache.
func NewWithCacheInstance(location string) (*PollList, error) {

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

	//Return a pointer to a new PollList struct
	return &PollList{
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
// poll:<number>.  This function will take an unsigned integer and
// return a string that can be used as a key in redis
func redisKeyFromId(id string) string {
	return fmt.Sprintf("%s%s", RedisKeyPrefix, id)
}

// Helper to return a Poll from redis provided a key
func (pl *PollList) getItemFromRedis(key string, poll *Poll) error {

	//Lets query redis for the item, note we can return parts of the
	//json structure, the second parameter "." means return the entire
	//json structure
	itemObject, err := pl.jsonHelper.JSONGet(key, ".")
	if err != nil {
		return err
	}

	//JSONGet returns an "any" object, or empty interface,
	//we need to convert it to a byte array, which is the
	//underlying type of the object, then we can unmarshal
	//it into our Poll struct
	err = json.Unmarshal(itemObject.([]byte), poll)
	if err != nil {
		return err
	}

	return nil
}

//------------------------------------------------------------
// POLL APP FUNCTIONS
//------------------------------------------------------------

// returns all Polls (as a Slice)
func (pl *PollList) GetAllPolls() ([]Poll, error) {

	var polls []Poll

	//Lets query redis for all of the items
	key := RedisKeyPrefix + "*"
	ks, _ := pl.cacheClient.Keys(pl.context, key).Result()
	for _, key := range ks {
		var poll Poll
		err := pl.getItemFromRedis(key, &poll)
		if err != nil {
			return nil, err
		}
		polls = append(polls, poll)
	}

	return polls, nil
}

// returns the Poll with the PollID pollID
func (pl *PollList) GetPoll(pollOptionID string) (Poll, error) {

	var poll Poll
	key := redisKeyFromId(pollOptionID)
	err := pl.getItemFromRedis(key, &poll)
	if err != nil {
		return Poll{}, err
	}

	return poll, nil
}

// AddPoll accepts a Poll and adds it to Polls.
// its PollOptions is always initialized to an empty slice
func (pl *PollList) AddPoll(poll Poll) error {

	key := redisKeyFromId(poll.PollID)
	var existingVoter Poll
	if err := pl.getItemFromRedis(key, &existingVoter); err == nil {
		return errors.New(fmt.Sprintf("A Poll with the ID %v already exists.", poll.PollID))
	}

	if poll.PollOptions == nil || len(poll.PollOptions) > 0 {
		poll.PollOptions = make([]pollOption, 0)
	}

	if _, err := pl.jsonHelper.JSONSet(key, ".", poll); err != nil {
		return err
	}

	return nil
}

// returns the Poll's pollPoll where the PollID matches pollOptionID
func (pl *PollList) GetPollOption(pollID, pollOptionID string) (pollOption, error) {
	key := redisKeyFromId(pollID)
	var poll Poll
	if err := pl.getItemFromRedis(key, &poll); err != nil {
		return pollOption{}, errors.New(fmt.Sprintf("Poll with ID %v does not exist.", pollID))
	}

	if len(poll.PollOptions) == 0 {
		return pollOption{}, errors.New(fmt.Sprintf("pollOption with ID %v not found in Poll %v's PollOptions.", pollOptionID, pollID))
	}

	relevantOptions := make([]pollOption, 0)
	for i := 0; i < len(poll.PollOptions); i++ {
		pollOption := poll.PollOptions[i]
		if pollOption.PollOptionID == pollOptionID {
			relevantOptions = append(relevantOptions, pollOption)
		}
	}
	if len(relevantOptions) == 0 {
		return pollOption{}, errors.New(fmt.Sprintf("pollOption with ID %v not found in Poll %v's PollOptions.", pollOptionID, pollID))
	} else if len(relevantOptions) > 1 {
		return pollOption{}, errors.New(fmt.Sprintf("There is an error with the internal state. There is an error with the internal state. Multiple instances of pollOption with ID %v in Poll %v's PollOptions.", pollOptionID, pollID))
	} else {
		return relevantOptions[0], nil
	}

}

// AddPollOption accepts the pollID, the pollOptionID, and a new Poll and adds the pollOption to the Poll's PollOptions
func (pl *PollList) AddPollOption(pollID, pollOptionID string, newPoll Poll) error {

	key := redisKeyFromId(pollID)
	var existingPoll Poll
	if err := pl.getItemFromRedis(key, &existingPoll); err != nil {
		return errors.New(fmt.Sprintf("Poll with ID %v does not exist.", pollID))
	}

	if len(newPoll.PollOptions) > 1 || len(newPoll.PollOptions) == 0 {
		return errors.New(fmt.Sprintf("Only allowed to add one new pollOption at a time, and %v given.", len(newPoll.PollOptions)))
	}

	pollOption := newPoll.PollOptions[0]

	pollOption.PollOptionID = pollOptionID

	if len(existingPoll.PollOptions) != 0 {
		for i := 0; i < len(existingPoll.PollOptions); i++ {
			currOption := existingPoll.PollOptions[i]
			if currOption.PollOptionID == pollOption.PollOptionID {
				return errors.New(fmt.Sprintf("PollOption with ID %v already exists in Poll %v's PollOptions.", pollOptionID, pollID))
			}
		}
	}
	existingPoll.PollOptions = append(existingPoll.PollOptions, pollOption)
	if _, err := pl.jsonHelper.JSONSet(key, ".", existingPoll); err != nil {
		return err
	}
	return nil
}

// // deletes the Poll with the PollID pollID from Polls
func (pl *PollList) DeletePoll(pollID string) error {

	key := redisKeyFromId(pollID)
	numDeleted, err := pl.cacheClient.Del(pl.context, key).Result()
	if err != nil {
		return err
	}
	if numDeleted == 0 {
		return errors.New(fmt.Sprintf("An poll with the ID %v does not exist, thus it cannot be removed.", pollID))
	}

	return nil
}

// deletes the pollOption pollOptionID from the Poll PollID
func (pl *PollList) DeletePollOption(pollID, pollOptionID string) error {

	key := redisKeyFromId(pollID)
	var poll Poll
	if err := pl.getItemFromRedis(key, &poll); err != nil {
		return errors.New(fmt.Sprintf("Poll with ID %v does not exist.", pollOptionID))
	}

	i := -1
	for index, poll := range poll.PollOptions {
		if poll.PollOptionID == pollOptionID {
			i = index
			break
		}
	}

	if i == -1 {
		return errors.New(fmt.Sprintf("pollOption %v does not exist in Poll %v's PollOptions:", pollOptionID, poll.PollID))
	}

	poll.PollOptions = append(poll.PollOptions[:i], poll.PollOptions[i+1:]...)

	if _, err := pl.jsonHelper.JSONSet(key, ".", poll); err != nil {
		return err
	} else {
		return nil
	}
}

// // updates an existing Poll with the newPoll's fields (PollTitle and PollQuestion)
// func (pl *PollList) UpdatePoll(newPoll Poll) error {

// 	key := redisKeyFromId(newPoll.PollID)
// 	var existingPoll Poll
// 	if err := pl.getItemFromRedis(key, &existingPoll); err != nil {
// 		return errors.New(fmt.Sprintf("The poll to be updated Poll %v, does not exist.", newPoll.PollID))
// 	}

// 	// Only updates the included Poll fields, and leaves not included ones unchanged
// 	if newPoll.PollTitle != "" {
// 		existingPoll.PollTitle = newPoll.PollTitle
// 	}
// 	if newPoll.PollQuestion != "" {
// 		existingPoll.PollQuestion = newPoll.PollQuestion
// 	}

// 	//Add item to database with JSON Set.  Note there is no update
// 	//functionality, so we just overwrite the existing item
// 	if _, err := pl.jsonHelper.JSONSet(key, ".", existingPoll); err != nil {
// 		return err
// 	}

// 	//If everything is ok, return nil for the error
// 	return nil
// }

// // updates an existing pollOption in Poll ID's PollOptions
// func (pl *PollList) UpdatePollData(pollID, pollOptionID string, newPoll Poll) error {

// 	key := redisKeyFromId(pollID)
// 	var poll Poll
// 	if err := pl.getItemFromRedis(key, &poll); err != nil {
// 		return errors.New(fmt.Sprintf("Poll with ID %v does not exist.", pollID))
// 	}

// 	if len(newPoll.PollOptions) > 1 || len(newPoll.PollOptions) == 0 {
// 		return errors.New(fmt.Sprintf("Only allowed to update one pollOption at a time, and %v given.", len(newPoll.PollOptions)))
// 	}

// 	newPollOption := newPoll.PollOptions[0]

// 	newPollOption.PollOptionID = pollOptionID

// 	for index, currOption := range poll.PollOptions {
// 		if currOption.PollOptionID == pollOptionID {
// 			poll.PollOptions[index] = newPollOption
// 			if _, err := pl.jsonHelper.JSONSet(key, ".", poll); err != nil {
// 				return err
// 			} else {
// 				return nil
// 			}
// 		}
// 	}

// 	return errors.New(fmt.Sprintf("Poll with ID %v does not exist in Poll %v's VoteHistory", newPoll.PollID, pollOptionID))
// }
