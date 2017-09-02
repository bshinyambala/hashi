package main

import (
	"testing"
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
