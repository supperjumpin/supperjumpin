package bdd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/cucumber/godog"
)

type testContext struct {
	baseURL    string
	token       string
	playerID    string
	groupID     string
	seasonID    string
	ideaID      string
	jumpID      string
	authID      string
	lastResponse *http.Response
	lastBody     []byte
}

func (c *testContext) request(method, path string, body interface{}, token string) {
	var bodyReader *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewBuffer(b)
	} else {
		bodyReader = bytes.NewBuffer([]byte{})
	}

	req, _ := http.NewRequest(method, c.baseURL+path, bodyReader)
	req.Header.Set("Authorization", "Bearer "+token)
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
	ctx := &testContext{
		baseURL: "http://localhost:8080",
		token:    "dev-token",
	}

	sc.Step(`^a player "([^"]*)" exists$`, ctx.playerExists)
	sc.Step(`^a group "([^"]*)" exists and "([^"]*)" is a member$`, ctx.groupExists)
	sc.Step(`^an active season exists for the group$`, func() error {
		return ctx.seasonExists("")
	})
	sc.Step(`^"([^"]*)" creates an idea for "([^"]*)" at "([^"]*)" with "([^"]*)"$`, ctx.createIdea)
	sc.Step(`^"([^"]*)" promotes the idea to a planned jump$`, ctx.promoteIdea)
	sc.Step(`^"([^"]*)" authorizes an upload for "([^"]*)"$`, ctx.authorizeUpload)
	sc.Step(`^"([^"]*)" submits evidence with caption "([^"]*)"$`, ctx.submitEvidence)
	sc.Step(`^the jump status should be "([^"]*)"$`, ctx.verifyJumpStatus)
}

func (c *testContext) playerExists(name string) error {
	c.request("GET", "/v1/me", nil, c.token)
	if c.lastResponse.StatusCode != http.StatusOK {
		return fmt.Errorf("player bootstrap failed: %d", c.lastResponse.StatusCode)
	}
	var res struct { Player struct { ID string `json:\"id\"` } `json:\"player\"` }
	json.Unmarshal(c.lastBody, &res)
	c.playerID = res.Player.ID
	return nil
}

func (c *testContext) groupExists(groupName, playerName string) error {
	c.request("POST", "/v1/groups", map[string]string{"name": groupName}, c.token)
	if c.lastResponse.StatusCode != http.StatusCreated {
		return fmt.Errorf("group creation failed: %d", c.lastResponse.StatusCode)
	}
	var res struct { Group struct { ID string `json:\"id\"` } `json:\"group\"` }
	json.Unmarshal(c.lastBody, &res)
	c.groupID = res.Group.ID
	return nil
}

func (c *testContext) seasonExists(name string) error {
	body := map[string]string{
		"submissionDeadline": "2099-01-01T00:00:00Z",
		"judgingDeadline":    "2099-01-02T00:00:00Z",
	}
	c.request("POST", "/v1/groups/"+c.groupID+"/seasons", body, c.token)
	if c.lastResponse.StatusCode != http.StatusCreated {
		return fmt.Errorf("season creation failed: %d", c.lastResponse.StatusCode)
	}
	var res struct { ActiveSeason struct { ID string `json:\"id\"` } `json:\"activeSeason\"` }
	json.Unmarshal(c.lastBody, &res)
	c.seasonID = res.ActiveSeason.ID
	return nil
}

func (c *testContext) createIdea(player, source, dest, food string) error {
	body := map[string]string{"source": source, "destination": dest, "food": food}
	c.request("POST", "/v1/groups/"+c.groupID+"/ideas", body, c.token)
	if c.lastResponse.StatusCode != http.StatusCreated {
		return fmt.Errorf("idea creation failed: %d", c.lastResponse.StatusCode)
	}
	var res struct { ID string `json:\"id\"` }
	json.Unmarshal(c.lastBody, &res)
	c.ideaID = res.ID
	return nil
}

func (c *testContext) promoteIdea(player string) error {
	body := map[string]bool{"offSeason": false}
	c.request("POST", "/v1/ideas/"+c.ideaID+"/planned-jump", body, c.token)
	if c.lastResponse.StatusCode != http.StatusCreated {
		return fmt.Errorf("idea promotion failed: %d", c.lastResponse.StatusCode)
	}
	var res struct { ID string `json:\"id\"` }
	json.Unmarshal(c.lastBody, &res)
	c.jumpID = res.ID
	return nil
}

func (c *testContext) authorizeUpload(player, contentType string) error {
	body := map[string]string{"contentType": contentType}
	c.request("POST", "/v1/jumps/"+c.jumpID+"/evidence-upload-authorizations", body, c.token)
	if c.lastResponse.StatusCode != http.StatusCreated {
		return fmt.Errorf("upload auth failed: %d", c.lastResponse.StatusCode)
	}
	var res struct { ID string `json:\"id\"` }
	json.Unmarshal(c.lastBody, &res)
	c.authID = res.ID
	return nil
}

func (c *testContext) submitEvidence(player, caption string) error {
	body := map[string]string{
		"uploadAuthorizationId": c.authID,
		"caption":               caption,
	}
	c.request("POST", "/v1/jumps/"+c.jumpID+"/evidence", body, c.token)
	if c.lastResponse.StatusCode != http.StatusCreated {
		return fmt.Errorf("evidence submission failed: %d", c.lastResponse.StatusCode)
	}
	return nil
}

func (c *testContext) verifyJumpStatus(expectedStatus string) error {
	var res struct { Jump struct { Status string `json:\"status\"` } `json:\"jump\"` }
	json.Unmarshal(c.lastBody, &res)
	if res.Jump.Status != expectedStatus {
		return fmt.Errorf("expected status %s, got %s", expectedStatus, res.Jump.Status)
	}
	return nil
}

func TestFeatures(t *testing.T) {
	opts := godog.Options{
		Format: "pretty",
		Paths:  []string{"jumps.feature"},
	}

	suite := godog.TestSuite{
		Name: "supperjumpin-bdd",
		ScenarioInitializer: InitializeScenario,
		Options: &opts,
	}

	status := suite.Run()
	if status != 0 {
		t.Errorf("tests failed with status %d", status)
	}
}
