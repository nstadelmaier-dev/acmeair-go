# Acme Air in Go

An implementation of the Acme Air sample application for Go.  This implementation is still under development, and only supports MongoDB for data store.  

## Content

### Runtime Environment

[Go](https://golang.org/)

### Datastore Choices

MongoDB is supported.

* [MongoDB](http://www.mongodb.org) 

### Application Mode

Micro-Service mode is not supported.

## How to get started

Assume MongoDB started on 127.0.0.1:27017

### Resolve module dependencies

	mkdir ~/go
	export GOPATH=~/go
	go get github.com/gin-gonic/gin
	go get github.com/globalsign/mgo
	go get github.com/streamrail/concurrent-map
	go get github.com/DeanThompson/ginpprof
	go get github.com/satori/go.uuid

### Build Acmeair

	GOPATH+=":$PWD"
	go build -o acmeair

### Run Acmeair in Monolithic on Local

	./acmeair
		
### Access Application 

	http://localhost:8080/
	
	Load Database 
		preload 10k customers uid[0..9999]@email.com:password, 5 days' flights.  Defined in src/loader/settings.go
	Login
	Flights
		such as Singapore to HongKong or Paris to Moscow 
	Checkin
		cancel some booked flights
	Account
		update account info
	Logout	
	
	
## More on Configurations

### Configuration for Runtime

Default values are defined [here](settings.json)

Name | Default | Meaning
--- |:---:| ---
mongoHost | 127.0.0.1 | MongoDB host ip
mongoPort | 27017 | MongoDB port
mongoConnectionPoolSize | 10 | MongoDB connection pool size


### Configuration for Preload

Default values are defined [here](src/loader/settings.go)

Name | Default | Meaning
--- |:---:| ---
MAX_CUSTOMERS | 10000 |  number of customers
MAX_DAYS_TO_SCHEDULE_FLIGHTS | 5 | max number of days to schedule flights
MAX_FLIGHTS_PER_DAY | 1 | max flights per day
