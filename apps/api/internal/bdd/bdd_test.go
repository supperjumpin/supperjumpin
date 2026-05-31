package bdd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/cucumber/godog"
)

type testContext struct {
	baseURL      string
	token        string
	playerID     string
	groupID      string
	seasonID     string
	ideaID       string
	jumpID       string
	authID       string
	lastResponse *http.Response
	lastBody     []byte
}

func (c *testContext) request(method, path string, body interface{}) {
	var bodyReader *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewBuffer(b)
	} else {
		bodyReader = bytes.NewBuffer([]byte{})
	}

	req, _ := http.NewRequest(method, c.baseURL+path, bodyReader)
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	c.lastResponse = resp
	c.lastBody, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
}

func InitializeScenario(sc *godog.ScenarioContext) {
	baseURL := os.Getenv("BDD_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8090"
	}
	token := os.Getenv("BDD_AUTH_TOKEN")
	if token == "" {
		token = "test-token"
	}

	ctx := &testContext{
		baseURL: baseURL,
		token:   token,
	}

	// Jump lifecycle steps
	sc.Step(`^a player "([^"]*)" exists$`, ctx.playerExists)
	sc.Step(`^a group "([^"]*)" exists and "([^"]*)" is a member$`, ctx.groupExists)
	sc.Step(`^an active season exists for the group$`, ctx.seasonExists)
	sc.Step(`^"([^"]*)" creates an idea for "([^"]*)" at "([^"]*)" with "([^"]*)"$`, ctx.createIdea)
	sc.Step(`^"([^"]*)" promotes the idea to a planned jump$`, ctx.promoteIdea)
	sc.Step(`^"([^"]*)" authorizes an upload for "([^"]*)"$`, ctx.authorizeUpload)
	sc.Step(`^"([^"]*)" submits evidence with caption "([^"]*)"$`, ctx.submitEvidence)
	sc.Step(`^the jump status should be "([^"]*)"$`, ctx.verifyJumpStatus)
	sc.Step(`^the idea status should be "([^"]*)"$`, ctx.verifyIdeaStatus)
	sc.Step(`^the request to authorize upload should fail with status (\d+)$`, ctx.authorizeUploadFails)
	sc.Step(`^the request to submit evidence without authorization should fail$`, ctx.submitEvidenceFails)

	// Judging steps
	sc.Step(`^"([^"]*)" submits a judgment with commitment (\d+), transgression (\d+), creativity (\d+), presentation (\d+)$`, ctx.submitJudgment)
	sc.Step(`^the jump final score should be (\d+)$`, ctx.verifyFinalScore)

	// Group home steps
	sc.Step(`^the group's recent jumps should include "([^"]*)"$`, ctx.verifyRecentJump)

	// Error case steps
	sc.Step(`^the request to submit judgment with commitment (\d+) should fail with status (\d+)$`, ctx.submitJudgmentFails)
}

func (c *testContext) playerExists(name string) error {
	c.request("GET", "/v1/me", nil)
	if c.lastResponse.StatusCode != http.StatusOK {
		return fmt.Errorf("player bootstrap failed: %d %s", c.lastResponse.StatusCode, string(c.lastBody))
	}
	var res struct {
		Player struct {
			ID string `json:"id"`
		} `json:"player"`
	}
	json.Unmarshal(c.lastBody, &res)
	c.playerID = res.Player.ID
	return nil
}

func (c *testContext) groupExists(groupName, playerName string) error {
	c.request("POST", "/v1/groups", map[string]string{"name": groupName})
	if c.lastResponse.StatusCode != http.StatusCreated {
		return fmt.Errorf("group creation failed: %d %s", c.lastResponse.StatusCode, string(c.lastBody))
	}
	var res struct {
		Group struct {
			ID string `json:"id"`
		} `json:"group"`
	}
	json.Unmarshal(c.lastBody, &res)
	c.groupID = res.Group.ID
	return nil
}

func (c *testContext) seasonExists() error {
	body := map[string]string{
		"submissionDeadline": "2099-01-01T00:00:00Z",
		"judgingDeadline":    "2099-12-31T00:00:00Z",
	}
	c.request("POST", "/v1/groups/"+c.groupID+"/seasons", body)
	if c.lastResponse.StatusCode != http.StatusCreated {
		return fmt.Errorf("season creation failed: %d %s", c.lastResponse.StatusCode, string(c.lastBody))
	}
	var res struct {
		ActiveSeason struct {
			ID string `json:"id"`
		} `json:"activeSeason"`
	}
	json.Unmarshal(c.lastBody, &res)
	c.seasonID = res.ActiveSeason.ID
	return nil
}

func (c *testContext) createIdea(player, source, dest, food string) error {
	body := map[string]string{"source": source, "destination": dest, "food": food}
	c.request("POST", "/v1/groups/"+c.groupID+"/ideas", body)
	if c.lastResponse.StatusCode != http.StatusCreated {
		return fmt.Errorf("idea creation failed: %d %s", c.lastResponse.StatusCode, string(c.lastBody))
	}
	var res struct {
		ID string `json:"id"`
	}
	json.Unmarshal(c.lastBody, &res)
	c.ideaID = res.ID
	return nil
}

func (c *testContext) promoteIdea(player string) error {
	body := map[string]bool{"offSeason": false}
	c.request("POST", "/v1/ideas/"+c.ideaID+"/planned-jump", body)
	if c.lastResponse.StatusCode != http.StatusCreated {
		return fmt.Errorf("idea promotion failed: %d %s", c.lastResponse.StatusCode, string(c.lastBody))
	}
	var res struct {
		ID string `json:"id"`
	}
	json.Unmarshal(c.lastBody, &res)
	c.jumpID = res.ID
	return nil
}

func (c *testContext) authorizeUpload(player, contentType string) error {
	body := map[string]string{"contentType": contentType}
	c.request("POST", "/v1/jumps/"+c.jumpID+"/evidence-upload-authorizations", body)
	if c.lastResponse.StatusCode != http.StatusCreated {
		return fmt.Errorf("upload auth failed: %d %s", c.lastResponse.StatusCode, string(c.lastBody))
	}
	var res struct {
		ID string `json:"id"`
	}
	json.Unmarshal(c.lastBody, &res)
	c.authID = res.ID
	return nil
}

func (c *testContext) submitEvidence(player, caption string) error {
	body := map[string]string{
		"uploadAuthorizationId": c.authID,
		"caption":               caption,
	}
	c.request("POST", "/v1/jumps/"+c.jumpID+"/evidence", body)
	if c.lastResponse.StatusCode != http.StatusCreated {
		return fmt.Errorf("evidence submission failed: %d %s", c.lastResponse.StatusCode, string(c.lastBody))
	}
	return nil
}

func (c *testContext) verifyJumpStatus(expectedStatus string) error {
	var res struct {
		Jump struct {
			Status string `json:"status"`
		} `json:"jump"`
	}
	json.Unmarshal(c.lastBody, &res)
	if res.Jump.Status != expectedStatus {
		return fmt.Errorf("expected status %q, got %q", expectedStatus, res.Jump.Status)
	}
	return nil
}

func (c *testContext) submitJudgment(player string, commitment, transgression, creativity, presentation int) error {
	body := map[string]int{
		"commitment":    commitment,
		"transgression": transgression,
		"creativity":    creativity,
		"presentation":  presentation,
	}
	c.request("POST", "/v1/jumps/"+c.jumpID+"/judgment", body)
	if c.lastResponse.StatusCode != http.StatusCreated && c.lastResponse.StatusCode != http.StatusOK {
		return fmt.Errorf("judgment submission failed: %d %s", c.lastResponse.StatusCode, string(c.lastBody))
	}
	return nil
}

func (c *testContext) verifyFinalScore(expectedScore int) error {
	var res struct {
		Jump struct {
			FinalScore *int `json:"finalScore"`
		} `json:"jump"`
	}
	json.Unmarshal(c.lastBody, &res)
	if res.Jump.FinalScore == nil {
		return fmt.Errorf("expected final score %d, got nil", expectedScore)
	}
	if *res.Jump.FinalScore != expectedScore {
		return fmt.Errorf("expected final score %d, got %d", expectedScore, *res.Jump.FinalScore)
	}
	return nil
}

func (c *testContext) verifyRecentJump(food string) error {
	c.request("GET", "/v1/groups/"+c.groupID+"/home", nil)
	if c.lastResponse.StatusCode != http.StatusOK {
		return fmt.Errorf("group home failed: %d", c.lastResponse.StatusCode)
	}
	var res struct {
		RecentJumps []struct {
			Jump struct {
				Food string `json:"food"`
			} `json:"jump"`
		} `json:"recentJumps"`
	}
	json.Unmarshal(c.lastBody, &res)
	if len(res.RecentJumps) == 0 {
		return fmt.Errorf("no recent jumps found in group home")
	}
	for _, j := range res.RecentJumps {
		if j.Jump.Food == food {
			return nil
		}
	}
	return fmt.Errorf("recent jumps do not include %q", food)
}

func (c *testContext) verifyIdeaStatus(expectedStatus string) error {
	var res struct {
		Status string `json:"status"`
	}
	json.Unmarshal(c.lastBody, &res)
	if res.Status != expectedStatus {
		return fmt.Errorf("expected status %q, got %q", expectedStatus, res.Status)
	}
	return nil
}

func (c *testContext) authorizeUploadFails(statusCode int) error {
	body := map[string]string{"contentType": "image/png"}
	c.request("POST", "/v1/jumps/"+c.ideaID+"/evidence-upload-authorizations", body)
	if c.lastResponse.StatusCode != statusCode {
		return fmt.Errorf("expected status %d, got %d", statusCode, c.lastResponse.StatusCode)
	}
	return nil
}

func (c *testContext) submitEvidenceFails() error {
	body := map[string]string{
		"uploadAuthorizationId": "fake-auth-id",
		"caption":               "No auth caption",
	}
	c.request("POST", "/v1/jumps/"+c.jumpID+"/evidence", body)
	// Should fail with 404 (auth not found) or 403 (no membership) or 400 (bad auth)
	if c.lastResponse.StatusCode < 200 || c.lastResponse.StatusCode >= 300 {
		return nil // expected
	}
	return fmt.Errorf("expected evidence submission to fail, got %d", c.lastResponse.StatusCode)
}

func (c *testContext) submitJudgmentFails(commitment int, statusCode int) error {
	body := map[string]int{
		"commitment":    commitment,
		"transgression": 5,
		"creativity":    5,
		"presentation":  5,
	}
	c.request("POST", "/v1/jumps/"+c.jumpID+"/judgment", body)
	if c.lastResponse.StatusCode != statusCode {
		return fmt.Errorf("expected status %d, got %d", statusCode, c.lastResponse.StatusCode)
	}
	return nil
}

func TestFeatures(t *testing.T) {
	opts := godog.Options{
		Format: "pretty",
		Paths:  []string{"jumps.feature", "judging.feature"},
	}

	suite := godog.TestSuite{
		Name:                "supperjumpin-bdd",
		ScenarioInitializer: InitializeScenario,
		Options:             &opts,
	}

	status := suite.Run()
	if status != 0 {
		t.Errorf("BDD tests failed with status %d", status)
	}
}