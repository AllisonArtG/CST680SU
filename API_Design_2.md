# API Design Assignment - Part 2
By Allison Gong

## Question 1

A fairly simple change that will that will help our API Architecture better conform to Hypermedia as the Engine of Application State (HATEOAS), is to simply convert all internal representations of identity (i.e. all our ID fields in our data structures) into the endpoint associated with getting (GET) that object. This means the fields `VoterID`, `VoteID`, `PollID`, and `PollOptionID` would all be represented by `string`s rather than `uint`s.

Similarly, when the user makes a GET request to the `/voters/{id}` or `/polls/{pollid}` rather than have the fields of `VoteHistory` and `PollOptions` return arrays of the `voterPoll` and `pollOptions` respectively, they should just return the endpoints as well.

For example, If I query GET `voters/1`, the JSON object that is returned should look like this:

```
{
	"VoterID":     "/voters/1",
	"FirstName":   "John",
	"LastName":    "Doe",
	"VoteHistory": "/voters/1/polls", 
}
```
If the user wants more information on the polls `"/voters/1"` has previously voted in, they can just follow the URI. Calling GET `voters/1/polls` would return something like this:

```
[
    {
        "VoterPollID" : "/voters/1/polls/1"
        "PollID": "/polls/1", 
        "VoteDate": "2023-08-21 01:40:19.235605+00:00"
    }
]
```

Finally, the user can now call the Poll API's endpoint `/polls/1` to get more information about the Poll with ID `"/polls/1"`.

Given these changes, it may also make sense to change the internal representations of these data structures accordingly.

```
type voterPoll struct {
	PollID   string
	VoteDate time.Time
}
```

#### Option 1:
```
type Voter struct {
	VoterID     string
	FirstName   string
	LastName    string
	VoteHistory string
}

type VoterPolls struct {
    Polls map[string][]voterPolls
}
```
This option matches the returned JSON objects exactly but introduces a little bit of redundancy.

#### Option 2
```
type Voter struct {
	VoterID     string
	FirstName   string
	LastName    string
	VoteHistory []voterPoll
}
```
This option eliminates any redundancy but would require VoteHistory to be changed to the endpoint string before returning the Voter JSON object to the user.

Proceeding with the 2nd Option for the remaining reworked data structures.

```
type pollOption struct {
	PollOptionID    string
	PollOptionText  string
}

type Poll struct {
	PollID       string
	PollTitle    string
	PollQuestion string
	PollOptions  []pollOption
}
```

```
type Vote struct {
	VoteID    string
	VoterID   string
	PollID    string
	VoteValue string
}
```

## Question 2

The use of endpoints/URIs as identifiers helps to somewhat decouple all these APIs from each other. Building upon the previous example, the Voter API depends on the Poll API, but now in order to verify that a specific Poll exists, it only needs to follow the endpoint "/polls/1" given. In our previous architecture, the Voter API's implementation would need to know to use PollID 1 to query the specific endpoint "/polls/{pollid}".

## Question 3

I only used the two resources you linked:

[Designing Quality APIs (Cloud Next '18)](https://www.youtube.com/watch?v=P0a7PwRNLVU)

[HATEOAS Wikipedia Page](https://en.wikipedia.org/wiki/HATEOAS)

And these additional resources:

[Hypermedia](https://en.wikipedia.org/wiki/Hypermedia)

[REST Web Services 08 - HATEOAS](https://www.youtube.com/watch?v=NK3HNEwDXUk)

