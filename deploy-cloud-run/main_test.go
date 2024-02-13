package main

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type webServerSuite struct {
	client *http.Client
	suite.Suite
}

func (suite *webServerSuite) SetupSuite() {
	suite.client = http.DefaultClient
}

func (suite *webServerSuite) TearDownSuite() {
	suite.client.CloseIdleConnections()
}

func (suite *webServerSuite) TestSuccessfulRequest() {
	req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/weather?cep=01001000", nil)
	assert.NoError(suite.T(), err)

	resp, err := suite.client.Do(req)
	assert.NoError(suite.T(), err)
	io.Copy(io.Discard, resp.Body)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
}

func (suite *webServerSuite) TestInvalidCepLength() {
	req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/weather?cep=0", nil)
	assert.NoError(suite.T(), err)

	resp, err := suite.client.Do(req)
	assert.NoError(suite.T(), err)
	io.Copy(io.Discard, resp.Body)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusUnprocessableEntity, resp.StatusCode)
}

func (suite *webServerSuite) TestCepNotFound() {
	req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/weather?cep=99999999", nil)
	assert.NoError(suite.T(), err)

	resp, err := suite.client.Do(req)
	assert.NoError(suite.T(), err)
	io.Copy(io.Discard, resp.Body)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusNotFound, resp.StatusCode)
}

func TestWebServerSuite(t *testing.T) {
	suite.Run(t, new(webServerSuite))
}
