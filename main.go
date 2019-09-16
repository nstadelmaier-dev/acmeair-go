package main

import (
	"dataaccess/connect"
	"loader"
	"routes"

	"encoding/json"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/DeanThompson/ginpprof"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

func main() {
	settings_file := "./settings.json"

	if len(os.Args) > 1 {
		settings_file = os.Args[1]
	}

	file, err := os.Open(settings_file)
	if err != nil {
		panic(err.Error())
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	var settings map[string]interface{}

	err = decoder.Decode(&settings)
	if err != nil {
		panic(err.Error())
	}

	host := "localhost"
	if settings["host"] != nil {
		host = settings["host"].(string)
	}

	port := 8080
	if settings["port"] != nil {
		port = int(settings["port"].(float64))
	}

	contextRoot := "/rest/api"
	if settings["contextRoot"] != nil {
		contextRoot = settings["contextRoot"].(string)
	}

	if settings["useFlightDataRelatedCaching"] != nil {
		routes.UseFlightDataRelatedCaching = settings["useFlightDataRelatedCaching"].(bool)
	}

	//router := gin.Default()
	router := gin.New()
	ginpprof.Wrapper(router)
	p := ginprometheus.NewPrometheus("gin")
	p.Use(router)
	
	db := connect.Connect("mongo", settings)

	api := router.Group(contextRoot)
	api.Use(func(c *gin.Context) {
		s := db.New()
		defer s.Close()

		c.Set("dataaccess", s)
		c.Next()
	})

	api.POST("/login", routes.Login)
	api.GET("/login/logout", routes.Logout)
	api.POST("/flights/queryflights", routes.QueryFlights)
	api.POST("/bookings/bookflights", routes.BookFlights)
	api.POST("/bookings/cancelbooking", routes.CancelBooking)
	api.GET("/bookings/byuser/:user", routes.BookingByUser)
	api.GET("/customer/byid/:user", routes.GetCustomerById)
	api.POST("/customer/byid/:user", routes.PutCustomerById)
	api.GET("/config/runtime", routes.GetRuntimeInfo)
	api.GET("/config/dataServices", routes.GetDataServiceInfo)
	api.GET("/config/activeDataService", routes.GetActiveDataServiceInfo)
	api.GET("/config/countBookings", routes.CountBookings)
	api.GET("/config/countCustomers", routes.CountCustomers)
	api.GET("/config/countSessions", routes.CountSessions)
	api.GET("/config/countFlights", routes.CountFlights)
	api.GET("/config/countFlightSegments", routes.CountFlightSegments)
	api.GET("/config/countAirports", routes.CountAirports)
	api.GET("/loader/load", loader.Load)
	api.GET("/loader/query", loader.GetNumConfiguredCustomers)

	router.Static("/css", "./public/css")
	router.Static("/images", "./public/images")
	router.Static("/js", "./public/js")

	router.StaticFile("/", "./public/index.html")
	router.StaticFile("/checkin.html", "./public/checkin.html")
	router.StaticFile("/config.html", "./public/config.html")
	router.StaticFile("/customerprofile.html", "./public/customerprofile.html")
	router.StaticFile("/flights.html", "./public/flights.html")
	router.StaticFile("/index.html", "./public/index.html")
	router.StaticFile("/loader.html", "./public/loader.html")
	router.StaticFile("/favicon.ico", "./public/favicon.ico")

	hostport := fmt.Sprintf("%s:%d", host, port)
	router.Run(hostport)
}
