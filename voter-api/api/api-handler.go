package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"drexel.edu/voter-api/voter"
	"github.com/gin-gonic/gin"
)

type VoterAPI struct {
	voterList voter.VoterList
}

func NewVoterApi() *VoterAPI {
	return &VoterAPI{
		voterList: voter.VoterList{
			Voters: make(map[uint]*voter.Voter),
		},
	}
}

// The Professor's Functions

// func (v *VoterAPI) AddVoter(voterID uint, firstName, lastName string) {
// 	v.voterList.Voters[voterID] = *voter.NewVoter(voterID, firstName, lastName)
// }

// func (v *VoterAPI) AddPoll(voterID, pollID uint) {
// 	voter := v.voterList.Voters[voterID]
// 	voter.AddPoll(pollID)
// 	v.voterList.Voters[voterID] = voter
// }

// func (v *VoterAPI) GetVoterJson(voterID uint) string {
// 	voter := v.voterList.Voters[voterID]
// 	return voter.ToJson()
// }

func (v *VoterAPI) GetVoterList() voter.VoterList {
	return v.voterList
}

func (v *VoterAPI) GetVoterListJson() string {
	b, _ := json.Marshal(v.voterList)
	return string(b)
}

// THE API FUNCTIONS

// implementation for GET /voters
// returns all Voters
func (v *VoterAPI) GetAllVoters(c *gin.Context) {
	voters := v.voterList.GetAllVoters()
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

	c.JSON(http.StatusOK, *voter)
}

// implementation for POST /voters/:id
// adds a new Voter
func (v *VoterAPI) AddVoter(c *gin.Context) {
	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Println("Error reading in JSON request body: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	voter, err := v.voterList.UnmarshalVoter(jsonData)
	if err != nil {
		log.Println("Error unmarshalling JSON: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := v.voterList.AddVoter(voter); err != nil {
		log.Println(fmt.Sprintf("Error adding Voter with the ID %v: ", voter.GetVoterID()), err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, *voter)
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

	c.JSON(http.StatusOK, voter.GetVoteHistory())
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

	voter, err := v.voterList.GetVoter(uint(id64))
	if err != nil {
		log.Println(fmt.Sprintf("Voter with the ID %v not found: ", id64), err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	pollidS := c.Param("pollid")
	pollid64, err := strconv.ParseUint(pollidS, 10, 32)
	if err != nil {
		log.Println(fmt.Sprintf("Error converting poll ID %v to uint64: ", pollidS), err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	poll, err := voter.GetVoterPoll(uint(pollid64))
	if err != nil {
		log.Println(fmt.Sprintf("Error finding poll with ID %v in Voter %v's history: ", pollid64, id64), err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, poll)
}

// implementation for POST /voters/:id/polls/:pollid
// adds the poll data (voterPoll) for the Voter with ID id and voterPoll pollid
func (v *VoterAPI) AddPollData(c *gin.Context) {

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

	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Println("Error reading in JSON request body: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	poll, err := v.voterList.UnmarshalVoterPoll(jsonData)

	if err := voter.AddVoterPoll(poll); err != nil {
		log.Println(fmt.Sprintf("Error adding poll with the ID %v: ", poll.GetPollID), err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, poll)
}

// implementation of GET /voters/health
func (v *VoterAPI) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK,
		gin.H{
			"status":             "ok",
			"version":            "1.0.0",
			"uptime":             100,
			"users_processed":    1000,
			"errors_encountered": 10,
		})
}

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
		c.AbortWithStatus(http.StatusInternalServerError)
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

	voter, err := v.voterList.GetVoter(uint(id64))
	if err != nil {
		log.Println(fmt.Sprintf("Voter with the ID %v not found: ", id64), err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	pollidS := c.Param("pollid")
	pollid64, err := strconv.ParseUint(pollidS, 10, 32)
	if err != nil {
		log.Println(fmt.Sprintf("Error converting poll ID %v to uint64: ", pollidS), err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	err = voter.DeleteVoterPoll(uint(pollid64))
	if err != nil {
		log.Println(fmt.Sprintf("Error finding poll with ID %v in Voter %v's history: ", pollid64, id64), err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.Status(http.StatusOK)
}

func (v *VoterAPI) UpdateVoter(c *gin.Context) {
	// TODO
	c.Status(http.StatusOK)

}

func (v *VoterAPI) UpdatePollData(c *gin.Context) {
	// TODO
	c.Status(http.StatusOK)

}

// implementation for GET /v2/todo
// returns todos that are either done or not done
// depending on the value of the done query parameter
// for example, /v2/todo?done=true will return all
// todos that are done.  Note you can have multiple
// query parameters, for example /v2/todo?done=true&foo=bar
// func (td *VoterAPI) ListSelectTodos(c *gin.Context) {
// 	//lets first load the data
// 	todoList, err := td.db.GetAllItems()
// 	if err != nil {
// 		log.Println("Error Getting Database Items: ", err)
// 		c.AbortWithStatus(http.StatusNotFound)
// 		return
// 	}
// 	//If the database is empty, make an empty slice so that the
// 	//JSON marshalling works correctly
// 	if todoList == nil {
// 		todoList = make([]db.ToDoItem, 0)
// 	}

// 	//Note that the query parameter is a string, so we
// 	//need to convert it to a bool
// 	doneS := c.Query("done")

// 	//if the doneS is empty, then we will return all items
// 	if doneS == "" {
// 		c.JSON(http.StatusOK, todoList)
// 		return
// 	}

// 	//Now we can handle the case where doneS is not empty
// 	//and we need to filter the list based on the doneS value

// 	done, err := strconv.ParseBool(doneS)
// 	if err != nil {
// 		log.Println("Error converting done to bool: ", err)
// 		c.AbortWithStatus(http.StatusBadRequest)
// 		return
// 	}

// 	//Now we need to filter the list based on the done value
// 	//that was passed in.  We will create a new slice and
// 	//only add items that match the done value
// 	var filteredList []db.ToDoItem
// 	for _, item := range todoList {
// 		if item.IsDone == done {
// 			filteredList = append(filteredList, item)
// 		}
// 	}

// 	//Note that the database returns a nil slice if there are no items
// 	//in the database.  We need to convert this to an empty slice
// 	//so that the JSON marshalling works correctly.  We want to return
// 	//an empty slice, not a nil slice. This will result in the json being []
// 	if filteredList == nil {
// 		filteredList = make([]db.ToDoItem, 0)
// 	}

// 	c.JSON(http.StatusOK, filteredList)
// }

// implementation for PUT /todo
// Web api standards use PUT for Updates
// func (td *VoterAPI) UpdateToDo(c *gin.Context) {
// 	var todoItem db.ToDoItem
// 	if err := c.ShouldBindJSON(&todoItem); err != nil {
// 		log.Println("Error binding JSON: ", err)
// 		c.AbortWithStatus(http.StatusBadRequest)
// 		return
// 	}

// 	if err := td.db.UpdateItem(todoItem); err != nil {
// 		log.Println("Error updating item: ", err)
// 		c.AbortWithStatus(http.StatusInternalServerError)
// 		return
// 	}

// 	c.JSON(http.StatusOK, todoItem)
// }

// implementation for DELETE /todo
// deletes all todos
// func (td *VoterAPI) DeleteAllToDo(c *gin.Context) {

// 	if err := td.db.DeleteAll(); err != nil {
// 		log.Println("Error deleting all items: ", err)
// 		c.AbortWithStatus(http.StatusInternalServerError)
// 		return
// 	}

// 	c.Status(http.StatusOK)
// }

/*   SPECIAL HANDLERS FOR DEMONSTRATION - CRASH SIMULATION AND HEALTH CHECK */

// implementation for GET /crash
// This simulates a crash to show some of the benefits of the
// gin framework
func (v *VoterAPI) CrashSim(c *gin.Context) {
	//panic() is go's version of throwing an exception
	panic("Simulating an unexpected crash")
}
