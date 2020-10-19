package main

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

type Participant struct {
	Name  string `json:"Name" bson:"Name"`
	Email string `json:"Email" bson:"Email"`
	RSVP  string `json:"RSVP" bson:"RSVP"`
}

type Meeting struct {
	Id                int           `json:"Id" bson:"Id"`
	Title             string        `json:"Title" bson:"Title"`
	Participants      []Participant `json:"Participants" bson:"Participants"`
	StartTime         time.Time     `json:"StartTime" bson:"StartTime"`
	EndTime           time.Time     `json:"EndTime" bson:"EndTime"`
	CreationTimestamp time.Time     `json:"CreationTimestamp" bson:"CreationTimestamp"`
}

var Participants []Participant

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Println( r.Method)
	fmt.Fprintf(w, "Welcome to the Meeting Schedule API!")
}

func handleRequests() {
	http.HandleFunc("/meetings", handleMeetingsPath)
	http.HandleFunc("/meetings/", getMeetingWithId)
	http.HandleFunc("/", homePage)
	log.Fatal(http.ListenAndServe(":10000", nil))
}

func handleMeetingsPath (w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		var meeting Meeting
		parseError := json.NewDecoder(r.Body).Decode(&meeting)
		if parseError != nil {
			http.Error(w, parseError.Error(), http.StatusBadRequest)
			return
		}
		createMeeting(meeting, w);
	
	case "GET":
		startTimePath := r.URL.Query()["start"]
		endTimePath := r.URL.Query()["end"]

		participantEmailPath := r.URL.Query()["participant"]

		if startTimePath != nil && endTimePath != nil {
			endTime := endTimePath[0]
			startTime := startTimePath[0]
			getMeetingsByTimeRange(startTime, endTime, w)
		} else if participantEmailPath != nil {
			participantEmail := participantEmailPath[0]
			getMeetingsByPartcipantsEmail(participantEmail, w)
		}

	default:
		w.WriteHeader(http.StatusNotImplemented)
	    w.Write([]byte(http.StatusText(http.StatusNotImplemented)))	
	}
}

func connectToMongo() {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func getMeetingsByTimeRange(startTime string, endTime string, w http.ResponseWriter) {
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
	json.NewEncoder(w).Encode(Participants)

	filtCursor, err := meetingsCollection.Find(ctx,
		bson.M{"StartTime": bson.M{"$gte": startTime}, "EndTime": bson.M{"$lte": endTime}})

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
	// w.WriteHeader(http.StatusOK)
	w.Write(res)

}


func getMeetingsByPartcipantsEmail(participantEmail string, w http.ResponseWriter) {
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
	filtCursor, err := meetingsCollection.Find(ctx, bson.M{"Participants": bson.M{"$elemMatch": bson.M{"Email": participantEmail }}})
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

	meetingsCollection := client.Database("Collaboration").Collection("Meetings")

	meetingId, err := strconv.Atoi(getFirstParam(r.URL.Path))
	filtCursor, err := meetingsCollection.Find(ctx, bson.M{"Id": meetingId})
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

func main() {
	handleRequests()
}
