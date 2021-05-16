package main

type MongoStore struct {
	// db: mongo.Database
}

func NewMongoStore() *MongoStore {
	return &MongoStore{}
}

func (m *MongoStore) CheckAvailability(name string) (bool, error) {
	return true, nil
}

func (m *MongoStore) CreateMeeting(meeting *Meeting) error {
	return nil
}

func AddMember(meeting string, member *Member) error {
	return nil
}
