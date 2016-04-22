package huddles

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/search"
	"github.com/stretchr/testify/suite"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/dbtest"
)

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestHuddleSearchSuite(t *testing.T) {
	suite.Run(t, new(HuddleSearchSuite))
}

type HuddleSearchSuite struct {
	suite.Suite
	DBServer      *dbtest.DBServer
	DBServerPath  string
	Session       *mgo.Session
	MongoSearcher *search.MongoSearcher
	EST           *time.Location
	Local         *time.Location
}

func (m *HuddleSearchSuite) SetupSuite() {
	require := m.Require()

	m.EST = time.FixedZone("EST", -5*60*60)
	m.Local, _ = time.LoadLocation("Local")

	//turnOnDebugLog()

	// Set up the database
	var err error
	m.DBServer = &dbtest.DBServer{}
	m.DBServerPath, err = ioutil.TempDir("", "mongotestdb")
	require.NoError(err)
	m.DBServer.SetPath(m.DBServerPath)

	m.Session = m.DBServer.Session()
	db := m.Session.DB("fhir-test")
	m.MongoSearcher = search.NewMongoSearcher(db)

	// Read in the data in FHIR format
	data, err := ioutil.ReadFile("../fixtures/huddle.json")
	require.NoError(err)

	// Put the huddle into the database
	var huddle models.Group
	err = json.Unmarshal(data, &huddle)
	require.NoError(err)
	err = db.C("groups").Insert(&huddle)
	require.NoError(err)
}

func (m *HuddleSearchSuite) TearDownSuite() {
	m.Session.Close()
	m.DBServer.Wipe()
	m.DBServer.Stop()

	if err := os.RemoveAll(m.DBServerPath); err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: Error cleaning up temp directory: %s", err.Error())
	}
}

func turnOnDebugLog() {
	mgo.SetDebug(true)
	var aLogger *log.Logger
	aLogger = log.New(os.Stderr, "", log.LstdFlags)
	mgo.SetLogger(aLogger)
}

func (m *HuddleSearchSuite) TestSearchHuddleByLeader() {
	assert := m.Assert()
	require := m.Require()

	var huddles []*models.Group

	q := search.Query{Resource: "Group", Query: "leader=Practitioner/9999999999999999999"}
	mq := m.MongoSearcher.CreateQuery(q)
	err := mq.All(&huddles)
	require.NoError(err)
	assert.Len(huddles, 1)

	// Need to update id and lastUpdated on expected in order to test the match correctly
	expected := newExampleHuddle()
	expected.Id = huddles[0].Id
	expected.Meta.LastUpdated = huddles[0].Meta.LastUpdated
	assertDeepEqualHuddles(assert, expected, huddles[0])

	// This should not match anything
	q = search.Query{Resource: "Group", Query: "leader=Practitioner/8888888888888888888"}
	mq = m.MongoSearcher.CreateQuery(q)
	count, err := mq.Count()
	require.NoError(err)
	assert.Equal(0, count)
}

func (m *HuddleSearchSuite) TestSearchHuddleByActiveDateTime() {
	assert := m.Assert()
	require := m.Require()

	q := search.Query{Resource: "Group", Query: "activedatetime=2016"}
	mq := m.MongoSearcher.CreateQuery(q)
	count, err := mq.Count()
	require.NoError(err)
	assert.Equal(1, count)

	q = search.Query{Resource: "Group", Query: "activedatetime=2016-02"}
	mq = m.MongoSearcher.CreateQuery(q)
	count, err = mq.Count()
	require.NoError(err)
	assert.Equal(1, count)

	q = search.Query{Resource: "Group", Query: "activedatetime=2016-02-02"}
	mq = m.MongoSearcher.CreateQuery(q)
	count, err = mq.Count()
	require.NoError(err)
	assert.Equal(1, count)

	q = search.Query{Resource: "Group", Query: "activedatetime=lt2016-02-15"}
	mq = m.MongoSearcher.CreateQuery(q)
	count, err = mq.Count()
	require.NoError(err)
	assert.Equal(1, count)

	q = search.Query{Resource: "Group", Query: "activedatetime=lt2016-02-01"}
	mq = m.MongoSearcher.CreateQuery(q)
	count, err = mq.Count()
	require.NoError(err)
	assert.Equal(0, count)
}

func (m *HuddleSearchSuite) TestSearchHuddleByMemberReviewedDate() {
	assert := m.Assert()
	require := m.Require()

	q := search.Query{Resource: "Group", Query: "member-reviewed=2016-02-02"}
	mq := m.MongoSearcher.CreateQuery(q)
	count, err := mq.Count()
	require.NoError(err)
	assert.Equal(1, count)

	q = search.Query{Resource: "Group", Query: "member-reviewed=2016-02-02T09:08:15Z"}
	mq = m.MongoSearcher.CreateQuery(q)
	count, err = mq.Count()
	require.NoError(err)
	assert.Equal(1, count)

	q = search.Query{Resource: "Group", Query: "member-reviewed=lte2016-02-02T09:20:00Z"}
	mq = m.MongoSearcher.CreateQuery(q)
	count, err = mq.Count()
	require.NoError(err)
	assert.Equal(1, count)

	q = search.Query{Resource: "Group", Query: "member-reviewed=gte2016-02-02T09:20:00Z"}
	mq = m.MongoSearcher.CreateQuery(q)
	count, err = mq.Count()
	require.NoError(err)
	assert.Equal(1, count)

	q = search.Query{Resource: "Group", Query: "member-reviewed=lt2016-02-02T09:00:00Z"}
	mq = m.MongoSearcher.CreateQuery(q)
	count, err = mq.Count()
	require.NoError(err)
	assert.Equal(0, count)

	q = search.Query{Resource: "Group", Query: "member-reviewed=gt2016-02-02T10:00:00Z"}
	mq = m.MongoSearcher.CreateQuery(q)
	count, err = mq.Count()
	require.NoError(err)
	assert.Equal(0, count)
}
