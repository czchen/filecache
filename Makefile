.DEFAULT_GOAL:=test

doc:
	godoc -http=127.0.0.1:8080

test:
	go test ./...
