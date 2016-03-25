package huddles

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/search"
	"github.com/pebbe/util"
	. "gopkg.in/check.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/dbtest"
)

type MongoSearchSuite struct {
	DBServer      *dbtest.DBServer
	Session       *mgo.Session
	MongoSearcher *search.MongoSearcher
	EST           *time.Location
	Local         *time.Location
}

var _ = Suite(&MongoSearchSuite{})

func (m *MongoSearchSuite) SetUpSuite(c *C) {
	m.EST = time.FixedZone("EST", -5*60*60)
	m.Local, _ = time.LoadLocation("Local")

	//turnOnDebugLog()

	// Register the custom search parameters
	RegisterCustomSearchDefinitions()

	// Set up the database
	m.DBServer = &dbtest.DBServer{}
	m.DBServer.SetPath(c.MkDir())

	m.Session = m.DBServer.Session()
	db := m.Session.DB("fhir-test")
	m.MongoSearcher = search.NewMongoSearcher(db)

	// Read in the data in FHIR format
	data, err := ioutil.ReadFile("../fixtures/huddle.json")
	util.CheckErr(err)

	// Put the huddle into the database
	var huddle models.Group
	err = json.Unmarshal(data, &huddle)
	util.CheckErr(err)
	err = db.C("groups").Insert(&huddle)
	util.CheckErr(err)
}

func (m *MongoSearchSuite) TearDownSuite(c *C) {
	m.Session.Close()
	m.DBServer.Wipe()
	m.DBServer.Stop()
}

func turnOnDebugLog() {
	mgo.SetDebug(true)
	var aLogger *log.Logger
	aLogger = log.New(os.Stderr, "", log.LstdFlags)
	mgo.SetLogger(aLogger)
}

func (m *MongoSearchSuite) TestSearchHuddleByLeader(c *C) {
	var huddles []*models.Group

	q := search.Query{Resource: "Group", Query: "leader=Practitioner/9999999999999999999"}
	mq := m.MongoSearcher.CreateQuery(q)
	err := mq.All(&huddles)
	util.CheckErr(err)
	c.Assert(huddles, HasLen, 1)

	// Need to update id and lastUpdated on expected in order to test the match correctly
	expected := newExampleHuddle()
	expected.Id = huddles[0].Id
	expected.Meta.LastUpdated = huddles[0].Meta.LastUpdated
	assertDeepEqualHuddles(c, huddles[0], expected)

	// This should not match anything
	q = search.Query{Resource: "Group", Query: "leader=Practitioner/8888888888888888888"}
	mq = m.MongoSearcher.CreateQuery(q)
	count, err := mq.Count()
	util.CheckErr(err)
	c.Assert(count, Equals, 0)
}

func (m *MongoSearchSuite) TestSearchHuddleByActiveDateTime(c *C) {
	q := search.Query{Resource: "Group", Query: "activedatetime=2016"}
	mq := m.MongoSearcher.CreateQuery(q)
	count, err := mq.Count()
	util.CheckErr(err)
	c.Assert(count, Equals, 1)

	q = search.Query{Resource: "Group", Query: "activedatetime=2016-02"}
	mq = m.MongoSearcher.CreateQuery(q)
	count, err = mq.Count()
	util.CheckErr(err)
	c.Assert(count, Equals, 1)

	q = search.Query{Resource: "Group", Query: "activedatetime=2016-02-02"}
	mq = m.MongoSearcher.CreateQuery(q)
	count, err = mq.Count()
	util.CheckErr(err)
	c.Assert(count, Equals, 1)

	q = search.Query{Resource: "Group", Query: "activedatetime=lt2016-02-15"}
	mq = m.MongoSearcher.CreateQuery(q)
	count, err = mq.Count()
	util.CheckErr(err)
	c.Assert(count, Equals, 1)

	q = search.Query{Resource: "Group", Query: "activedatetime=lt2016-02-01"}
	mq = m.MongoSearcher.CreateQuery(q)
	count, err = mq.Count()
	util.CheckErr(err)
	c.Assert(count, Equals, 0)
}

func (m *MongoSearchSuite) TestSearchHuddleByMemberReviewedDate(c *C) {
	q := search.Query{Resource: "Group", Query: "member-reviewed=2016-02-02"}
	mq := m.MongoSearcher.CreateQuery(q)
	count, err := mq.Count()
	util.CheckErr(err)
	c.Assert(count, Equals, 1)

	q = search.Query{Resource: "Group", Query: "member-reviewed=2016-02-02T09:08:15Z"}
	mq = m.MongoSearcher.CreateQuery(q)
	count, err = mq.Count()
	util.CheckErr(err)
	c.Assert(count, Equals, 1)

	q = search.Query{Resource: "Group", Query: "member-reviewed=lte2016-02-02T09:20:00Z"}
	mq = m.MongoSearcher.CreateQuery(q)
	count, err = mq.Count()
	util.CheckErr(err)
	c.Assert(count, Equals, 1)

	q = search.Query{Resource: "Group", Query: "member-reviewed=gte2016-02-02T09:20:00Z"}
	mq = m.MongoSearcher.CreateQuery(q)
	count, err = mq.Count()
	util.CheckErr(err)
	c.Assert(count, Equals, 1)

	q = search.Query{Resource: "Group", Query: "member-reviewed=lt2016-02-02T09:00:00Z"}
	mq = m.MongoSearcher.CreateQuery(q)
	count, err = mq.Count()
	util.CheckErr(err)
	c.Assert(count, Equals, 0)

	q = search.Query{Resource: "Group", Query: "member-reviewed=gt2016-02-02T10:00:00Z"}
	mq = m.MongoSearcher.CreateQuery(q)
	count, err = mq.Count()
	util.CheckErr(err)
	c.Assert(count, Equals, 0)
}
