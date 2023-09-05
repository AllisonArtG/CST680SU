package schema

import (
	"time"
)

type Vote struct {
	VoteID    string
	VoterID   string
	PollID    string
	VoteValue string
}

type VoterPoll struct {
	PollID   string
	VoteDate time.Time
}

type Voter struct {
	VoterID     string
	FirstName   string `json:",omitempty"`
	LastName    string `json:",omitempty"`
	VoteHistory []VoterPoll
}

type PollOption struct {
	PollOptionID   string
	PollOptionText string
}

type Poll struct {
	PollID       string
	PollTitle    string `json:",omitempty"`
	PollQuestion string `json:",omitempty"`
	PollOptions  []PollOption
}
