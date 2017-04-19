package testutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/stretchr/testify/suite"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/mgo.v2/dbtest"
)

// MongoSuite is a testify Suite that adds functions to setup and teardown test Mongo DBs, as well as other provides
// functions for other common Mongo-related tasks such as inserting a fixture into the database.
type MongoSuite struct {
	suite.Suite
	dbServer      *dbtest.DBServer
	dbServerPath  string
	dbServerMutex sync.Mutex
	db            *mgo.Database
	dbMutex       sync.Mutex
}

// DBServer returns a pointer to the suite's test Mongo DBServer, initializing a new one if necessary.  Subsequent calls
// will return the same DBServer until TearDownDBServer is called.  In most cases, calls to DBServer() are completely
// unnecessary, as calls to DB() will automatically instantiate the DBServer if needed.  If a test suite calls
// DBServer(), then TearDownDBServer() MUST be called in the TearDownSuite function (or TearDownTest, if appropriate).
func (m *MongoSuite) DBServer() *dbtest.DBServer {
	m.dbServerMutex.Lock()
	defer m.dbServerMutex.Unlock()

	if m.dbServer == nil {
		var err error
		m.dbServer = &dbtest.DBServer{}
		m.dbServerPath, err = ioutil.TempDir("", "mongotestdb")
		m.Require().NoError(err)
		m.dbServer.SetPath(m.dbServerPath)
	}
	return m.dbServer
}

// TearDownDBServer cleans up the suite's test Mongo DBServer, including closing any open sessions, wiping the data,
// stopping the test DBServer, and deleting the temporary directory where the DBServer stored files.  This function
// would typically be called in the test suite's TearDownSuite function.
func (m *MongoSuite) TearDownDBServer() {
	m.TearDownDB()

	m.dbServerMutex.Lock()
	defer m.dbServerMutex.Unlock()

	if m.dbServer != nil {
		m.dbServer.Wipe()
		m.dbServer.Stop()
		m.dbServer = nil
	}

	if m.dbServerPath != "" {
		if err := os.RemoveAll(m.dbServerPath); err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: Error cleaning up temp directory: %s", err.Error())
		}
	}
}

// DB returns a pointer to the suite's test Database, named "ie-test", initializing a new Database, Session, and/or
// DBServer if necessary.  Subsequent calls will return the same Database until TearDownDB() is called.  If a test suite
// calls DB(), then TearDownDBServer() MUST be called in the TearDownSuite function (or TearDownTest, if appropriate).
func (m *MongoSuite) DB() *mgo.Database {
	m.dbMutex.Lock()
	defer m.dbMutex.Unlock()

	if m.db == nil {
		m.db = m.DBServer().Session().DB("ie-test")
	}
	return m.db
}

// TearDownDB closes the suite's test Session and wipes the DBServer (thereby removing the Database).  This function
// would typically be called in the test suite's TearDownTest function.
func (m *MongoSuite) TearDownDB() {
	m.dbMutex.Lock()
	defer m.dbMutex.Unlock()

	if m.db != nil {
		m.db.Session.Close()
		m.db = nil
		// suite.dbServer should never be nil at this point, but better safe than sorry
		m.dbServerMutex.Lock()
		if m.dbServer != nil {
			m.dbServer.Wipe()
		}
		m.dbServerMutex.Unlock()
	}
}

// InsertFixture inserts a test fixture into the Mongo database.  If the fixture does not have an _id set, InsertFixture
// will attempt to find the _id field using reflection and will set it to a new BSON ID before inserting.
func (m *MongoSuite) InsertFixture(collection string, pathToFixture string, doc interface{}) {
	require := m.Require()

	// Read the fixture file and unmarshal it to the doc
	data, err := ioutil.ReadFile(pathToFixture)
	require.NoError(err)
	err = json.Unmarshal(data, doc)
	require.NoError(err)

	// If it's a slice, store each element, otherwise just store the thing
	value := reflect.ValueOf(doc).Elem()
	if value.Kind() == reflect.Slice {
		for i := 0; i < value.Len(); i++ {
			m.insertValue(collection, value.Index(i))
		}
	} else {
		m.insertValue(collection, value)
	}
}

// PrintDBServerInfo logs out the connection info for the current test database.
func (m *MongoSuite) PrintDBServerInfo() {
	m.dbServerMutex.Lock()
	defer m.dbServerMutex.Unlock()

	dbs := m.dbServer
	if dbs == nil {
		log.Println("Test database server is not running.")
		return
	}

	m.dbMutex.Lock()
	defer m.dbMutex.Unlock()

	var session *mgo.Session
	if m.db != nil {
		session = m.db.Session
	} else {
		session = dbs.Session()
		defer session.Close()
	}

	log.Println("Test Mongo Database Info {")
	for _, addr := range session.LiveServers() {
		log.Println("  Address:       ", addr)
	}
	if dbNames, err := session.DatabaseNames(); err == nil {
		log.Println("  Databases:     [ ", strings.Join(dbNames, ", "), " ]")
	}
	if info, err := session.BuildInfo(); err == nil {
		log.Println("  Version:       ", info.Version)
		log.Println("  GitVersion:    ", info.GitVersion)
		log.Println("  OpenSSLVersion:", info.OpenSSLVersion)
		log.Println("  Bits:          ", info.Bits)
		log.Println("  MaxObjectSize: ", info.MaxObjectSize)
		log.Println("  Debug:         ", info.Debug)
	}
	log.Println("}")

}

func (m *MongoSuite) insertValue(collection string, value reflect.Value) {
	// Find the ID field and set it, if necessary
	if field, err := findBSONIDField(value); err == nil && field.CanSet() && field.String() == "" {
		field.SetString(bson.NewObjectId().Hex())
	}

	// Store it
	err := m.DB().C(collection).Insert(value.Interface())
	m.Require().NoError(err)
}

func findBSONIDField(value reflect.Value) (reflect.Value, error) {
	for i := 0; i < value.NumField(); i++ {
		valueField := value.Field(i)
		typeField := value.Type().Field(i)
		bsonTag := typeField.Tag.Get("bson")
		if strings.HasPrefix(bsonTag, ",inline") {
			// It's an inline struct, so check it for an _id
			if idField, err := findBSONIDField(valueField); err == nil {
				return idField, nil
			}
		} else if strings.HasPrefix(bsonTag, "_id") {
			return valueField, nil
		} else if typeField.Name == "_id" && (bsonTag == "" || strings.HasPrefix(bsonTag, ",")) {
			return valueField, nil
		}
	}
	return reflect.Value{}, errors.New("No _id field found")
}
