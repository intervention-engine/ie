package controllers

import (
	"net/http"
	"reflect"

	mgo "gopkg.in/mgo.v2"

	"github.com/gin-gonic/gin"
)

// IEController Wrapped for Patient Collection
type IEController struct {
	DB             *mgo.Database
	CollectionName string
	Item           interface{}
	IDScope        string
}

// All List All Items
func (ie *IEController) All(c *gin.Context) {
	items := ie.itemSlice()
	ie.getCollection().Find(nil).All(items)

	c.JSON(http.StatusOK, items)
}

// Create a Patient resource
func (ie *IEController) Create(c *gin.Context) {
	err := ie.getCollection().Insert(getJSONBody(c))
	handleMongoError(c, err)
}

// Read Find a Patient
func (ie *IEController) Read(c *gin.Context) {
	ie.getCollection().FindId(c.Param("id")).One(&ie.Item)
	c.JSON(http.StatusOK, &ie.Item)
}

// Update Update a Patient
func (ie *IEController) Update(c *gin.Context) {
	ie.getCollection().UpdateId(c.Param("id"), getJSONBody(c))
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// Delete remove a item
func (ie *IEController) Delete(c *gin.Context) {
	err := ie.getCollection().RemoveId(c.Param("id"))
	handleMongoError(c, err)
}

func (ie *IEController) getCollection() *mgo.Collection {
	return ie.DB.C(ie.CollectionName)
}

func (ie *IEController) itemSlice() interface{} {
	if ie.Item == nil {
		panic("No Item Supplied")
	}
	it := reflect.TypeOf(ie.Item)
	rSlice := reflect.MakeSlice(reflect.SliceOf(it), 0, 0).Interface()
	rSlicePtr := reflect.New(reflect.TypeOf(rSlice))
	rSlicePtr.Elem().Set(reflect.ValueOf(rSlice))
	return rSlicePtr.Interface()
}
