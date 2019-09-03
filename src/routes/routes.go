
package routes

import (
	"dataaccess"
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"

	cmap "github.com/streamrail/concurrent-map"
	"github.com/satori/go.uuid"
)


var flightCache = cmap.New()
var flightSegmentCache = cmap.New()
var UseFlightDataRelatedCaching = false


type Address struct {
	StreetAddress1 string `json:"streetAddress1" bson:"streetAddress1"`
	City string `json:"city" bson:"city"`
	StateProvince string `json:"stateProvince" bson:"stateProvince"`
	Country string `json:"country" bson:"country"`
	PostalCode string `json:"postalCode" bson:"postalCode"`
}


type Customer struct {
	Id         string `json:"_id,omitempty" bson:"_id,omitempty"`
	Password string `json:"password" bson:"password"`
	Status string `json:"status" bson:"status"`
	TotalMiles int `json:"total_miles" bson:"total_miles"`
	MilesYtd int `json:"miles_ytd" bson:"miles_ytd"`
	Address Address `json:"address" bson:"address"`
	PhoneNumber string `json:"phoneNumber" bson:"phoneNumber"`
	PhoneNumberType string `json:"phoneNumberType" bson:"func"`
}


type CustomerSession struct {
	Id string `json:"_id,omitempty" bson:"_id,omitempty"`
	CustomerID string `json:"customerId" bson:"customerId"`
	LastAccessedTime time.Time `json:"lastAccessedTime" bson:"lastAccessedTime"`
	TimeoutTime time.Time `json:"timeoutTime" bson:"timeoutTime"`	
}


type AirportCodeMapping struct {
	Id  string `json:"_id,omitempty"       bson:"_id,omitempty"`
	AirportName string `json:"airportName" bson:"airportName"`
}


type Flight struct {
	Id                     string `json:"_id,omitempty"          bson:"_id,omitempty"`
	FirstClassBaseCost     int           `json:"firstClassBaseCost"     bson:"firstClassBaseCost"`
	EconomyClassBaseCost   int           `json:"economyClassBaseCost"   bson:"economyClassBaseCost"`
	NumFirstClassSeats     int           `json:"numFirstClassSeats"     bson:"numFirstClassSeats"`
	NumEconomyClassSeats   int           `json:"numEconomyClassSeats"   bson:"numEconomyClassSeats"`
	AirplaneTypeId         string        `json:"airplaneTypeId"         bson:"airplaneTypeId"`
	FlightSegmentId        string        `json:"flightSegmentId"        bson:"flightSegmentId"`
	FlightSegment          FlightSegment `json:"flightSegment"          bson:",omitempty"`
	ScheduledDepartureTime time.Time     `json:"scheduledDepartureTime" bson:"scheduledDepartureTime"`
	ScheduledArrivalTime   time.Time     `json:"scheduledArrivalTime"   bson:"scheduledArrivalTime"`
}


type FlightSegment struct { 
	Id         string `json:"_id,omitempty" bson:"_id,omitempty"`
	OriginPort string        `json:"originPort"    bson:"originPort"`
	DestPort   string        `json:"destPort"      bson:"destPort"`
	Miles      int           `json:"miles"         bson:"miles"`
}


type TripFlight struct {
	NumPages int `json:"numPages"`
	FlightsOptions []Flight `json:"flightsOptions"`
	CurrentPage int `json:"currentPage"`
	HasMoreOptions bool `json:"hasMoreOptions"`
	PageSize int `json:"pageSize"`
}


type Booking struct {
	Id         string `json:"_id,omitempty" bson:"_id,omitempty"`
	CustomerId string `json:"customerId" bson:"customerId"`
	FlightId string `json:"flightId" bson:"flightId"`
	DateOfBooking time.Time `json:"dateOfBooking"   bson:"dateOfBooking"`
}


func getDB(c *gin.Context) dataaccess.DataAccess {
	return c.MustGet("dataaccess").(dataaccess.DataAccess)
}


func validateCustomer(db dataaccess.DataAccess, username, password string) (bool, error) {

	var customer Customer
	var results []Customer

	err := db.FindOne(db.GetDBNames().CustomerName, username, &results)

	if err != nil {
		return false, err
	}
	customer = results[0]

	return (customer.Password == password), nil
}


func createSession(db dataaccess.DataAccess, customerId string) (string, error) {
	var now = time.Now()
	var later = now.Add(1000 * 24 * time.Hour)

	sessionId := uuid.Must(uuid.NewV4()).String()

	document := &CustomerSession{ Id : sessionId, CustomerID : customerId, LastAccessedTime : now, TimeoutTime : later }
			
	err := db.InsertOne(db.GetDBNames().CustomerSessionName, document)
	
	return sessionId, err
}


func Login(c *gin.Context) {
	db := getDB(c)

	var login    = c.PostForm("login")
	var password = c.PostForm("password")

	c.SetCookie("sessionid", "", 0, "", "", false, false)
	
	valid, err := validateCustomer(db, login, password)	
	if err != nil {
		c.Status(http.StatusInternalServerError)

		panic(err.Error())
		return
	} 

	if !valid {
		c.Status(http.StatusForbidden)
		return
	}

	sessionId, err := createSession(db, login)

	if err != nil {
		c.Status(http.StatusInternalServerError)

		panic(err.Error())
		return
	}

	//c.SetCookie("sessionid", sessionId.Hex(), 0, "", "", false, false)
	c.SetCookie("sessionid", sessionId, 0, "", "", false, false)
	c.String(http.StatusOK, "logged in")
}


func Logout(c *gin.Context) {
        db := getDB(c)

        sessionId, _ := c.Cookie("sessionid")
        if len(sessionId) > 0  {
                err := db.Remove(db.GetDBNames().CustomerSessionName, sessionId)
                if err != nil {
                        c.Status(http.StatusInternalServerError)
                        return
                }

                c.SetCookie("sessionid", "", 0, "", "", false, false)
                c.String(http.StatusOK, "logged out")
        } else {
        c.String(http.StatusInternalServerError, "No Session ID")
        }
}

func CheckForValidateSessionCookie(c *gin.Context, db dataaccess.DataAccess) bool {
	id, err := c.Cookie("sessionid")
	if err != nil {
		c.AbortWithStatus(http.StatusForbidden)
		return false
	}

	if id == "" {
		c.AbortWithStatus(http.StatusForbidden)
		return false
	}

	//sessionId := bson.ObjectIdHex(id)
	sessionId := id

	session := &CustomerSession{}
	var results []CustomerSession
	err = db.FindOne(db.GetDBNames().CustomerSessionName, sessionId, &results)
	if err != nil {
		c.AbortWithStatus(http.StatusForbidden)
		return false
	}
	session = &results[0]

	customerId := session.CustomerID

	if customerId == "" {
		c.AbortWithStatus(http.StatusForbidden)
		return false
	}

	c.Set("sessionId", sessionId)
	c.Set("customerId", customerId)

	return true
}


func findFlights(db dataaccess.DataAccess, fromAirport, toAirport string, date time.Time) ([]Flight, error) {
	var flightSegment *FlightSegment

	segmentCacheKey := fromAirport + toAirport
	found := false

	if UseFlightDataRelatedCaching {
		if tmp, ok := flightSegmentCache.Get(segmentCacheKey); ok {
			//fmt.Printf("segment cache hit! %s\n", segmentCacheKey)
			flightSegment = tmp.(*FlightSegment)
			found = true
		}
	}

	if !found {
		//fmt.Printf("segment cache miss! %s\n", segmentCacheKey)
		var segments []FlightSegment
		if err := db.FindBy(db.GetDBNames().FlightSegmentName, bson.M{"originPort": fromAirport, "destPort": toAirport}, &segments); err != nil {
			return nil, err
		}

		if segments != nil && len(segments) > 0 {
			flightSegment = &segments[0]
		}

		if UseFlightDataRelatedCaching {
			flightSegmentCache.Set(segmentCacheKey, flightSegment)
		}
	}

	if flightSegment == nil {
		return make([]Flight, 0), nil
	}

	var flights []Flight
	flightCacheKey := flightSegment.Id + "-" + date.String()

	found = false
	if UseFlightDataRelatedCaching {
		if tmp, ok := flightCache.Get(flightCacheKey); ok {
			//fmt.Printf("flight cache hit! %s\n", flightCacheKey)
			flights = tmp.([]Flight)
			found = true
		}
	}

	if !found {
		//fmt.Printf("flight cache miss! %s\n", flightCacheKey)
		if err := db.FindBy(db.GetDBNames().FlightName, bson.M{"flightSegmentId": flightSegment.Id, "scheduledDepartureTime" : date}, &flights); err != nil {
			return nil, err
		}

		for i, _ := range flights {
			flights[i].FlightSegment = *flightSegment
		}

		if UseFlightDataRelatedCaching {
			flightCache.Set(flightCacheKey, flights)
		}
	}

	if flights == nil {
		return make([]Flight, 0), nil
	}

	return flights, nil
}


func QueryFlights(c *gin.Context) {
	db := getDB(c)

	//fmt.Printf("QueryFlights")

	if ! CheckForValidateSessionCookie(c, db) {
		return
	}

	//fmt.Printf("Session validated")

	//layout := "Mon Jan 02 2006 15:04:05 GMT-0700 (MST)"
	layout := "Mon Jan 02 15:04:05 MST 2006"
	fromAirport := c.PostForm("fromAirport")
	toAirport := c.PostForm("toAirport")
	//fmt.Printf("fromDate=%s\n", c.PostForm("fromDate"))
	fromDateWeb, _ := time.Parse(layout, c.PostForm("fromDate"))
	fromDate := time.Date(fromDateWeb.Year(), fromDateWeb.Month(), fromDateWeb.Day(), 0, 0, 0, 0, time.Local) // convert date to local timezone
	//fromDate := time.Date(2016, 10, 5, 0, 0, 0, 0, time.Local) // convert date to local timezone

	//fmt.Printf("fromAirport=%s, toAirport=%s, fromDate=%s\n", fromAirport, toAirport, fromDate)

	oneWay := c.PostForm("oneWay") == "true";
	returnDateWeb, _ := time.Parse(layout, c.PostForm("returnDate"))
	returnDate := time.Date(returnDateWeb.Year(), returnDateWeb.Month(), returnDateWeb.Day(), 0, 0, 0, 0, time.Local) // convert date to local timezone
	//returnDate := time.Date(2016, 10, 5, 0, 0, 0, 0, time.Local) // convert date to local timezone

	flightsOutbound, err := findFlights(db, fromAirport, toAirport, fromDate)
        //fmt.Printf("flightsOutbound=%s\n", flightsOutbound)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		panic(err.Error())
		return
	}

	if !oneWay {
		flightsReturn, err := findFlights(db, toAirport, fromAirport, returnDate)
                //fmt.Printf("flightsReturn=%s\n", flightsReturn)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			panic(err.Error())
			return
		}
		tripFlights := [...]TripFlight{
			TripFlight{ 
				NumPages : 1,
				FlightsOptions : flightsOutbound,
				CurrentPage : 0,
				HasMoreOptions : false, PageSize: 10,
			},
			TripFlight{
				NumPages : 1,
				FlightsOptions : flightsReturn,
				CurrentPage : 0,
				HasMoreOptions : false,
				PageSize: 10,
			},
		}

		c.JSON(http.StatusOK, gin.H{"tripFlights" : tripFlights, "tripLegs" : 2 })

	} else {
		tripFlights := [...]TripFlight{ 
			TripFlight{ 
				NumPages : 1, 
				FlightsOptions : flightsOutbound,
				CurrentPage : 0,
				HasMoreOptions : false,
				PageSize: 10,
			},
		}

		c.JSON(http.StatusOK, gin.H{"tripFlights" : tripFlights, "tripLegs" : 1 })
	}
}


func BookFlights(c *gin.Context) {
	db := getDB(c)

	if ! CheckForValidateSessionCookie(c, db) {
		return
	}

	userid := c.PostForm("userid")
	toFlight  := c.PostForm("toFlightId")
	retFlight := c.PostForm("retFlightId")

	oneWay := c.PostForm("oneWay") == "true";


	var now = time.Now()
	docId := uuid.Must(uuid.NewV4()).String()

	document := &Booking{ Id : docId, CustomerId : userid, FlightId : toFlight, DateOfBooking : now }
	err := db.InsertOne(db.GetDBNames().BookingName, document)
	if err != nil {
		c.Status(http.StatusInternalServerError)

		panic(err.Error())
		return
	}

	if !oneWay {
		retDocId := uuid.Must(uuid.NewV4()).String()

		retDocument := &Booking{ Id : retDocId, CustomerId : userid, FlightId : retFlight, DateOfBooking : now }
		err = db.InsertOne(db.GetDBNames().BookingName, retDocument)
		if err != nil {
			c.Status(http.StatusInternalServerError)

			panic(err.Error())
			return
		}

		c.JSON(http.StatusOK, gin.H{"oneWay": false, "returnBookingId":retDocId,"departBookingId":docId})
		
	} else {
		c.JSON(http.StatusOK, gin.H{"oneWay": true, "departBookingId": docId})
	}
}


func CancelBooking(c *gin.Context) {
	db := getDB(c)

	if ! CheckForValidateSessionCookie(c, db) {
		return
	}

	number := c.PostForm("number")
	//id := bson.ObjectIdHex(number)
	id := number

	err := db.Remove(db.GetDBNames().BookingName, id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status" : "error"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"status" : "success"})
}


func BookingByUser(c *gin.Context) {
	db := getDB(c)

	if ! CheckForValidateSessionCookie(c, db) {
		return
	}

	userid := c.Param("user")

	var booking []Booking
	err := db.FindBy(db.GetDBNames().BookingName, bson.M{"customerId": userid}, &booking)
	
	if err != nil {
		c.Status(http.StatusInternalServerError)

		panic(err.Error())
		return
	}

	if booking == nil {
		booking = make([]Booking, 0)
	}

	c.JSON(http.StatusOK, booking)
}


func GetCustomerById(c *gin.Context) {
	db := getDB(c)

	if ! CheckForValidateSessionCookie(c, db) {
		return
	}

	userid := c.Param("user")

	var customer Customer
	var results []Customer
	err := db.FindOne(db.GetDBNames().CustomerName, userid, &results)
	
	if err != nil {
		c.Status(http.StatusInternalServerError)

		panic(err.Error())
		return
	}
	customer = results[0]

	c.JSON(http.StatusOK, customer)
}


func PutCustomerById(c *gin.Context) {
	db := getDB(c)

	if ! CheckForValidateSessionCookie(c, db) {
		return
	}

	userid := c.Param("user")
	var customer Customer

	c.BindJSON(&customer)

	err := db.Update(db.GetDBNames().CustomerName, userid, &customer)
	
	if err != nil {
		c.Status(http.StatusInternalServerError)

		panic(err.Error())
		return
	}

	c.JSON(http.StatusOK, customer)
}


func countItems(c *gin.Context, db dataaccess.DataAccess, dbName string) {
	n, err := db.Count(dbName)

	if err != nil {
		c.String(http.StatusOK, "-1")
	} else {
		c.String(http.StatusOK, "%d", n)
	}
}


func CountBookings(c *gin.Context) {
	db := getDB(c)

	countItems(c, db, db.GetDBNames().BookingName)
}


func CountCustomers(c *gin.Context) {
	db := getDB(c)

	countItems(c, db, db.GetDBNames().CustomerName)
}


func CountSessions(c *gin.Context) {
	db := getDB(c)

	countItems(c, db, db.GetDBNames().CustomerSessionName)
}


func CountFlights(c *gin.Context) {
	db := getDB(c)

	countItems(c, db, db.GetDBNames().FlightName)
}


func CountFlightSegments(c *gin.Context) {
	db := getDB(c)

	countItems(c, db, db.GetDBNames().FlightSegmentName)
}


func CountAirports(c *gin.Context) {
	db := getDB(c)

	countItems(c, db, db.GetDBNames().AirportCodeMappingName)
}


func GetRuntimeInfo(c *gin.Context) {
	runtimeInfo := map[string]string{
		"name" : "Runtime",
		"description" : "Go",
	}
	c.JSON(http.StatusOK, [](map[string]string){runtimeInfo})
}


func GetActiveDataServiceInfo(c *gin.Context) {
	db := getDB(c)

	dbType := db.GetDBType()
	c.String(http.StatusOK, dbType)
}


func GetDataServiceInfo(c *gin.Context) {
	/*
	cassandora := map[string]string{
		"name" : "cassandra",
		"description":"Apache Cassandra NoSQL DB",
	}
	
	cloudant := map[string]string{
		"name" : "cloudant",
		"description" : "IBM Distributed DBaaS",
	}
        */
	mongo := map[string]string{
		"name" : "mongo",
		"description" : "MongoDB NoSQL DB",
	}

	c.JSON(http.StatusOK, [](map[string]string){ mongo })
}
