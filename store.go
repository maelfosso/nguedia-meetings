package main

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoStore struct {
	db *mongo.Database
}

func NewMongoStore(db *mongo.Database) *MongoStore {
	return &MongoStore{db}
}

func (m *MongoStore) CheckAvailability(name string) (bool, error) {
	collection := m.db.Collection("meetings")

	var meeting Meeting
	if err := collection.
		FindOne(context.Background(), bson.M{"name": name}).
		Decode(&meeting); err != nil {

		if err == mongo.ErrNoDocuments {
			return true, nil
		} else {
			return false, err
		}
	}

	return false, nil
}

func (m *MongoStore) isMeetingExists(meetingID string) (bool, error) {
	collection := m.db.Collection("meetings")

	var meeting Meeting
	if err := collection.
		FindOne(context.Background(), bson.M{"_id": meetingID}).
		Decode(&meeting); err != nil {

		if err == mongo.ErrNoDocuments {
			return false, nil
		} else {
			return false, fmt.Errorf("error when fetching the meeting - %s - %q", meetingID, err)
		}
	}

	return true, nil
}

func (m *MongoStore) CreateMeeting(meeting *Meeting) error {
	collection := m.db.Collection("meetings")

	insertResult, err := collection.InsertOne(context.Background(), meeting)
	if err != nil {
		return fmt.Errorf("error when inserting a new meeting document - %q", err)
	}
	meeting.ID = insertResult.InsertedID.(primitive.ObjectID)

	return nil
}

func (m *MongoStore) AddMember(meetingID string, member *Member) error {
	meetingsColl := m.db.Collection("meetings")
	membersColl := m.db.Collection("members")

	insertResult, err := membersColl.InsertOne(context.Background(), member)
	if err != nil {
		return fmt.Errorf("error when inserting a new member document - %q", err)
	}
	member.ID = insertResult.InsertedID.(primitive.ObjectID)

	_, err = meetingsColl.UpdateByID(
		context.Background(),
		meetingID,
		bson.M{
			"$push": bson.M{
				"members": bson.M{
					"id":   member.ID,
					"name": member.Name,
				},
			},
		},
	)
	if err != nil {
		return fmt.Errorf("error when updating meeting with new member - %q", err)
	}

	return nil
}
