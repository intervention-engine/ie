package controllers

import (
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/intervention-engine/ie/db"
	"github.com/intervention-engine/ie/models"
)

// Patients Wrapped for Patient Collection
type Patients struct {
	patients []models.Patient
	patient  models.Patient
}

var collection = db.GetDB().C("patients")

// All List All Patients
func (pc *Patients) All(c *gin.Context) {
	// var patients []models.Patient
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage := 10

	if page < 1 {
		page = 1
	}

	q := collection.Find(nil)
	meta := PaginationMeta(q, page, perPage)

	q.Skip(*meta.Offset).Limit(perPage).All(&pc.patients)

	c.JSON(http.StatusOK, gin.H{"patients": &pc.patients, "meta": &meta})
}

// Create a Patient resource
func (pc *Patients) Create(c *gin.Context) {
	err := collection.Insert(getJSONBody(c))
	handleMongoError(c, err)
}

// Read Find a Patient
func (pc *Patients) Read(c *gin.Context) {
	// var patient models.Patient
	collection.FindId(c.Param("id")).One(&pc.patient)
	c.JSON(http.StatusOK, gin.H{"patient": &pc.patient})
}

// Update Update a Patient
func (pc *Patients) Update(c *gin.Context) {
	// var patient models.Patient
	// collection.Update({}, update)
	collection.UpdateId(c.Param("id"), getJSONBody(c))
	c.JSON(http.StatusOK, gin.H{"patient": &pc.patient})
}

// Delete remove a patient
func (pc *Patients) Delete(c *gin.Context) {
	err := collection.RemoveId(c.Param("id"))
	handleMongoError(c, err)
}

func handleMongoError(c *gin.Context, err error) {
	if err != nil {
		c.AbortWithError(500, err)
	}
}

func getJSONBody(c *gin.Context) string {

	defer c.Request.Body.Close()
	body, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		c.AbortWithError(400, err)
	}

	return string(body)

}
