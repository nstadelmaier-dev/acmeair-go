package connect

import (
	"dataaccess"
	"dataaccess/mongo"
)


func Connect(dbtype string, settings map[string]interface{}) dataaccess.DataAccess {
	var db dataaccess.DataAccess

	switch dbtype {
	case "mongo": 
		db = &mongo.Mongo{}
	default : 
		panic ("unknown DB type")
	}

	db.InitializeDatabaseConnections(settings)

	return db
}
