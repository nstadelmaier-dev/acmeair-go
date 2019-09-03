
package loader

import (
	"dataaccess"
	"routes"
	"github.com/gin-gonic/gin"
	//"gopkg.in/mgo.v2/bson"


	"net/http"
	"encoding/csv"
	"io"
	"os"
	"fmt"
	"strconv"
	"time"

	"github.com/satori/go.uuid"
)


func createCustomer(id string) routes.Customer {
	customer := routes.Customer {
		Id : id,
		Password : "password",
		Status : "GOLD",
		TotalMiles : 1000000,
		MilesYtd : 1000,
		PhoneNumber : "919-123-4567",
		PhoneNumberType : "BUSINESS",
		
		Address : routes.Address{
			StreetAddress1 : "123 Main St.",
			City : "Anytown",
			StateProvince : "NC",
			Country : "USA",
			PostalCode : "27617",
		},
	}

	return customer
}


func getDepartureTimeDaysFromDate(baseTime time.Time, days int) time.Time {
	return baseTime.Add(time.Duration(24 * days) * time.Hour)
}

func getArrivalTime(departureTime time.Time, mileage int) time.Time {
	averageSpeed := 600.0 // 600 miles/hours
	seconds := 60 * 60 * float64(mileage) / averageSpeed
	return departureTime.Add(time.Duration(seconds) * time.Second)
}

func Load(c *gin.Context) {
	db := c.MustGet("dataaccess").(dataaccess.DataAccess)

	n, err := db.Count(db.GetDBNames().CustomerName)

	if n > 0 {
		c.String(http.StatusOK, "Already loaded")
		return
	}

	for i := 0; i <  MAX_CUSTOMERS; i++ {
		uid := fmt.Sprintf("uid%d@email.com", i)
		customer := createCustomer(uid)
		db.InsertOne(db.GetDBNames().CustomerName, customer)
	}


	csvFile, err := os.Open("./src/loader/mileage.csv")
	if err != nil {
		panic(err.Error())
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)

	airportNames, err := reader.Read()
	if err != nil {
		panic(err.Error())
	}

	airportCodes, err := reader.Read()
	if err != nil {
		panic(err.Error())
	}

	for i := 0; i < len(airportNames); i++ {
		mapping := routes.AirportCodeMapping{
			Id :          airportCodes[i],
			AirportName : airportNames[i],
		}
			
		db.InsertOne(db.GetDBNames().AirportCodeMappingName, mapping)
	}

	reader.FieldsPerRecord = 16

	now := time.Now()
	nowAtMidnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	flightSegmentId := 0
	// actual mileages start on the third (2) row
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err.Error())
		}

		// format of the row is "long airport name name" (0), "airport code" (1), mileage to first airport in rows 0/1 (2), mileage to second airport in rows 0/1 (3), ... mileage to last airport in rows 0/1 (length)

		fromAirportCode := record[1]
		
		for j := 0; j < len(record) - 2; j++ {
			toAirportCode := airportCodes[j]
			mileage := record[j + 2];
			if (mileage != "NA") {
				miles, err := strconv.Atoi(mileage)
				if err != nil {
					panic(err.Error())
				}
				
				flightSegment := routes.FlightSegment{
					Id : fmt.Sprintf("AA%d", flightSegmentId),
					OriginPort : fromAirportCode,
					DestPort : toAirportCode,
					Miles : miles,
				}
				
				db.InsertOne(db.GetDBNames().FlightSegmentName, flightSegment)

				for k := 0; k < MAX_DAYS_TO_SCHEDULE_FLIGHTS; k++ {
					for l := 0; l < MAX_FLIGHTS_PER_DAY; l++ {
						departureTime := getDepartureTimeDaysFromDate(nowAtMidnight, k)
						flight := routes.Flight{
							Id : uuid.Must(uuid.NewV4()).String(),
							FlightSegmentId : flightSegment.Id,
							ScheduledDepartureTime : departureTime,
							ScheduledArrivalTime : getArrivalTime(departureTime, miles),
							FirstClassBaseCost : 500,
							EconomyClassBaseCost : 200,
							NumFirstClassSeats : 10,
							NumEconomyClassSeats : 200,
							AirplaneTypeId : "B747",
						}
						db.InsertOne(db.GetDBNames().FlightName, flight)
					}
				}
				
			}
			flightSegmentId += 1
		}
	}


	c.String(http.StatusOK, "Database Finished Loading")
}


func GetNumConfiguredCustomers(c *gin.Context) {
	c.String(http.StatusOK, "%d", MAX_CUSTOMERS)
}
