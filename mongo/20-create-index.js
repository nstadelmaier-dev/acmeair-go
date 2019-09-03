db.auth("mongo", "password");
db.booking.createIndex({customerId:1});
db.flightSegment.createIndex({originPort:1, destPort:1});
db.flight.db.flightSegment.createIndex({flightSegmentId:1,scheduledDepartureTime:1});
