package main

/// Importing 
import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Creating a template for Participant
type Participant struct {
	Name  string `json:"Name" bson:"Name"`
	Email string `json:"Email" bson:"Email"`
	RSVP  string `json:"RSVP" bson:"RSVP"`
}

// Creating a template for Meeting
type Meeting struct {
	Id                int           `json:"Id" bson:"Id"`
	Title             string        `json:"Title" bson:"Title"`
	Participants      []Participant `json:"Participants" bson:"Participants"`
	StartTime         time.Time     `json:"StartTime" bson:"StartTime"`
	EndTime           time.Time     `json:"EndTime" bson:"EndTime"`
	CreationTimestamp time.Time     `json:"CreationTimestamp" bson:"CreationTimestamp"`
}


// homepage
func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the Meeting Schedule API!")
}

// handling requests
func handleRequests() {
	// Either create a new meeting or search existing meetings based on query parameter.
	http.HandleFunc("/meetings", handleMeetingsPath)

	// get meeting with given ID using path parameter
	http.HandleFunc("/meetings/", getMeetingWithId)

	// homepage  for default route
	http.HandleFunc("/", homePage)
	
	log.Fatal(http.ListenAndServe(":10000", nil))
}


// Either create a new meeting or search existing meetings based on GET and POST .
func handleMeetingsPath (w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		var meeting Meeting
		parseError := json.NewDecoder(r.Body).Decode(&meeting)
		if parseError != nil {
			http.Error(w, parseError.Error(), http.StatusBadRequest)
			return
		}

		if checkMeetingValidity(meeting) {
            // creating a new meeting
		    createMeeting(meeting, w);
		} else {
			fmt.Fprintf(w, "Some paticipants have conflicting meetings")
		}

	
	case "GET":
		startTimeQuery := r.URL.Query()["start"]
		endTimeQuery := r.URL.Query()["end"]

		participantEmailPath := r.URL.Query()["participant"]

		// pagination
		limitQuery := r.URL.Query()["limit"]

		if startTimeQuery != nil && endTimeQuery != nil {
			endTime := endTimeQuery[0]
			startTime := startTimeQuery[0]
			limit, _ := strconv.Atoi(limitQuery[0])

			getMeetingsByTimeRange(startTime, endTime, limit, w)
		} else if participantEmailPath != nil {
			participantEmail := participantEmailPath[0]
			limit, _ := strconv.Atoi(limitQuery[0])

			getMeetingsByPartcipantsEmail(participantEmail, limit, w)
		}

	default:
		w.WriteHeader(http.StatusNotImplemented)
	    w.Write([]byte(http.StatusText(http.StatusNotImplemented)))	
	}
}


// get all the meetings within a time range
func getMeetingsByTimeRange(startTime string, endTime string, limitQuery int,  w http.ResponseWriter) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	meetingsCollection := client.Database("Collaboration").Collection("Meetings")

	fmt.Println("Endpoint Hit: Time frame")
	options := options.Find()
	options.SetLimit(int64(limitQuery))

	filtCursor, err := meetingsCollection.Find(ctx, bson.M{"StartTime": bson.M{"$gte": startTime}, "EndTime": bson.M{"$lte": endTime}}, options)

	if err != nil {
		log.Fatal(err)
	}
	var meetingFiltered []bson.M
	if err = filtCursor.All(ctx, &meetingFiltered); err != nil {
		log.Fatal(err)
	}

	fmt.Println(meetingFiltered)

	res, _ := json.Marshal(meetingFiltered)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

// get all the meetings with a given Email Id
func getMeetingsByPartcipantsEmail(participantEmail string, limitQuery int, w http.ResponseWriter) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	meetingsCollection := client.Database("Collaboration").Collection("Meetings")

	fmt.Println(participantEmail)
	options := options.Find()
	options.SetLimit(int64(limitQuery))
	filtCursor, err := meetingsCollection.Find(ctx, bson.M{"Participants": bson.M{"$elemMatch": bson.M{"Email": participantEmail }}}, options)
	if err != nil {
		log.Fatal(err)
	}
	var meetingFiltered []bson.M
	if err = filtCursor.All(ctx, &meetingFiltered); err != nil {
		log.Fatal(err)
	}

	res, _ := json.Marshal(meetingFiltered)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

// to check whether a meeting is valid or not 
func checkMeetingValidity(meeting Meeting) (isValid bool){
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	meetingsCollection := client.Database("Collaboration").Collection("Meetings")

	participantsList := meeting.Participants
	newMeetingstartTime := meeting.StartTime
	//newMeetingendTime := meeting.EndTime

	isMeetingValid := true
	for _, participant := range participantsList {
		// get meetings of participant where the start time of current meeting is later than 

		filtCursor, err := meetingsCollection.Find(ctx, 
			bson.M{
				"StartTime": bson.M{"$lte": newMeetingstartTime},
				"EndTime": bson.M{"$gte": newMeetingstartTime},
				"Participants": bson.M{"$elemMatch": bson.M{"Email": participant.Email}},
			})

	    if err != nil {
		   log.Fatal(err)
	    }
	    var meetingFiltered []bson.M
	    if err = filtCursor.All(ctx, &meetingFiltered); err != nil {
		   log.Fatal(err)
	    }


		if len(meetingFiltered) > 0 {
		   // if overlapping meetings are present
		   isMeetingValid = false
		}
	}

	return isMeetingValid
}

// creating a new meeeting
func createMeeting(meeting Meeting, w http.ResponseWriter) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	meeting.CreationTimestamp = time.Now()
	collection := client.Database("Collaboration").Collection("Meetings")
	insertResult, err := collection.InsertOne(context.TODO(), meeting)

	if err != nil {
		log.Fatal(err)
	}

	res, _ := json.Marshal(insertResult)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(res)
}

// get all the meetings with a given ID
func getMeetingWithId(w http.ResponseWriter, r *http.Request) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	// pagination

	limitQuery := r.URL.Query()["limit"]
	limit, _ := strconv.Atoi(limitQuery[0])
	options := options.Find()
	options.SetLimit(int64(limit))

	meetingsCollection := client.Database("Collaboration").Collection("Meetings")

	meetingId, err := strconv.Atoi(getFirstParam(r.URL.Path))
	filtCursor, err := meetingsCollection.Find(ctx, bson.M{"Id": meetingId}, options)
	if err != nil {
		log.Fatal(err)
	}
	var meetingFiltered []bson.M
	if err = filtCursor.All(ctx, &meetingFiltered); err != nil {
		log.Fatal(err)
	}

	res, _ := json.Marshal(meetingFiltered)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

func getFirstParam(path string) (ps string) {
	// ignore first '/' and when it hits the second '/'
	// get whatever is after it as a parameter
	for i := 1; i < len(path); i++ {
		if path[i] == '/' {
			ps = path[i+1:]
		}
	}
	return
}

// main func
func main() {
	handleRequests()
}
