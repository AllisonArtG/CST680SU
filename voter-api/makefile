SHELL := /bin/bash

.PHONY: help
help:
	@echo "Usage make <TARGET>"
	@echo ""
	@echo "  Targets:"
	@echo "	   build				Build the voter executable"
	@echo "	   run					Run the voter program from code"
	@echo "	   run-bin				Run the voter executable"
	@echo "	   load-db				Add sample data via curl"
	@echo "	   get-by-id			Get a voter by id pass id=<id> on command line"
	@echo "	   get-all				Get all voters"
	@echo "	   update-2				Update record 2, pass a new title in using title=<title> on command line"
	@echo "	   delete-by-id			Delete a voter by id pass id=<id> on command line"
	@echo "	   build-amd64-linux	Build amd64/Linux executable"
	@echo "	   build-arm64-linux	Build arm64/Linux executable"





.PHONY: build
build:
	go build .

.PHONY: build-amd64-linux
build-amd64-linux:
	GOOS=linux GOARCH=amd64 go build -o ./voter-linux-amd64 .

.PHONY: build-arm64-linux
build-arm64-linux:
	GOOS=linux GOARCH=arm64 go build -o ./voter-linux-arm64 .

	
.PHONY: run
run:
	go run main.go

.PHONY: run-bin
run-bin:
	./voter

.PHONY: restore-db
restore-db:
	(cp ./data/voter.json.bak ./data/voter.json)

.PHONY: restore-db-windows
restore-db-windows:
	(copy.\data\voter.json.bak .\data\voter.json)

.PHONY: load-db
load-db:
	curl -d '{ "id": 1, "first_name": "John", "last_name": "Doe", "vote_history": [{"poll_id": 1, "vote_date": "2020-11-30T14:20:28.000Z"}]}' -H "Content-Type: application/json" -X POST http://localhost:1080/voters/1
	curl -d '{ "id": 2, "first_name": "Jane", "last_name": "Lain"}' -H "Content-Type: application/json" -X POST http://localhost:1080/voters/2

.PHONY: update-2
update-2:
	curl -d '{ "id": 2, "title": "$(title)", "done": false }' -H "Content-Type: application/json" -X PUT http://localhost:1080/todo 

.PHONY: get-by-id
get-by-id:
	curl -w "HTTP Status: %{http_code}\n" -H "Content-Type: application/json" -X GET http://localhost:1080/voters/$(id) 

.PHONY: get-all
get-all:
	curl -w "HTTP Status: %{http_code}\n" -H "Content-Type: application/json" -X GET http://localhost:1080/voters 

.PHONY: delete-by-id
delete-by-id:
	curl -w "HTTP Status: %{http_code}\n" -H "Content-Type: application/json" -X DELETE http://localhost:1080/voters/$(id) 

.PHONY: delete-poll-data
delete-poll-data:
	curl -w "HTTP Status: %{http_code}\n" -H "Content-Type: application/json" -X DELETE http://localhost:1080/voters/$(id)/polls/$(pollid) 

.PHONY: get-history-by-id
get-history-by-id:
	curl -w "HTTP Status: %{http_code}\n" -H "Content-Type: application/json" -X GET http://localhost:1080/voters/$(id)/polls

.PHONY: get-poll-data
get-poll-data:
	curl -w "HTTP Status: %{http_code}\n" -H "Content-Type: application/json" -X GET http://localhost:1080/voters/$(id)/polls/$(pollid)

.PHONY: get-health
get-health:
	curl -w "HTTP Status: %{http_code}\n" -H "Content-Type: application/json" -X GET http://localhost:1080/voters/health

