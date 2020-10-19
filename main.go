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
	json.NewEncoder(w).Encode(Participants)
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

func main() {
	handleRequests()
}
