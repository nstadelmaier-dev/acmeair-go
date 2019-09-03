package mongo

import (
	"fmt"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"dataaccess"
)

type Mongo struct {
	Session *mgo.Session
	PooledSessions chan *mgo.Session
	Parent *Mongo
}

var mongoDBNames dataaccess.DBNames = dataaccess.DBNames {
	CustomerName           : "customer",
	FlightName             : "flight",
	FlightSegmentName      : "flightSegment",
	BookingName            : "booking",
	CustomerSessionName    : "customerSession",
	AirportCodeMappingName : "airportCodeMapping",
}


func (db *Mongo) New() dataaccess.DataAccess {
	return &Mongo{ Session : <- db.PooledSessions, Parent: db }
	//return &Mongo{ Session : db.Session.Clone() }
}

func (db *Mongo) Close() {
	if db.Parent != nil {
		go func() {
			db.Parent.PooledSessions <- db.Session
			defer func() {
				if r := recover(); r != nil {
					db.Session.Close()
				}
			}()
		}()
	} else {
		close(db.PooledSessions)
		for s := range db.PooledSessions {
			s.Close()
		}
		db.Session.Close()
	}
}


func (db *Mongo) GetDBType() string {
	return "mongo"
}


func (db *Mongo) GetDBNames() *dataaccess.DBNames {
	return &mongoDBNames
}


func (db *Mongo) InitializeDatabaseConnections(settings map[string]interface{}) {
	mongoDatabaseName := "acmeair"
	if settings["mongoDatabaseName"] != nil {
		mongoDatabaseName = settings["mongoDatabaseName"].(string)
	}

        mongoUsername := "user"
        if settings["mongoUsername"] != nil {
                mongoUsername = settings["mongoUsername"].(string)
        }

        mongoPassword := "pass"
        if settings["mongoPassword"] != nil {
                mongoPassword = settings["mongoPassword"].(string)
        }

        mongoHost := "127.0.0.1"
        if settings["mongoHost"] != nil {
                mongoHost = settings["mongoHost"].(string)
        }

	mongoPort := 27017
	if settings["mongoPort"] != nil {
		mongoPort = int(settings["mongoPort"].(float64))
	}

	mongoConnectionPoolSize := 10
	if settings["mongoConnectionPoolSize"] != nil {
		mongoConnectionPoolSize = int(settings["mongoConnectionPoolSize"].(float64))
	}

	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s", mongoUsername, mongoPassword, mongoHost, mongoPort, mongoDatabaseName)

	s, err := mgo.Dial(uri)
	if err != nil {
		fmt.Printf("Can't connect to mongo, go error %v\n", err)
		panic(err.Error())
	}

	s.SetMode(mgo.Monotonic, true)
	s.SetSafe(&mgo.Safe{})
	s.SetPoolLimit(mongoConnectionPoolSize)

	db.Session = s
	db.PooledSessions = make(chan *mgo.Session, mongoConnectionPoolSize)
	for i := 0; i < mongoConnectionPoolSize; i++ {
		db.PooledSessions <- s.Copy()
	}
}


func (db *Mongo) InsertOne(dbName string, doc interface{}) error {
	err := db.Session.DB("acmeair").C(dbName).Insert(doc)
	return err
}


func (db *Mongo) Update(dbName string, key interface{}, doc interface{}) error {
	err := db.Session.DB("acmeair").C(dbName).Update(bson.M{ "_id" : key }, doc)
	return err
}


func (db *Mongo) FindOne(dbName string, key interface{}, doc interface{}) error {
	err := db.Session.DB("acmeair").C(dbName).FindId(key).All(doc)
	return err
}


func (db *Mongo) Remove(dbName string, key interface{}) error {
	err := db.Session.DB("acmeair").C(dbName).RemoveId(key)
	return err
}


func (db *Mongo) RemoveBy(dbName string, query bson.M) error {
	err := db.Session.DB("acmeair").C(dbName).Remove(query)
	return err
}


func (db *Mongo) FindBy(dbName string, query bson.M, docs interface{}) error {
	err := db.Session.DB("acmeair").C(dbName).Find(query).All(docs)
	return err
}


func (db *Mongo) Count(dbName string) (int, error) {
	n, err := db.Session.DB("acmeair").C(dbName).Count()
	return n, err
}
