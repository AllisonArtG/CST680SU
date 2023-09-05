package api

import (
	"fmt"
	"log"
	"net/http"
	"regexp"

	"drexel.edu/poll-api/poll"
	"github.com/gin-gonic/gin"
)

type PollAPI struct {
	pollList *poll.PollList
}

func NewPollApi() (*PollAPI, error) {
	pollListHandler, err := poll.New()
	if err != nil {
		return nil, err
	}

	return &PollAPI{pollList: pollListHandler}, nil
}

// THE API FUNCTIONS

// implementation for GET /polls
// returns all Polls
func (p *PollAPI) GetAllPolls(c *gin.Context) {
	polls, err := p.pollList.GetAllPolls()
	if err != nil {
		log.Println("Error getting all Polls: ", err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	if polls == nil {
		polls = make([]poll.Poll, 0)
	}

	c.JSON(http.StatusOK, polls)
}

// implementation for GET /polls/:id
// returns a single Poll
func (p *PollAPI) GetPoll(c *gin.Context) {

	idS := c.Request.URL.String()

	poll, err := p.pollList.GetPoll(idS)
	if err != nil {
		log.Println(fmt.Sprintf("Poll with the ID %v not found: ", idS), err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, poll)
}

// implementation for POST /polls/id
// adds a new Poll
// any data included in the Poll's PollOptions is ignored
// and because the Poll.PollID field is redundant (it's equivalent to the URL),
// if the user includes PollID in the JSON it is simply overridden by the URL
func (p *PollAPI) AddPoll(c *gin.Context) {

	var poll poll.Poll
	if err := c.ShouldBindJSON(&poll); err != nil {
		log.Println("Error binding JSON: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	poll.PollID = c.Request.URL.String()

	if err := p.pollList.AddPoll(poll); err != nil {
		log.Println(fmt.Sprintf("Error adding Poll with the ID %v: ", poll.PollID), err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

// implementation for GET /polls/:id/options
// returns the poll options (PollOptions) for the Poll with ID id
func (p *PollAPI) GetPollOptions(c *gin.Context) {

	url := c.Request.URL.String()
	re := regexp.MustCompile(`^/polls/\d+`)
	idS := string(re.Find([]byte(url)))

	poll, err := p.pollList.GetPoll(idS)
	if err != nil {
		log.Println(fmt.Sprintf("Poll with the ID %v not found: ", idS), err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, poll.PollOptions)
}

// implementation for GET /polls/:id/options/:optionid
// returns the poll option (pollOption) for the Poll with ID id for pollOption optionid
func (v *PollAPI) GetPollOption(c *gin.Context) {

	url := c.Request.URL.String()
	re_id := regexp.MustCompile(`^/polls/\d+`)
	idS := string(re_id.Find([]byte(url)))

	optionidS := url

	poll, err := v.pollList.GetPollOption(idS, optionidS)
	if err != nil {
		log.Println(fmt.Sprintf("Error finding PollOptionID %v in Poll %v's PollOptions: ", optionidS, idS), err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, poll)
}

// implementation for POST /polls/:id/options/:optionid
// adds the poll option (pollOption) to the PollOptions of the Poll with ID id
// only one poll option can be added at a time, and additional fields in Poll
// outside of PollOptions are ignored (PollID, PollTitle, PollQuestion)
// and because the pollOption.PollOptionID field is redundant (included in the
// URL - "/options/:optionid"), if the user includes PollOptionID in the JSON it is
// simply overridden by the PollOptionID in the URL
func (v *PollAPI) AddPollOption(c *gin.Context) {

	url := c.Request.URL.String()
	re_id := regexp.MustCompile(`^/polls/\d+`)
	idS := string(re_id.Find([]byte(url)))

	optionidS := url

	// TODO: query poll-api to see if this poll even exists first before adding
	// Determine if this is necessary and where best to do this.

	var poll poll.Poll
	if err := c.ShouldBindJSON(&poll); err != nil {
		log.Println("Error binding JSON: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := v.pollList.AddPollOption(idS, optionidS, poll); err != nil {
		log.Println("Error adding pollOption: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

// implementation of GET /voters/health
// returns the health of the Voter-API Application
func (v *PollAPI) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK,
		gin.H{
			"status":  "ok",
			"version": "1.0.0",
		})
}

// Extra Credit Handlers

// implementation for DELETE /polls/:pollid
// deletes a Poll
func (v *PollAPI) DeletePoll(c *gin.Context) {

	idS := c.Request.URL.String()

	if err := v.pollList.DeletePoll(idS); err != nil {
		log.Println(fmt.Sprintf("Error deleting Polls with ID %v: ", idS), err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.Status(http.StatusOK)
}

// implementation for DELETE /polls/:pollid/options/:optionid
// deletes pollOption for the Poll
func (v *PollAPI) DeletePollOption(c *gin.Context) {

	url := c.Request.URL.String()
	re_id := regexp.MustCompile(`^/polls/\d+`)
	pollidS := string(re_id.Find([]byte(url)))

	optionidS := url

	err := v.pollList.DeletePollOption(pollidS, optionidS)
	if err != nil {
		log.Println(fmt.Sprintf("Error deleting pollOption %v from Poll %v's: ", optionidS, pollidS), err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.Status(http.StatusOK)
}

// // implementation for PUT /voters/:id/polls/:pollid
// // updates a voterPoll, specifically its VoteDate, of the Voter with ID id
// // only one voterPoll is allowed to be updated at a time
// // any data in the Voter fields outside of VoteHistory will be ignored because
// // voter.VoterID and voterPoll.PollID fields are redundant (they are included
// // in the URL), if the user includes either of them in the JSON, they are overridden
// func (v *PollAPI) UpdatePollData(c *gin.Context) {

// 	url := c.Request.URL.String()
// 	re_id := regexp.MustCompile(`^/voters/\d+`)
// 	idS := string(re_id.Find([]byte(url)))
// 	re_pollid := regexp.MustCompile(`/polls/\d+$`)
// 	pollidS := string(re_pollid.Find([]byte(url)))

// 	var voter voter.Voter

// 	if err := c.ShouldBindJSON(&voter); err != nil {
// 		log.Println("Error binding JSON: ", err)
// 		c.AbortWithStatus(http.StatusBadRequest)
// 		return
// 	}

// 	if err := v.pollList.UpdatePollData(idS, pollidS, voter); err != nil {
// 		log.Println(fmt.Sprintf("Error updating poll in Voter %v's history: ", idS), err)
// 		c.AbortWithStatus(http.StatusInternalServerError)
// 		return
// 	}

// 	c.Status(http.StatusOK)

// }
