# Containerized Voter API

This application uses the Golang Gin framework to create a Voter API.

It keeps `Voter`s which are representations of a voter. Each `Voter` has the following fields: 
`VoterID`, `FirstName`, `LastName`, `VoteHistory`. `VoteHistory` stores the poll data for different
polls the `Voter` has voted in.  

In this version, the data is stored in a Redis database. There are two containers the container for the Voter API (`voter-api-2`) and the one for the Redis database (`voter-cache`).

## To Run

NOTE: This version (v2) of the Voter API does not persist changes to the database. If the Redis Container goes down, so will the data.

### Run the following scripts (order does matter here)

```
./start-redis.sh
./build-better-docker.sh
./run-better-docker.sh
```

### Alternatively in One Step

```
docker compose up
```


## The Makefile

Everything remains the same as the previous assignment. 

To see everything you can do you can just run `make` and some of the make targets take parameters.

```
➜  todo-api git:(main) make
Usage make <TARGET>

  Targets:

          build                         Build the voter executable
          build-amd64-linux	            Build amd64/Linux executable
          build-arm64-linux	            Build arm64/Linux executable
          run                           Run the voter program from code
          run-bin                       Run the voter executable
          load-db                       Add sample data via curl
          get-all-voters                Get all voters
          get-voter-by-id               Get a voter by id pass id=<id> on command line
          add-voter                     Add a voter pass voter=<voter> on command line"
                                        e.g. voter='{"VoterID": 3, "FirstName": "James", "LastName": "Liu"}'
                                        Ignores any data in VoteHistory and simply initializes an empty slice
          get-history-by-id             Get a voter's voting history by id pass id=<id> on command line
          get-poll-data                 Get a voter's poll data pass id=<id> pollid=<pollid> on command line
          add-poll-data                 Add poll data to a voter's history pass id=<id> voter=<voter> on    
                                        command line
                                        e.g. id=3 voter='{[{"PollID": 4, "VoteDate": "2022-11-30T14:20:28.000Z"}'
                                        Any other voter fields will be ignored (VoterID, FirstName, LastName)
                                        Only one poll data item is allowed to be added at a time
          delete-voter-by-id            Delete a voter by id pass id=<id> on command line
          delete-poll-data              Delete a voter's poll data pass id=<id> pollid=<pollid> on command  
                                        line
          update-voter                  Update a voter pass voter=<voter> on command line"
                                        e.g. voter='{"VoterID": 3, "FirstName": "Jimmy", "LastName": "Liu"}'
                                        This only updates FirstName and Lastname, so anything in VoteHistory will be ignored
          update-poll-data              Update a voter's poll data pass id=<id> voter=<voter> on command  
                                        line
                                        e.g. id=3 voter='{[{"PollID": 4, "VoteDate": "2022-12-10T14:20:28.000Z"})'
                                        This only updates VoteDate of a poll, so any of the other in Voter fields outside of
                                        VoteHistory will be ignored
          get-health                    Get health of the Voter-API Application
```