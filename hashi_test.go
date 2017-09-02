package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

var TestHash = "Io0yGSqR3RGSwHCW54RLn9QYJr7Ypmra2qeTTWJ7v9z8WKqpudctlPrCl+bzLSBpo2C3YR6+sMo1M1NduS6TGw=="
var TestPassword = "puppymonkeybaby"
var TestJStore = CreateJobStore()

func TestGenerateHash(t *testing.T) {
	generatedHash := GenerateHash(TestPassword)
	if TestHash != generatedHash {
		t.Error("Incorrect hash result.")
	}
}

func TestCreateJob(t *testing.T) {
	jobID := TestJStore.CreateJob()
	if jobID != 1 {
		t.Error("Job ID Must be greater than 0")
	}
}

func TestRecordAndRetrieveHash(t *testing.T) {
	jobID := TestJStore.CreateJob()
	if jobID == 0 {
		t.Error("Job ID cannot be 0.")
	}

	TestJStore.RecordHash(jobID, GenerateHash(TestPassword), 5000)

	if _, ok := TestJStore.RetrieveHash(jobID); !ok {
		t.Error("Hash not found.")
	}

	if hash, ok := TestJStore.RetrieveHash(jobID); ok {
		if TestHash != hash {
			t.Error("Hash does not match.")
		}
	}
}

func TestStats(t *testing.T) {
	lastTotal := TestJStore.getTotal()

	for i := 0; i < 100; i++ {
		jobID := TestJStore.CreateJob()
		if jobID == 0 {
			t.Error("Job ID cannot be 0.")
		}
		TestJStore.RecordHash(jobID, GenerateHash(TestPassword), 5000)
	}

	if total, average := TestJStore.GetStats(); total != 100+lastTotal || average != 5000 {
		t.Error("Incorrect statistics [Total : ", total, " , Average : ", average)
	}
}

func TestHashApiPOST(t *testing.T) {
	formData := url.Values{}
	formData.Set("password", TestPassword)
	req, err := http.NewRequest("POST", "/hash", strings.NewReader(formData.Encode()))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	recorder := httptest.NewRecorder()
	HashAPI(recorder, req)
	// Check the status code is what we expect.
	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("Got wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := `1`
	if recorder.Body.String() != expected {
		t.Errorf("Returned unexpected body: got %v want %v",
			recorder.Body.String(), expected)
	}
}
func TestHashApiPOSTMissingPassword(t *testing.T) {
	formData := url.Values{}
	req, err := http.NewRequest("POST", "/hash", strings.NewReader(formData.Encode()))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	recorder := httptest.NewRecorder()
	HashAPI(recorder, req)
	// Check the status code is what we expect.
	if status := recorder.Code; status != http.StatusBadRequest {
		t.Errorf("Got wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}

	// Check the response body is what we expect.
	body := recorder.Body.String()
	expected := `Bad Request. Password is empty or missing.`

	if strings.Trim(body, "\n") != expected {
		t.Errorf("Returned unexpected body: \ngot [%v] \nwant [%v]",
			recorder.Body.String(), expected)
	}
}

func TestHashApiGET(t *testing.T) {
	time.Sleep(time.Second * time.Duration(6))
	req, err := http.NewRequest("GET", "/hash/1", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	HashAPI(recorder, req)
	// Check the status code is what we expect.
	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("Got wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	if recorder.Body.String() != TestHash {
		t.Errorf("Returned unexpected body: got %v want %v",
			recorder.Body.String(), TestHash)
	}
}

func TestStatsApiGET(t *testing.T) {
	req, err := http.NewRequest("GET", "/stats", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	StatsAPI(recorder, req)
	// Check the status code is what we expect.
	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("Got wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	body := recorder.Body.String()
	expected := "{\"total\":1,\"average\""
	if !strings.Contains(body, expected) {
		t.Errorf("Returned unexpected body: expected %v to contain %v",
			body, expected)
	}
}
