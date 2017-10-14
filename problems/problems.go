package problems

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

type CodeforcesProblem struct {
	ContestID uint    `json:"contestId"`
	Index     string  `json:"index"`
	Name      string  `json:"name"`
	Points    float32 `json:"points"`
}

type CodeforcesSubmission struct {
	ID        uint              `json:"id"`
	ContestID uint              `json:"contestId"`
	Verdict   string            `json:"verdict"`
	Problem   CodeforcesProblem `json:"problem"`
}

type CodeforcesSubmissionResponse struct {
	Status string                 `json:"status"`
	Result []CodeforcesSubmission `json:"result"`
}

// Problem describes a problem in our system.
type Problem struct {
	ID         uint
	OwnerID    int
	ExternalID string
	Title      string
	Source     string
	HasSolved  bool
	HasTried   bool
}

// DefaultDataProvider implements the DataProvider interface.
type DefaultDataProvider struct {
}

// GetSets returns all the problems in the database.
func (t *DefaultDataProvider) GetSets() ([]byte, error) {
	fileName := os.Getenv("APP_ROOT") + "/problems/fixtures/problemsets.json"
	return ioutil.ReadFile(fileName)
}

// FindProblem retrieves problem details for a given user and problem ID. For now it only works with
// Codeforces.
func (t *DefaultDataProvider) FindProblem(userID string, problemID uint) (*Problem, error) {
	problem, err := t.findProblem(problemID)
	if err != nil {
		return nil, err
	}

	// NOTE: Assuming this is a Codeforces problem for now
	externalID := problem.ExternalID
	contestID, err := strconv.Atoi(externalID[0 : len(externalID)-1])
	index := externalID[len(externalID)-1:]

	url := fmt.Sprintf(
		"http://codeforces.com/api/contest.status?contestId=%d&handle=%s&showUnofficial=true",
		contestID, userID,
	)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var submissionResponse CodeforcesSubmissionResponse
	if err := json.Unmarshal(data, &submissionResponse); err != nil {
		log.Printf("Error: %s", err)
		return nil, err
	}
	for _, submission := range submissionResponse.Result {
		if submission.Problem.Index == index {
			problem.HasTried = true
		}
		if submission.Verdict == "OK" {
			problem.HasSolved = true
		}
	}
	return problem, nil
}

func (t *DefaultDataProvider) findProblem(problemID uint) (*Problem, error) {
	fileName := os.Getenv("APP_ROOT") + "/problems/fixtures/problems.json"
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	var problems []Problem
	if err := json.Unmarshal(data, &problems); err != nil {
		return nil, err
	}

	for _, problem := range problems {
		if problem.ID == problemID {
			return &problem, nil
		}
	}
	return nil, errors.New("problem not found")
}

// NewDataProvider returns a new DataProvider instance.
func NewDataProvider() DataProvider {
	return &DefaultDataProvider{}
}
