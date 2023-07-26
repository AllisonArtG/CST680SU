package voter

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

//STRUCTS

type voterPoll struct {
	PollID   uint      `json:"poll_id"`
	VoteDate time.Time `json:"vote_date"`
}

type Voter struct {
	VoterID     uint        `json:"id" binding:"required"`
	FirstName   string      `json:"first_name" binding:"required"`
	LastName    string      `json:"last_name" binding:"required"`
	VoteHistory []voterPoll `json:"vote_history"`
}

type VoterList struct {
	Voters map[uint]*Voter //A map of VoterIDs as keys and Voter structs as values
}

func NewVoter(voterID uint, first string, last string) (*Voter, error) {
	voteHistory := make([]voterPoll, 0)
	return &Voter{VoterID: voterID, FirstName: first, LastName: last, VoteHistory: voteHistory}, nil
}

// UNMARSHAL FUNCTIONS

// returns the pointer to the new Voter
func (*VoterList) UnmarshalVoter(jsonData []byte) (*Voter, error) {
	var voter Voter
	err := json.Unmarshal(jsonData, &voter)
	if err != nil {
		return &Voter{}, nil
	}
	if voter.VoteHistory == nil {
		voter.VoteHistory = make([]voterPoll, 0)
	}
	return &voter, nil
}

// returns the pointer to the new voterPoll
func (*VoterList) UnmarshalVoterPoll(jsonData []byte) (*voterPoll, error) {
	var poll voterPoll
	err := json.Unmarshal(jsonData, &poll)
	if err != nil {
		return &voterPoll{}, nil
	}
	return &poll, nil
}

// BASIC GETTERS (return struct fields as they are, meant for other packages)

// returns the Voter's VoterID
func (v *Voter) GetVoterID() uint {
	return v.VoterID
}

// returns the Voter's VoteHistory
func (v *Voter) GetVoteHistory() []voterPoll {
	return v.VoteHistory
}

// returns the voterPoll's PollID
func (vp *voterPoll) GetPollID() uint {
	return vp.PollID
}

// ALL OTHER FUNCTIONS (Ordered in appearance in the api package)

// returns all Voters (as a Slice)
func (vl *VoterList) GetAllVoters() []Voter {

	var voters []Voter

	for _, voter := range vl.Voters {
		voters = append(voters, *voter)
	}

	return voters
}

// returns the Voter with the ID voterID.
func (vl *VoterList) GetVoter(voterID uint) (*Voter, error) {

	voter, ok := vl.Voters[voterID]
	if !ok {
		return &Voter{}, errors.New(fmt.Sprintf("Voter with ID %v does not exist.", voterID))
	}

	return voter, nil
}

// AddVoter accepts a *Voter and adds it to Voters.
func (vl *VoterList) AddVoter(voter *Voter) error {

	_, ok := vl.Voters[voter.VoterID]
	if ok {
		return errors.New(fmt.Sprintf("A voter with the ID %v already exists.", voter.VoterID))
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
	vl.Voters[voter.VoterID] = voter

	return nil
}

// returns the Voter's voterPoll where the PollID matches pollID
func (v *Voter) GetVoterPoll(pollID uint) (voterPoll, error) {
	voteHistory := v.VoteHistory
	if len(voteHistory) == 0 {
		return voterPoll{}, errors.New(fmt.Sprintf("Poll with ID %v not found in voter %v's history.", v.VoterID, pollID))
	}

	relevantPolls := make([]voterPoll, 0)
	for i := 0; i < len(voteHistory); i++ {
		poll := voteHistory[i]
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

// AddVoterPoll accepts a *voterPoll and adds it to the Voter's VoteHistory
func (v *Voter) AddVoterPoll(poll *voterPoll) error {

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

// EXTRA CREDIT

// DeleteVoter accepts an voter ID and removes it from Voters.
func (vl *VoterList) DeleteVoter(voterID uint) error {

	if _, ok := vl.Voters[voterID]; ok {
		delete(vl.Voters, voterID)
		return nil
	} else {
		return errors.New(fmt.Sprintf("An voter with the ID %v does not exist, thus they cannot be removed.", voterID))
	}
}

func (v *Voter) DeleteVoterPoll(pollID uint) error {

	i := -1
	for index, poll := range v.VoteHistory {
		if poll.PollID == pollID {
			i = index
			break
		}
	}

	if i == -1 {
		return errors.New(fmt.Sprintf("Poll with ID %v does not exist in Voter %v's VoteHistory", v.VoterID, pollID))
	} else {
		v.VoteHistory = append(v.VoteHistory[:i], v.VoteHistory[i+1:]...)
		return nil
	}
}

func (vl *VoterList) UpdateVoter(voter *Voter) error {
	// TODO
	return nil
}

func (v *Voter) UpdatePollData(poll *voterPoll) error {

	for index, currPoll := range v.VoteHistory {
		if currPoll.PollID == poll.PollID {
			v.VoteHistory[index] = *poll
			return nil
		}
	}

	return errors.New(fmt.Sprintf("Poll with ID %v does not exist in Voter %v's VoteHistory", v.VoterID, poll.PollID))
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
