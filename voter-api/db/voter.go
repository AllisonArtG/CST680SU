package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// ToDoItem is the struct that represents a single ToDo item
type ToDoItem struct {
	Id     int    `json:"id"`
	Title  string `json:"title"`
	IsDone bool   `json:"done"`
}

type voterPoll struct {
	PollID   uint      `json:"poll_id"`
	VoteDate time.Time `json:"vote_date"`
}

func NewVoterPoll(pollID uint) (*voterPoll, error) {
	voteDate := time.Now()
	return &voterPoll{PollID: pollID, VoteDate: voteDate}, nil
}

type Voter struct {
	VoterID     uint        `json:"id" binding:"required"`
	FirstName   string      `json:"first_name" binding:"required"`
	LastName    string      `json:"last_name" binding:"required"`
	VoteHistory []voterPoll `json:"vote_history"`
}

func NewVoter(voterID uint, first string, last string) (*Voter, error) {
	voteHistory := make([]voterPoll, 0)
	return &Voter{VoterID: voterID, FirstName: first, LastName: last, VoteHistory: voteHistory}, nil
}

func (v *Voter) GetVoterID() uint {
	return v.VoterID
}

func (v *Voter) GetVoterHistory() []voterPoll {
	if v.VoteHistory == nil {
		return make([]voterPoll, 0)
	} else {
		return v.VoteHistory
	}
}

func (v *Voter) GetVoterPoll(pollID uint) (voterPoll, error) {
	voterHistory := v.GetVoterHistory()
	if len(voterHistory) == 0 {
		return voterPoll{}, errors.New(fmt.Sprintf("Poll with ID %v not found in voter %v's history.", v.VoterID, pollID))
	}

	relevantPolls := make([]voterPoll, 0)
	for i := 0; i < len(voterHistory); i++ {
		poll := voterHistory[i]
		if poll.PollID == pollID {
			relevantPolls = append(relevantPolls, poll)
		}
	}
	if len(relevantPolls) == 0 {
		return voterPoll{}, errors.New(fmt.Sprintf("Poll with ID %v not found in voter %v's history.", v.VoterID, pollID))
	} else if len(relevantPolls) > 1 {
		return voterPoll{}, errors.New(fmt.Sprintf("There is an error with the internal state. Voter %v was allowed to vote more than once in poll %v.", v.VoterID, pollID))
	} else {
		return relevantPolls[0], nil
	}

}

func (p *voterPoll) GetPollID() uint {
	return p.PollID
}

func (v *Voter) AddVoterPoll(poll *voterPoll) error {
	if v.VoteHistory == nil {
		v.VoteHistory = make([]voterPoll, 0)
	}

	if len(v.VoteHistory) != 0 {
		for i := 0; i < len(v.VoteHistory); i++ {
			currPoll := v.VoteHistory[i]
			if currPoll.PollID == poll.PollID {
				return errors.New(fmt.Sprintf("Poll with ID %v already exists in Voter %v's VoteHistory. Voters are only allowed to vote once per poll.", poll.PollID, v.VoterID))
			}
		}
	}
	v.VoteHistory = append(v.VoteHistory, *poll)
	return nil
}

type VoterList struct {
	Voters map[uint]Voter //A map of VoterIDs as keys and Voter structs as values
}

// constructor for VoterList struct
func NewVoterList() (*VoterList, error) {
	voters := make(map[uint]Voter)
	return &VoterList{Voters: voters}, nil

}

//Add receivers to any structs you want, but at the minimum you should add the API behavior to the
//VoterList struct as its managing the collection of voters.  Also dont forget in the constructor
//that you need to make the map before you can use it - make map[uint]Voter

//------------------------------------------------------------
// THESE ARE THE PUBLIC FUNCTIONS THAT SUPPORT OUR VOTER APP
//------------------------------------------------------------

// AddItem accepts a *Voter and adds it to the DB.
// Preconditions:   (1) The database file must exist and be a valid
//
//					(2) The item must not already exist in the DB
//	    				because we use the item.Id as the key, this
//						function must check if the item already
//	    				exists in the DB, if so, return an error
//
// Postconditions:
//
//	 (1) The item will be added to the DB
//		(2) The DB file will be saved with the item added
//		(3) If there is an error, it will be returned
func (vl *VoterList) AddVoter(voter *Voter) error {

	//Before we add a voter to the DB, lets make sure
	//it does not exist, if it does, return an error
	_, ok := vl.Voters[voter.VoterID]
	if ok {
		return errors.New(fmt.Sprintf("A voter with the ID %v already exists.", voter.VoterID))
	}

	if voter.VoteHistory == nil {
		voter.VoteHistory = make([]voterPoll, 0)
	}

	if len(voter.VoteHistory) > 1 {
		voteIDs := make(map[uint]int)
		for i := 0; i < len(voter.VoteHistory); i++ {
			poll := voter.VoteHistory[i]
			_, ok := voteIDs[poll.PollID]
			if ok {
				return errors.New(fmt.Sprintf("The voter %v provided has multiple votes per the poll %v", voter.VoterID, poll.PollID))
			} else {
				voteIDs[poll.PollID] = 1
			}
		}
	}
	//Now that we know the voter doesn't already exist, lets add it to our map
	vl.Voters[voter.VoterID] = *voter

	//If everything is ok, return nil for the error
	return nil
}

// DeleteItem accepts an item id and removes it from the DB.
// Preconditions:   (1) The database file must exist and be a valid
//
//					(2) The item must exist in the DB
//	    				because we use the item.Id as the key, this
//						function must check if the item already
//	    				exists in the DB, if not, return an error
//
// Postconditions:
//
//	 (1) The item will be removed from the DB
//		(2) The DB file will be saved with the item removed
//		(3) If there is an error, it will be returned
func (vl *VoterList) DeleteVoter(voterID uint) error {

	// we should check if the voter exists before trying to delete it
	// this is a good practice, return an error if the
	// item does not exist

	//Now lets use the built-in go delete() function to remove
	//the item from our map

	if _, ok := vl.Voters[voterID]; ok {
		delete(vl.Voters, voterID)
		return nil
	} else {
		return errors.New(fmt.Sprintf("An voter with the ID %v does not exist, thus they cannot be removed.", voterID))
	}
}

// DeleteAllVoters removes all voters from VoterList.
// It will be exposed via a DELETE /todo endpoint
func (vl *VoterList) DeleteAll() error {
	//To delete everything, we can just create a new map
	//and assign it to our existing map.  The garbage collector
	//will clean up the old map for us
	vl.Voters = make(map[uint]Voter)

	return nil
}

// UpdateItem accepts a ToDoItem and updates it in the DB.
// Preconditions:   (1) The database file must exist and be a valid
//
//					(2) The item must exist in the DB
//	    				because we use the item.Id as the key, this
//						function must check if the item already
//	    				exists in the DB, if not, return an error
//
// Postconditions:
//
//	 (1) The item will be updated in the DB
//		(2) The DB file will be saved with the item updated
//		(3) If there is an error, it will be returned
// func (vl *VoterList) UpdateItem(item ToDoItem) error {

// 	// Check if item exists before trying to update it
// 	// this is a good practice, return an error if the
// 	// item does not exist
// 	_, ok := vl.toDoMap[item.Id]
// 	if !ok {
// 		return errors.New("item does not exist")
// 	}

// 	//Now that we know the item exists, lets update it
// 	t.toDoMap[item.Id] = item

// 	return nil
// }

// GetItem accepts an item id and returns the item from the DB.
// Preconditions:   (1) The database file must exist and be a valid
//
//					(2) The item must exist in the DB
//	    				because we use the item.Id as the key, this
//						function must check if the item already
//	    				exists in the DB, if not, return an error
//
// Postconditions:
//
//	 (1) The item will be returned, if it exists
//		(2) If there is an error, it will be returned
//			along with an empty ToDoItem
//		(3) The database file will not be modified
func (vl *VoterList) GetVoter(voterID uint) (Voter, error) {

	// Check if item exists before trying to get it
	// this is a good practice, return an error if the
	// item does not exist
	voter, ok := vl.Voters[voterID]
	if !ok {
		return Voter{}, errors.New(fmt.Sprintf("Voter with ID %v does not exist.", voterID))
	}

	return voter, nil
}

// ChangeItemDoneStatus accepts an item id and a boolean status.
// It returns an error if the status could not be updated for any
// reason.  For example, the item itself does not exist, or an
// IO error trying to save the updated status.

// Preconditions:   (1) The database file must exist and be a valid
//
//					(2) The item must exist in the DB
//	    				because we use the item.Id as the key, this
//						function must check if the item already
//	    				exists in the DB, if not, return an error
//
// Postconditions:
//
//	 (1) The items status in the database will be updated
//		(2) If there is an error, it will be returned.
//		(3) This function MUST use existing functionality for most of its
//			work.  For example, it should call GetItem() to get the item
//			from the DB, then it should call UpdateItem() to update the
//			item in the DB (after the status is changed).
// func (t *ToDo) ChangeItemDoneStatus(id int, value bool) error {

// 	//update was successful
// 	return errors.New("not implemented")
// }

// GetAllItems returns all items from the DB.  If successful it
// returns a slice of all of the items to the caller
// Preconditions:   (1) The database file must exist and be a valid
//
// Postconditions:
//
//	 (1) All items will be returned, if any exist
//		(2) If there is an error, it will be returned
//			along with an empty slice
//		(3) The database file will not be modified
func (vl *VoterList) GetAllVoters() ([]Voter, error) {

	//Now that we have the DB loaded, lets crate a slice
	var voters []Voter

	//Now lets iterate over our map and add each item to our slice
	for _, voter := range vl.Voters {
		voters = append(voters, voter)
	}

	//Now that we have all of our items in a slice, return it
	return voters, nil
}

// PrintItem accepts a ToDoItem and prints it to the console
// in a JSON pretty format. As some help, look at the
// json.MarshalIndent() function from our in class go tutorial.
func (vl *VoterList) PrintVoter(voter Voter) {
	jsonBytes, _ := json.MarshalIndent(voter, "", "  ")
	fmt.Println(string(jsonBytes))
}

// PrintAllItems accepts a slice of ToDoItems and prints them to the console
// in a JSON pretty format.  It should call PrintItem() to print each item
// versus repeating the code.
func (vl *VoterList) PrintAllItems(voters []Voter) {
	for _, voter := range voters {
		vl.PrintVoter(voter)
	}
}

// JsonToItem accepts a json string and returns a ToDoItem
// This is helpful because the CLI accepts todo items for insertion
// and updates in JSON format.  We need to convert it to a ToDoItem
// struct to perform any operations on it.
func (vl *VoterList) JsonToItem(jsonString string) (Voter, error) {
	var voter Voter
	err := json.Unmarshal([]byte(jsonString), &voter)
	if err != nil {
		return Voter{}, err
	}

	return voter, nil
}
