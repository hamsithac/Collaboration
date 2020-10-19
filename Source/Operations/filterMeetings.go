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