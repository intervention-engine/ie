package controllers_test

import (
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/intervention-engine/ie/controllers"
	"github.com/intervention-engine/ie/testutil"
)

type IEControllerSuite struct {
	testutil.MongoSuite
	ctx             *gin.Context
	reponseRecorder *httptest.ResponseRecorder
	target          *controllers.IEController
}

func (c *IEControllerSuite) SetupTest() {
	c.mockContext()
}

func (c *IEControllerSuite) TearDownTest() {
	c.TearDownDB()
}

func (c *IEControllerSuite) TearDownSuite() {
	c.TearDownDBServer()
}

func (c *IEControllerSuite) mockContext() {

	gin.SetMode(gin.TestMode)

	c.ctx, c.reponseRecorder, _ = gin.CreateTestContext()
}

func (c *IEControllerSuite) buildParams(params map[string]string) {
	var paramsSlice []gin.Param
	for key, value := range params {
		paramsSlice = append(paramsSlice, gin.Param{
			Key:   key,
			Value: value,
		})
	}

	c.ctx.Params = paramsSlice
}
