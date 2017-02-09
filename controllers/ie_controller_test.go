package controllers

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/models"
	"github.com/intervention-engine/ie/testutil"
	"github.com/stretchr/testify/suite"
)

// func TestIEController_Create(t *testing.T) {
// 	type fields struct {
// 		idScope        string
// 		collectionName string
// 		itemType       reflect.Type
// 	}
// 	type args struct {
// 		c *gin.Context
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 	}{
// 	// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ie := IEController{
// 				idScope:        tt.fields.idScope,
// 				collectionName: tt.fields.collectionName,
// 				itemType:       tt.fields.itemType,
// 			}
// 			ie.Create(tt.args.c)
// 		})
// 	}
// }
//
// func TestIEController_Read(t *testing.T) {
// 	type fields struct {
// 		idScope        string
// 		collectionName string
// 		itemType       reflect.Type
// 	}
// 	type args struct {
// 		c *gin.Context
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 	}{
// 	// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ie := IEController{
// 				idScope:        tt.fields.idScope,
// 				collectionName: tt.fields.collectionName,
// 				itemType:       tt.fields.itemType,
// 			}
// 			ie.Read(tt.args.c)
// 		})
// 	}
// }
//
// func TestIEController_Update(t *testing.T) {
// 	type fields struct {
// 		idScope        string
// 		collectionName string
// 		itemType       reflect.Type
// 	}
// 	type args struct {
// 		c *gin.Context
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 	}{
// 	// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ie := IEController{
// 				idScope:        tt.fields.idScope,
// 				collectionName: tt.fields.collectionName,
// 				itemType:       tt.fields.itemType,
// 			}
// 			ie.Update(tt.args.c)
// 		})
// 	}
// }
//
// func TestIEController_Delete(t *testing.T) {
// 	type fields struct {
// 		idScope        string
// 		collectionName string
// 		itemType       reflect.Type
// 	}
// 	type args struct {
// 		c *gin.Context
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 	}{
// 	// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ie := IEController{
// 				idScope:        tt.fields.idScope,
// 				collectionName: tt.fields.collectionName,
// 				itemType:       tt.fields.itemType,
// 			}
// 			ie.Delete(tt.args.c)
// 		})
// 	}
// }

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestIEControllerSuite(t *testing.T) {
	suite.Run(t, new(IEControllerSuite))
}

type IEControllerSuite struct {
	testutil.MongoSuite
}

func (c *IEControllerSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	server.Database = c.DB()
}

func (c *IEControllerSuite) TearDownTest() {
	c.TearDownDB()
}

func (c *IEControllerSuite) TearDownSuite() {
	c.TearDownDBServer()
}

func (c *IEControllerSuite) TestAll() {
	type fields struct {
		idScope        string
		collectionName string
		item           interface{}
	}

	ctx, _, _ := gin.CreateTestContext()

	tests := []struct {
		name    string
		fields  fields
		context *gin.Context
	}{
		{"Patients", fields{collectionName: "patients", item: models.Patient{}}, ctx},
		{"Care Teams", fields{collectionName: "care_teams", item: models.CareTeam{}}, ctx},
	}

	for _, tt := range tests {
		c.T().Run(tt.name, func(t *testing.T) {
			ie := IEController{
				idScope:        "",
				db:             c.DB(),
				collectionName: tt.fields.collectionName,
				item:           tt.fields.item,
			}
			ie.All(tt.context)
		})
	}
}
