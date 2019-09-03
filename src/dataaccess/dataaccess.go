
package dataaccess

import (
	"github.com/globalsign/mgo/bson"
)
	
type DataAccess interface {
	New() DataAccess
	Close()
	GetDBType() string
	GetDBNames() *DBNames
	InitializeDatabaseConnections(settings map[string]interface{})
	InsertOne(dbName string, doc interface{}) error
	FindOne(dbName string, key interface{}, doc interface{}) error
	FindBy(dbName string, query bson.M, docs interface{}) error
	Remove(dbName string, key interface{}) error
	RemoveBy(dbName string, query bson.M) error
	Update(dbName string, key interface{}, doc interface{}) error
	Count(dbName string) (int, error)
}

type DBNames struct {
	CustomerName string
	FlightName string
	FlightSegmentName string
	BookingName string
	CustomerSessionName string
	AirportCodeMappingName string
}
