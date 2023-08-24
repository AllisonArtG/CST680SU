package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"drexel.edu/voter-api/voter"
	"github.com/gin-gonic/gin"
)

type VoterAPI struct {
	voterList *voter.VoterList
}

func NewVoterApi() (*VoterAPI, error) {
	voterListHandler, err := voter.New()
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

	idS := c.Param("id")
	id64, err := strconv.ParseUint(idS, 10, 32)
	if err != nil {
		log.Println(fmt.Sprintf("Error converting Voter ID %v to uint64: ", idS), err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	voter, err := v.voterList.GetVoter(uint(id64))
	if err != nil {
		log.Println(fmt.Sprintf("Voter with the ID %v not found: ", id64), err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, voter)
}

// implementation for POST /voters
// adds a new Voter
// any data included in the Voter's VoteHistory is ignored
func (v *VoterAPI) AddVoter(c *gin.Context) {

	var voter voter.Voter
	if err := c.ShouldBindJSON(&voter); err != nil {
		log.Println("Error binding JSON: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

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

	idS := c.Param("id")
	id64, err := strconv.ParseUint(idS, 10, 32)
	if err != nil {
		log.Println(fmt.Sprintf("Error converting Voter ID %v to uint64: ", idS), err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	voter, err := v.voterList.GetVoter(uint(id64))
	if err != nil {
		log.Println(fmt.Sprintf("Voter with the ID %v not found: ", id64), err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, voter.VoteHistory)
}

// implementation for GET /voters/:id/polls/:pollid
// returns the poll data (voterPoll) for the Voter with ID id for voterPoll pollid
func (v *VoterAPI) GetPollData(c *gin.Context) {

	idS := c.Param("id")
	id64, err := strconv.ParseUint(idS, 10, 32)
	if err != nil {
		log.Println(fmt.Sprintf("Error converting Voter ID %v to uint64: ", idS), err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	pollidS := c.Param("pollid")
	pollid64, err := strconv.ParseUint(pollidS, 10, 32)
	if err != nil {
		log.Println(fmt.Sprintf("Error converting poll with PollID %v to uint64: ", pollidS), err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	poll, err := v.voterList.GetVoterPoll(uint(id64), uint(pollid64))
	if err != nil {
		log.Println(fmt.Sprintf("Error finding PollID %v in Voter %v's VoteHistory: ", pollid64, id64), err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, poll)
}

// implementation for POST /voters/:id/polls
// adds the poll data (voterPoll) to the VoteHistory of the Voter with ID id
// only one voterPoll can be added at a time, and additional fields in Voter
// outside of VoteHistory are ignored (VoterID, FirstName, LastName)
func (v *VoterAPI) AddPollData(c *gin.Context) {

	idS := c.Param("id")
	id64, err := strconv.ParseUint(idS, 10, 32)
	if err != nil {
		log.Println(fmt.Sprintf("Error converting Voter ID %v to uint64: ", idS), err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var voter voter.Voter
	if err := c.ShouldBindJSON(&voter); err != nil {
		log.Println("Error binding JSON: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := v.voterList.AddVoterPoll(uint(id64), voter); err != nil {
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
			"status":             "ok",
			"version":            "1.0.0",
			"uptime":             100,
			"users_processed":    1000,
			"errors_encountered": 10,
		})
}

// Extra Credit Handlers

// implementation for DELETE /voters/:id
// deletes a Voter
func (v *VoterAPI) DeleteVoter(c *gin.Context) {
	idS := c.Param("id")
	id64, err := strconv.ParseUint(idS, 10, 32)
	if err != nil {
		log.Println(fmt.Sprintf("Error converting Voter ID %v to uint64: ", idS), err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := v.voterList.DeleteVoter(uint(id64)); err != nil {
		log.Println(fmt.Sprintf("Error deleting Voter with ID %v: ", id64), err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.Status(http.StatusOK)
}

// implementation for DELETE /voters/:id/polls/:pollid
// deletes the data (voterPoll) for the Voter with ID id and voterPoll pollid
func (v *VoterAPI) DeletePollData(c *gin.Context) {

	idS := c.Param("id")
	id64, err := strconv.ParseUint(idS, 10, 32)
	if err != nil {
		log.Println(fmt.Sprintf("Error converting Voter ID %v to uint64: ", idS), err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	pollidS := c.Param("pollid")
	pollid64, err := strconv.ParseUint(pollidS, 10, 32)
	if err != nil {
		log.Println(fmt.Sprintf("Error converting poll ID %v to uint64: ", pollidS), err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	err = v.voterList.DeleteVoterPoll(uint(id64), uint(pollid64))
	if err != nil {
		log.Println(fmt.Sprintf("Error deleting %v from Voter %v's history: ", pollid64, id64), err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.Status(http.StatusOK)
}

// implementation for PUT /voters
// updates the Voter fields (FirstName, LastName) of a Voter
// any data included in the Voter's VoteHistory is ignored
// if any of the Voter fields are omitted, then the original ones remain unchanged
func (v *VoterAPI) UpdateVoter(c *gin.Context) {

	var voter voter.Voter
	if err := c.ShouldBindJSON(&voter); err != nil {
		log.Println("Error binding JSON: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := v.voterList.UpdateVoter(voter); err != nil {
		log.Println(fmt.Sprintf("Error updating Voter %v: ", voter.VoterID), err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)

}

// implementation for PUT /voters/:id/polls
// updates a voterPoll, specifically its VoteDate, of the Voter with ID id
// only one voterPoll is allowed to be updated at a time
// any data in the Voter fields outside of VoteHistory will be ignored
func (v *VoterAPI) UpdatePollData(c *gin.Context) {
	idS := c.Param("id")
	id64, err := strconv.ParseUint(idS, 10, 32)
	if err != nil {
		log.Println(fmt.Sprintf("Error converting Voter ID %v to uint64: ", idS), err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var voter voter.Voter
	if err := c.ShouldBindJSON(&voter); err != nil {
		log.Println("Error binding JSON: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := v.voterList.UpdatePollData(uint(id64), voter); err != nil {
		log.Println(fmt.Sprintf("Error updating poll in Voter %v's history: ", id64), err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)

}

// Leftover Handler from Todo

/*   SPECIAL HANDLERS FOR DEMONSTRATION - CRASH SIMULATION AND HEALTH CHECK */

// implementation for GET /crash
// This simulates a crash to show some of the benefits of the
// gin framework
func (v *VoterAPI) CrashSim(c *gin.Context) {
	//panic() is go's version of throwing an exception
	panic("Simulating an unexpected crash")
}
