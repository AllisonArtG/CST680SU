package voter

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
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

type VoterList struct {
	Voters map[uint]Voter //A map of VoterIDs as keys and Voter structs as values
}

func NewVoter(voterID uint, first string, last string) (*Voter, error) {
	voteHistory := make([]voterPoll, 0)
	return &Voter{VoterID: voterID, FirstName: first, LastName: last, VoteHistory: voteHistory}, nil
}

// returns all Voters (as a Slice)
func (vl *VoterList) GetAllVoters() []Voter {

	var voters []Voter

	for _, voter := range vl.Voters {
		voters = append(voters, voter)
	}

	return voters
}

// returns the Voter with the VoterID voterID
func (vl *VoterList) GetVoter(voterID uint) (Voter, error) {

	voter, ok := vl.Voters[voterID]
	if !ok {
		return Voter{}, errors.New(fmt.Sprintf("Voter with ID %v does not exist.", voterID))
	}

	return voter, nil
}

// AddVoter accepts a Voter and adds it to Voters.
// its VoteHistory is always initialized to an empty slice
func (vl *VoterList) AddVoter(voter Voter) error {

	_, ok := vl.Voters[voter.VoterID]
	if ok {
		return errors.New(fmt.Sprintf("A Voter with the ID %v already exists.", voter.VoterID))
	}

	if voter.VoteHistory == nil || len(voter.VoteHistory) > 0 {
		voter.VoteHistory = make([]voterPoll, 0)
	}

	vl.Voters[voter.VoterID] = voter
	return nil
}

// returns the Voter's voterPoll where the PollID matches pollID
func (vl *VoterList) GetVoterPoll(voterID, pollID uint) (voterPoll, error) {
	voter, ok := vl.Voters[voterID]
	if !ok {
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
	voter, ok := vl.Voters[voterID]
	if !ok {
		return errors.New(fmt.Sprintf("Voter with ID %v does not exist.", voterID))
	}

	if len(newVoter.VoteHistory) > 1 || len(newVoter.VoteHistory) == 0 {
		return errors.New(fmt.Sprintf("Only allowed to add one new voterPoll at a time, and %v given.", len(newVoter.VoteHistory)))
	}

	poll := newVoter.VoteHistory[0]

	if len(voter.VoteHistory) != 0 {
		for i := 0; i < len(voter.VoteHistory); i++ {
			currPoll := voter.VoteHistory[i]
			if currPoll.PollID == poll.PollID {
				return errors.New(fmt.Sprintf("Poll with ID %v already exists in Voter %v's VoteHistory. Voters are only allowed to vote once per poll.", poll.PollID, voterID))
			}
		}
	}
	voter.VoteHistory = append(voter.VoteHistory, poll)
	vl.Voters[voterID] = voter
	return nil
}

// EXTRA CREDIT

// deletes the Voter with the VoterID voterID from Voters
func (vl *VoterList) DeleteVoter(voterID uint) error {

	if _, ok := vl.Voters[voterID]; ok {
		delete(vl.Voters, voterID)
		return nil
	} else {
		return errors.New(fmt.Sprintf("An voter with the ID %v does not exist, thus they cannot be removed.", voterID))
	}
}

// deletes the voterPoll with the PollID pollID from the Voter voterID
func (vl *VoterList) DeleteVoterPoll(voterID, pollID uint) error {

	voter, ok := vl.Voters[voterID]
	if !ok {
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
	} else {
		voter.VoteHistory = append(voter.VoteHistory[:i], voter.VoteHistory[i+1:]...)
		vl.Voters[voterID] = voter
		return nil
	}
}

// updates an existing Voter with the newVoter's fields (FirstName and LastName)
func (vl *VoterList) UpdateVoter(newVoter Voter) error {

	if _, ok := vl.Voters[newVoter.VoterID]; !ok {
		return errors.New(fmt.Sprintf("The voter to be updated Voter %v, does not exist.", newVoter.VoterID))
	}

	voter := vl.Voters[newVoter.VoterID]
	if newVoter.FirstName != "" {
		voter.FirstName = newVoter.FirstName
	}
	if newVoter.LastName != "" {
		voter.LastName = newVoter.LastName
	}
	vl.Voters[newVoter.VoterID] = voter
	return nil
}

// updates an existing voterPoll in Voter voterID's VoteHistory
func (vl *VoterList) UpdatePollData(voterID uint, newVoter Voter) error {

	voter, ok := vl.Voters[voterID]
	if !ok {
		return errors.New(fmt.Sprintf("Voter with ID %v does not exist.", voterID))
	}
	if len(newVoter.VoteHistory) > 1 || len(newVoter.VoteHistory) == 0 {
		return errors.New(fmt.Sprintf("Only allowed to update one voterPoll at a time, and %v given.", len(newVoter.VoteHistory)))
	}

	newPoll := newVoter.VoteHistory[0]

	for index, currPoll := range voter.VoteHistory {
		if currPoll.PollID == newPoll.PollID {
			voter.VoteHistory[index] = newPoll
			vl.Voters[voterID] = voter
			return nil
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
