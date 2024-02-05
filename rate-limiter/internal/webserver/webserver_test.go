package webserver

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type webServerSuite struct {
	suite.Suite
}

func (suite *webServerSuite) TestWebServerRunning() {
	client := http.DefaultClient

	req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/rate-limit", nil)
	assert.NoError(suite.T(), err)

	resp, err := client.Do(req)
	assert.NoError(suite.T(), err)
	io.Copy(io.Discard, resp.Body)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	client.CloseIdleConnections()
}

func (suite *webServerSuite) TestRateLimitWithGlobalLimits() {
	client := http.DefaultClient

	okResponses := []int{}
	blockedResponses := []int{}

	req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/rate-limit", nil)
	assert.NoError(suite.T(), err)

	for i := 0; i < 100; i++ {
		resp, err := client.Do(req)
		assert.NoError(suite.T(), err)
		io.Copy(io.Discard, resp.Body)
		defer resp.Body.Close()

		switch resp.StatusCode {
		case http.StatusOK:
			okResponses = append(okResponses, resp.StatusCode)
		case http.StatusTooManyRequests:
			blockedResponses = append(blockedResponses, resp.StatusCode)
		}
	}
	assert.Equal(suite.T(), 10, len(okResponses))
	assert.Equal(suite.T(), 90, len(blockedResponses))
	client.CloseIdleConnections()
}

func (suite *webServerSuite) TestRateLimitWithAPIKey() {
	client := http.DefaultClient

	okResponses := []int{}
	blockedResponses := []int{}

	req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/rate-limit", nil)
	req.Header.Set("API_KEY", "goExpert")
	assert.NoError(suite.T(), err)

	for i := 0; i < 1000; i++ {
		resp, err := client.Do(req)
		assert.NoError(suite.T(), err)
		io.Copy(io.Discard, resp.Body)
		defer resp.Body.Close()

		switch resp.StatusCode {
		case http.StatusOK:
			okResponses = append(okResponses, resp.StatusCode)
		case http.StatusTooManyRequests:
			blockedResponses = append(blockedResponses, resp.StatusCode)
		}
	}
	assert.Equal(suite.T(), 100, len(okResponses))
	assert.Equal(suite.T(), 900, len(blockedResponses))
	client.CloseIdleConnections()
}

func TestWebServerSuite(t *testing.T) {
	suite.Run(t, new(webServerSuite))
}
