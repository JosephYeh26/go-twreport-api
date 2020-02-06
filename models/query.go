package models

import (
	"errors"
	//"encoding/json"
	// log "github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
)

// In order to make the codes more flexible,
// we define the Query interface to generalize the type functions.
// By implementing the UnmarshalQueryString method like MongoQuery type does,
// we can parse the query string into different types by passing different type as parameter into GetQuery function.
// In the future, if we want to add another type to parse the query string,
// we simply add the new type and let it implement UnmarshalQueryString method.

// Query is an interface which defines the UmarshalQueryString method.
type Query interface {
	UnmarshalQueryString(string) error
}

// MongoQueryComparison ...
type MongoQueryComparison struct {
	In []bson.ObjectId `json:"in" bson:"$in,omitempty"`
}

// MongoQuery implements Query interface, which stores the JSON in Query field.
type MongoQuery struct {
	State      string               `bson:"state,omitempty" json:"state"`
	Slug       string               `bson:"slug,omitempty" json:"slug"`
	Style      string               `bson:"style,omitempty" json:"style"`
	IsFeatured bool                 `bson:"isFeatured,omitempty" json:"is_featured"`
	Categories MongoQueryComparison `bson:"categories,omitempty" json:"categories"`
	Tags       MongoQueryComparison `bson:"tags,omitempty" json:"tags"`
	Topics     MongoQueryComparison `bson:"topics,omitempty" json:"topics"`
	IDs        MongoQueryComparison `bson:"_id,omitempty" json:"ids"`
}

func (query MongoQuery) ValidObjectIds(ids []bson.ObjectId) bool {
	for _, id := range ids {
		if !id.Valid() {
			return false
		}
	}
	return true
}

// UnmarshalQueryString is type-specific functions of MongoQuery type
func (query *MongoQuery) UnmarshalQueryString(qs string) error {
	if err := bson.UnmarshalJSON([]byte(qs), &query); err != nil {
		return err
	}

	if !query.ValidObjectIds(query.Categories.In) {
		return errors.New("Category id should be a mongo ObjectId")
	}
	if !query.ValidObjectIds(query.Tags.In) {
		return errors.New("Tag id should be a mongo ObjectId")
	}
	if !query.ValidObjectIds(query.Topics.In) {
		return errors.New("Topic id should be a mongo ObjectId")
	}
	if !query.ValidObjectIds(query.IDs.In) {
		return errors.New("id should be a mongo ObjectId")
	}

	return nil
}

// GetQuery takes an Query interface value this is guaranteed to have an UnmarshlQueryString method.
// Then, use the method of parse the query string.
func GetQuery(qs string, query Query) error {
	return query.UnmarshalQueryString(qs)
}
