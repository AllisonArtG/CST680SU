package api

import (
	"fmt"
	"log"
	"net/http"
	"regexp"

	"drexel.edu/voter-api/voter"
	"github.com/gin-gonic/gin"
)

type VoterAPI struct {
	voterList *voter.VoterList
}

func NewVoterApi(location string, votesAPIurl string) (*VoterAPI, error) {
	voterListHandler, err := voter.New(location, votesAPIurl)
	if err != nil {
		return nil, err
	}

	return &VoterAPI{voterList: voterListHandler}, nil
}

// THE API FUNCTIONS

// implementation for GET /voters
// returns all Voters
func (v *VoterAPI) GetAllVoters(c *gin.Context) {
	voters, err := v.voterList.GetAllVoters()
	if err != nil {
		log.Println("Error getting all Voters: ", err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	if voters == nil {
		voters = make([]voter.Voter, 0)
	}

	c.JSON(http.StatusOK, voters)
}

// implementation for GET /voters/:id
// returns a single Voter
func (v *VoterAPI) GetVoter(c *gin.Context) {

	idS := c.Request.URL.String()

	voter, err := v.voterList.GetVoter(idS)
	if err != nil {
		log.Println(fmt.Sprintf("Voter with the ID %v not found: ", idS), err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, voter)
}

// implementation for POST /voters/id
// adds a new Voter
// any data included in the Voter's VoteHistory is ignored
// and because the voter.VoterID field is redundant (it's equivalent to the URL),
// if the user includes VoterID in the JSON it is simply overridden by the URL
func (v *VoterAPI) AddVoter(c *gin.Context) {

	var voter voter.Voter
	if err := c.ShouldBindJSON(&voter); err != nil {
		log.Println("Error binding JSON: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	voter.VoterID = c.Request.URL.String()

	if err := v.voterList.AddVoter(voter); err != nil {
		log.Println(fmt.Sprintf("Error adding Voter with the ID %v: ", voter.VoterID), err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

// implementation for GET /voters/:id/polls
// returns the voting history (VoteHistory) for the Voter with ID id
func (v *VoterAPI) GetVoteHistory(c *gin.Context) {

	url := c.Request.URL.String()
	re := regexp.MustCompile(`^/voters/\d+`)
	idS := string(re.Find([]byte(url)))

	voter, err := v.voterList.GetVoter(idS)
	if err != nil {
		log.Println(fmt.Sprintf("Voter with the ID %v not found: ", idS), err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, voter.VoteHistory)
}

// implementation for GET /voters/:id/polls/:pollid
// returns the poll data (voterPoll) for the Voter with ID id for voterPoll pollid
func (v *VoterAPI) GetPollData(c *gin.Context) {

	url := c.Request.URL.String()
	re_id := regexp.MustCompile(`^/voters/\d+`)
	idS := string(re_id.Find([]byte(url)))
	re_pollid := regexp.MustCompile(`/polls/\d+$`)
	pollidS := string(re_pollid.Find([]byte(url)))

	poll, err := v.voterList.GetVoterPoll(idS, pollidS)
	if err != nil {
		log.Println(fmt.Sprintf("Error finding PollID %v in Voter %v's VoteHistory: ", pollidS, idS), err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, poll)
}

// implementation for POST /voters/:id/polls/:pollid
// adds the poll data (voterPoll) to the VoteHistory of the Voter with ID id
// only one voterPoll can be added at a time, and additional fields in Voter
// outside of VoteHistory are ignored (VoterID, FirstName, LastName)
// and because the voterPoll.PollID field is redundant (included in the
// URL - "/polls/:pollid"), if the user includes VoterID in the JSON it is
// simply overridden by the PollID in the URL
func (v *VoterAPI) AddPollData(c *gin.Context) {

	url := c.Request.URL.String()
	re_id := regexp.MustCompile(`^/voters/\d+`)
	idS := string(re_id.Find([]byte(url)))
	re_pollid := regexp.MustCompile(`/polls/\d+$`)
	pollidS := string(re_pollid.Find([]byte(url)))

	var voter voter.Voter
	if err := c.ShouldBindJSON(&voter); err != nil {
		log.Println("Error binding JSON: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := v.voterList.AddVoterPoll(idS, pollidS, voter); err != nil {
		log.Println("Error adding poll: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

// implementation of GET /voters/health
// returns the health of the Voter-API Application
func (v *VoterAPI) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK,
		gin.H{
			"status":  "ok",
			"version": "3.0.0",
		})
}

// Extra Credit Handlers

// implementation for DELETE /voters/:id
// deletes a Voter
func (v *VoterAPI) DeleteVoter(c *gin.Context) {

	idS := c.Request.URL.String()

	if err := v.voterList.DeleteVoter(idS); err != nil {
		log.Println(fmt.Sprintf("Error deleting Voter with ID %v: ", idS), err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.Status(http.StatusOK)
}

// implementation for DELETE /voters/:id/polls/:pollid
// deletes the data (voterPoll) for the Voter with ID id and voterPoll pollid
func (v *VoterAPI) DeletePollData(c *gin.Context) {

	url := c.Request.URL.String()
	re_id := regexp.MustCompile(`^/voters/\d+`)
	idS := string(re_id.Find([]byte(url)))
	re_pollid := regexp.MustCompile(`/polls/\d+$`)
	pollidS := string(re_pollid.Find([]byte(url)))

	err := v.voterList.DeleteVoterPoll(idS, pollidS)
	if err != nil {
		log.Println(fmt.Sprintf("Error deleting %v from Voter %v's history: ", pollidS, idS), err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.Status(http.StatusOK)
}

// implementation for PUT /voters/:id/polls/:pollid
// updates a voterPoll, specifically its VoteDate, of the Voter with ID id
// only one voterPoll is allowed to be updated at a time
// any data in the Voter fields outside of VoteHistory will be ignored because
// voter.VoterID and voterPoll.PollID fields are redundant (they are included
// in the URL), if the user includes either of them in the JSON, they are overridden
func (v *VoterAPI) UpdatePollData(c *gin.Context) {

	url := c.Request.URL.String()
	re_id := regexp.MustCompile(`^/voters/\d+`)
	idS := string(re_id.Find([]byte(url)))
	re_pollid := regexp.MustCompile(`/polls/\d+$`)
	pollidS := string(re_pollid.Find([]byte(url)))

	var voter voter.Voter

	if err := c.ShouldBindJSON(&voter); err != nil {
		log.Println("Error binding JSON: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := v.voterList.UpdatePollData(idS, pollidS, voter); err != nil {
		log.Println(fmt.Sprintf("Error updating poll in Voter %v's history: ", idS), err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)

}
