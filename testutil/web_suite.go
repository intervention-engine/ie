package testutil

import (
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

// WebSuite Base Suite for web Handlers
type WebSuite struct {
	suite.Suite
	E *gin.Engine
}

// LoadGin initialize gin API
func (suite *WebSuite) LoadGin() *gin.RouterGroup {
	gin.SetMode(gin.TestMode)
	suite.E = gin.Default()
	return suite.E.Group("/api")
}

// AssertGetRequest Create a GET request and returns a response recorder for testing
func (suite *WebSuite) AssertGetRequest(path string, httpCode int) *httptest.ResponseRecorder {
	return suite.assertRequest(http.MethodGet, path, nil, httpCode)
}

// AssertPostRequest Create a POST request and returns a reponse recorder for testing
func (suite *WebSuite) AssertPostRequest(path string, body io.Reader, httpCode int) *httptest.ResponseRecorder {
	return suite.assertRequest(http.MethodPost, path, body, httpCode)
}

// AssertPuttRequest Create a PUT request and returns a reponse recorder for testing
func (suite *WebSuite) AssertPutRequest(path string, body io.Reader, httpCode int) *httptest.ResponseRecorder {
	return suite.assertRequest(http.MethodPut, path, body, httpCode)
}

func (suite *WebSuite) assertRequest(method string, path string, body io.Reader, httpCode int) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		suite.T().Errorf("failed to make request: %#+v\n", err)
	}
	w := httptest.NewRecorder()
	suite.E.ServeHTTP(w, req)
	suite.T().Logf("body: %v\n", w.Body.String())
	suite.Assert().Equal(httpCode, w.Code)

	return w
}
