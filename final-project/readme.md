# Containerized API System

By Allison Gong

## Description

This application uses the Golang Gin framework to create containerized APIs: Voter API, Poll API and Votes API. It uses Hypermedia to support the intra-API integration.

All the data is stored in a Redis database.

NOTE: The APIs do not persist changes to the Redis database. If the Redis Container goes down, so will the data.

```
               ┏━━━━━━━━━━━┓               
      ┌───────▶┃ Votes API ┃◀────────┐     
      ▼        ┗━━━━━━━━━━━┛         ▼     
┌───────────┐        │         ┌──────────┐
│ Voter API │        │         │ Poll API │
└───────────┘        │         └──────────┘
      │              │                │    
      ▼              ▼                ▼    
┌─────────────────────────────────────────┐
│              Cache (Redis)              │
└─────────────────────────────────────────┘
```

### The Votes API

The Votes API is the primarly API and drives the voting system. 

1. When a `Vote` is added, deleted, and updated, the Votes API queries the Voter API in order to properly update the `Voter`'s `VoteHistory`. It also validates that the provided `Vote` fields: `VoterID`, `PollID`, and `VoteValue` (`PollOptionID`) and makes sure they exist in the Poll API or Voter API before updating Redis.
2. The Votes API also has GET endpoints that essentially serve as relays to the GET endpoints of the Voter API and the Poll API. This allows the two other APIs to get the necessary information without querying each other directly.
3. The Votes API is not the master, so a real Voting Application utilizing these APIs would still need to query the other APIs to create, delete, and update a `Voter`/`Poll`.

### The Voter API

The Voter API manages all the `Voter`s.

1. The Voter API does validate that a `Poll` exists via the Votes API (relay) before adding a `voterPoll` to the `Voter`'s `VoteHistory`.

### The Poll API

The Poll API manages all the `Poll`s and their `PollOptions`.

1. The Poll API is the only API that does not rely on other APIs and thus does not do validation.

## To Run

Delete any lingering containers, particularly the Redis container. Then from the root of the current project (`CST680SU/final-project`) run the following to bring up all the containers.

```
docker compose up
```

## To Test

First import my Postman Collection `CST680SU.postman_collection.json` and my Postman Environment `CST680SU_Localhost.postman_environment.json` into Postman.

You will first need to load the test data into the Redis Database. Simply run the `Load Data` folder.

From there please run all the folders in order except the last one (`Delete Data`). Each folder performs system tests on each of the APIs, except the `Poll` folder which consists only of unit tests. It is important to run the Requests in each folder in order, as the tests rely on prior Requests in the folder.

If one of the tests (or Requests) fails, please run the `Delete Data` folder followed by the `Load Data` folder to ensure the data is correct before running the problem folder again in order to investigate what went wrong.

## The Structs

The Structs used by the APIs are the same as in my API_Design_2 document, however they are also listed below for your reference.

### Votes API

```
type Vote struct {
	VoteID    string
	VoterID   string
	PollID    string
	VoteValue string
}
```

### Voter API

```
type Voter struct {
	VoterID     string
	FirstName   string
	LastName    string
	VoteHistory []voterPoll
}

type voterPoll struct {
	PollID   string
	VoteDate time.Time
}
```

### Poll API
```
type Poll struct {
	PollID       string
	PollTitle    string
	PollQuestion string
	PollOptions  []pollOption
}

type pollOption struct {
	PollOptionID    string
	PollOptionText  string
}
```
