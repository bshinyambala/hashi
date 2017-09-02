package main

import (
	"crypto"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var JStore = CreateJobStore()

var waitGroup sync.WaitGroup

func main() {
	server := &http.Server{Addr: ":8080"}
	fmt.Println("Starting server on :8080")

	http.HandleFunc("/", HashAPI)
	http.HandleFunc("/stats", StatsAPI)

	var shutdownChannel = make(chan os.Signal)
	signal.Notify(shutdownChannel, syscall.SIGTERM)
	signal.Notify(shutdownChannel, syscall.SIGINT)

	go func() {
		if error := server.ListenAndServe(); error != nil {
			fmt.Println("Server no longer accepts new connections.")
		}
	}()

	sig := <-shutdownChannel
	fmt.Printf("Received Shutdown signal: %+v\n", sig)
	if err := server.Shutdown(nil); err != nil {
		fmt.Println("Failed to shutdown")
	}
	fmt.Println("Waiting for jobs to complete.")
	waitGroup.Wait()
	fmt.Println("Server Shutdown...Adios Amigo!")
}

/*HashAPI handles all requests to /hash */
func HashAPI(res http.ResponseWriter, req *http.Request) {

	if !strings.HasPrefix(req.URL.Path, "/hash") {
		http.Error(res, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	start := time.Now()

	switch req.Method {
	case "POST":
		password := req.FormValue("password")
		if password == "" {
			http.Error(res, "Bad Request. Password is empty or missing.", http.StatusBadRequest)
			return
		}
		jobID := JStore.CreateJob()
		fmt.Fprintf(res, strconv.Itoa(int(jobID)))

		// join wait group
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			time.Sleep(time.Second * time.Duration(5))

			hash := GenerateHash(password)
			JStore.RecordHash(jobID, hash, int32(time.Since(start).Seconds()*1000))
		}()
	case "GET":

		//read and verify job id is passed
		jobStr := strings.TrimPrefix(req.URL.Path, "/hash/")
		if jobStr == "" {
			http.Error(res, "Bad Request. Path is missing job ID.", http.StatusBadRequest)
			return
		}

		// make sure job id is a valid integer
		jobID, err := strconv.ParseUint(jobStr, 10, 64)
		if err != nil {
			http.Error(res, "Bad Request. Unable to read job ID", http.StatusBadRequest)
			return
		}

		// read the harsh and respond accordingly
		hash, ok := JStore.RetrieveHash(jobID)
		if !ok {
			http.Error(res, "Could not find the job", http.StatusNotFound)
			return
		}
		fmt.Fprintf(res, hash)
	default:
		http.Error(res, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

/*StatsAPI handles requests to /stats */
func StatsAPI(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		total, average := JStore.GetStats()
		stat := stats{Total: total, Average: average}
		res.Header().Set("Content-Type", "application/json")
		json.NewEncoder(res).Encode(stat)
	default:
		http.Error(res, http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed)
	}
}

/*JobStore struct*/
type JobStore struct {
	sync.RWMutex
	jobCounter uint64
	values     map[uint64]string
	average    int32
}

/*CreateJobStore constructor */
func CreateJobStore() *JobStore {
	store := new(JobStore)
	store.jobCounter = 0
	store.values = make(map[uint64]string)
	return store
}

//CreateJob creates the job adds them to store and returns the job id
func (store *JobStore) CreateJob() uint64 {
	//TODO: Not sure if there is need to lock
	// the store here if I'm using atomic counters
	// not sure I quite understand this yet. Taking my chances..
	atomic.AddUint64(&store.jobCounter, 1)
	return atomic.LoadUint64(&store.jobCounter)

}

//RetrieveHash retrieve job id's hash value
func (store *JobStore) RetrieveHash(id uint64) (string, bool) {
	store.RLock()
	defer store.RUnlock()
	hash, ok := store.values[id]
	return hash, ok
}

//RecordHash retrieve job id's hash value
func (store *JobStore) RecordHash(jobID uint64, hash string, duration int32) {
	store.Lock()
	defer store.Unlock()
	store.values[jobID] = hash
	totalJobs := int32(len(store.values))
	//fmt.Printf("Saving job [jobID = %d, hash = %s, duration = %d]\n", jobID, hash, duration)
	atomic.StoreInt32(&store.average, ((store.getAverage()*(totalJobs-1))+duration)/totalJobs)
}

// GetStats function
func (store *JobStore) GetStats() (int, int32) {
	return store.getTotal(), store.getAverage()
}

func (store *JobStore) getAverage() int32 {
	return atomic.LoadInt32(&store.average)
}

func (store *JobStore) getTotal() int {
	return len(store.values)
}

//GenerateHash will generate a hash and remember the results given a job id and password
func GenerateHash(password string) string {
	sha512 := crypto.SHA512.New()
	sha512.Write([]byte(password))
	return base64.StdEncoding.EncodeToString(sha512.Sum(nil))
}

type stats struct {
	Total   int   `json:"total"`
	Average int32 `json:"average"`
}
