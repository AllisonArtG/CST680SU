SHELL := /bin/bash

.PHONY: help
help:
	@echo "Usage make <TARGET>"
	@echo ""
	@echo "  Targets:"
	@echo "    build					Build the voter executable"
	@echo "	   build-amd64-linux			Build amd64/Linux executable"
	@echo "	   build-arm64-linux			Build arm64/Linux executable"
	@echo "	   run					Run the voter program from code"
	@echo "	   run-bin				Run the voter executable"
	@echo "	   load-db				Add sample data via curl"
	@echo "	   get-all-voters			Get all voters"
	@echo "	   get-voter-by-id			Get a voter by id pass id=<id> on command line"
	@echo "	   add-voter				Add a voter pass voter=<voter> on command line" 
	@echo "						e.g. voter='{\"VoterID\": 3, \"FirstName\": \"James\", \"LastName\": \"Liu\"}'"
	@echo "						Ignores any data in VoteHistory and simply initializes an empty slice"
	@echo "	   get-history-by-id			Get a voter's voting history by id pass id=<id> on command line"
	@echo "	   get-poll-data			Get a voter's poll data pass id=<id> pollid=<pollid> on command line"
	@echo "	   add-poll-data			Add poll data to a voter's history pass id=<id> voter=<voter> on command line"
	@echo "						e.g. id=3 voter='{[{\"PollID\": 4, \"VoteDate\": "2022-11-30T14:20:28.000Z"}'"
	@echo "						Any other voter fields will be ignored (VoterID, FirstName, LastName)"
	@echo "						Only one poll data item is allowed to be added at a time"
	@echo "	   delete-voter-by-id			Delete a voter by id pass id=<id> on command line"
	@echo "	   delete-poll-data			Delete a voter's poll data pass id=<id> pollid=<pollid> on command line"
	@echo "	   update-voter				Update a voter pass voter=<voter> on command line"
	@echo "						e.g. voter='{\"VoterID\": 3, \"FirstName\": \"Jimmy\", \"LastName\": \"Liu\"})'"
	@echo "						This only updates FirstName and Lastname, so anything in VoteHistory will be ignored"
	@echo "	   update-poll-data			Update a voter's poll data pass id=<id> voter=<voter> on command line" 
	@echo "						e.g. id=3 voter='{[{\"PollID\": 4, \"VoteDate\": "2022-12-10T14:20:28.000Z"})'"
	@echo "						This only updates VoteDate of a poll, so any of the other in Voter fields outside of"
	@echo "						VoteHistory will be ignored"






.PHONY: build
build:
	go build .

.PHONY: build-amd64-linux
build-amd64-linux:
	GOOS=linux GOARCH=amd64 go build -o ./voter-api-linux-amd64 .

.PHONY: build-arm64-linux
build-arm64-linux:
	GOOS=linux GOARCH=arm64 go build -o ./voter-api-linux-arm64 .


.PHONY: run
run:
	go run main.go

.PHONY: run-bin
run-bin:
	./voter-api

.PHONY: restore-db
restore-db:
	(cp ./data/voter.json.bak ./data/voter.json)

.PHONY: restore-db-windows
restore-db-windows:
	(copy.\data\voter.json.bak .\data\voter.json)

.PHONY: load-db
load-db:
	curl -d '{ "VoterID": 1, "FirstName": "John", "LastName": "Doe", "VoteHistory": []}' -H "Content-Type: application/json" -X POST http://localhost:1080/voters
	curl -d '{ "VoteHistory": [{"PollID": 1, "VoteDate": "2020-11-30T14:20:28.000Z"}]}'-H "Content-Type: application/json" -X POST http://localhost:1080/voters/1/polls
	curl -d '{ "VoterID": 2, "FirstName": "Jane", "LastName": "Lain"}' -H "Content-Type: application/json" -X POST http://localhost:1080/voters

.PHONY: get-all-voters
get-all-voters:
	curl -w "HTTP Status: %{http_code}\n" -H "Content-Type: application/json" -X GET http://localhost:1080/voters 

.PHONY: get-voter-by-id
get-voter-by-id:
	curl -w "HTTP Status: %{http_code}\n" -H "Content-Type: application/json" -X GET http://localhost:1080/voters/$(id) 

.PHONY: add-voter
add-voter:
	curl -d '$(voter)' -H "Content-Type: application/json" -X POST http://localhost:1080/voters

.PHONY: get-history-by-id
get-history-by-id:
	curl -w "HTTP Status: %{http_code}\n" -H "Content-Type: application/json" -X GET http://localhost:1080/voters/$(id)/polls

.PHONY: get-poll-data
get-poll-data:
	curl -w "HTTP Status: %{http_code}\n" -H "Content-Type: application/json" -X GET http://localhost:1080/voters/$(id)/polls/$(pollid)

.PHONY: add-poll-data
add-poll-data:
	curl -d '$(voter)' -H "Content-Type: application/json" -X POST http://localhost:1080/voters/$(id)/polls

.PHONY: delete-voter-by-id
delete-voter-by-id:
	curl -w "HTTP Status: %{http_code}\n" -H "Content-Type: application/json" -X DELETE http://localhost:1080/voters/$(id) 

.PHONY: delete-poll-data
delete-poll-data:
	curl -w "HTTP Status: %{http_code}\n" -H "Content-Type: application/json" -X DELETE http://localhost:1080/voters/$(id)/polls/$(pollid) 

.PHONY: update-voter
update-voter:
	curl -d '$(voter)' -H "Content-Type: application/json" -X PUT http://localhost:1080/voters 

.PHONY: update-poll-data
update-poll-data:
	curl -d '$(voter)' -H "Content-Type: application/json" -X PUT http://localhost:1080/voters/$(id)/polls

.PHONY: get-health
get-health:
	curl -w "HTTP Status: %{http_code}\n" -H "Content-Type: application/json" -X GET http://localhost:1080/voters/health

