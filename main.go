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
	fmt.Fprintf(w, "Welcome to the HomePage!")
}

func handleRequests() {
	http.HandleFunc("/me", getMeetingsByPartcipantsEmail) //***************
	http.HandleFunc("/meetings", createMeeting)
	http.HandleFunc("/meetings/", getMeetingWithId)
	http.HandleFunc("/", homePage)
	http.HandleFunc("/time",getMeetingsByTimeRange)   //*************
	log.Fatal(http.ListenAndServe(":10000", nil))
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

func getMeetingsByTimeRange(w http.ResponseWriter, r *http.Request) {
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

	startTimes, startTimeErr := r.URL.Query()["start"]
	endTimes, endTimeError := r.URL.Query()["end"]

	if !startTimeErr || len(startTimes[0]) < 1 {
		log.Println("Url Param 'startTimes' is missing")
		return
	}
	if !endTimeError || len(endTimes[0]) < 1 {
		log.Println("Url Param 'endTimes' is missing")
		return
	}

	endTime := endTimes[0]
	startTime := startTimes[0]

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


func getMeetingsByPartcipantsEmail(w http.ResponseWriter, r *http.Request) {
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

	fmt.Println("Endpoint Hit: Participants")
	json.NewEncoder(w).Encode(Participants)

	EMails, ok := r.URL.Query()["participant"]

	if !ok || len(EMails[0]) < 1 {
		log.Println("Url Param 'participant' is missing")
		return
	}
	EMail := EMails[0]
	EmailToFind := string(EMail)
	filtCursor, err := meetingsCollection.Find(ctx, bson.M{"Participants": bson.M{"$elemMatch": bson.M{"Email": EmailToFind }}})
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

func createMeeting(w http.ResponseWriter, r *http.Request) {
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

	var meeting Meeting
	parseError := json.NewDecoder(r.Body).Decode(&meeting)
	if parseError != nil {
		http.Error(w, parseError.Error(), http.StatusBadRequest)
		return
	}

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
