# Password Hashing API
REST HTTP Server
## Run Instructions
`go run server.go` to run the api server
### Running Port 
`:8080`
## Endpoint 

- /hash - supports HTTP POST and requires `password` field to be passed with the request and `Content-Type` header must be set to `application/x-www-form-urlencoded` to return a Job ID and initiate a background process that waits 5 seconds then begins SHA512 hashing the password and base64 encoding the result and then storing it with the assigned Job ID

    `curl --data "password=angryMonkey" http://localhost:8080/hash`

- /hash/{jobID} - supports HTTP GET with the Job ID to return the hashed and encoded password. 

    `curl http://localhost:8080/hash/1`

- /stats - supports HTTP GET to return the jobs that have been processed and saved and the average processing time

    `curl http://localhost:8080/stats`

## Testing
To execute tests run `go test` 