package news

import (
	"reflect"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/guregu/null.v3"
	"twreporter.org/go-api/internal/mongo"
	"twreporter.org/go-api/internal/query"
)

// tagMongo is used to map the query field to the corresponded field in real mongo document
const tagMongo = "mongo"

type mongoQuery struct {
	mongoPagination
	mongoFilter
	mongoSort
}

// BuildQueryStatments build query statements from pagination/filter/sort objects
// TODO(babygoat): this can be further refactor to accept generic storage query interface
func BuildQueryStatements(mq *mongoQuery) []bson.D {
	var stages []bson.D

	stages = append(stages, mq.mongoFilter.BuildStage()...)
	stages = append(stages, mq.mongoPagination.BuildStage()...)
	stages = append(stages, mq.mongoSort.BuildStage()...)

	return stages
}

func NewMongoQuery(q *Query) *mongoQuery {
	return &mongoQuery{
		fromPagination(q.Pagination),
		fromFilter(q.Filter),
		fromSort(q.Sort),
	}
}

type mongoPagination struct {
	Skip  int
	Limit int
}

func fromPagination(p query.Pagination) mongoPagination {
	return mongoPagination{
		Skip:  p.Offset,
		Limit: p.Limit,
	}
}

func (mp mongoPagination) BuildStage() []bson.D {
	var stages []bson.D
	if mp.Skip > 0 {
		stages = append(stages, mongo.BuildDocument(mongo.StageSkip, mp.Skip))
	}
	if mp.Limit > 0 {
		stages = append(stages, mongo.BuildDocument(mongo.StageLimit, mp.Limit))
	}

	return stages
}

type mongoFilter struct {
	Slug       string               `mongo:"slug"`
	State      string               `mongo:"state"`
	Style      string               `mongo:"style"`
	IsFeatured null.Bool            `mongo:"isFeatured"`
	Categories []primitive.ObjectID `mongo:"categories"`
	Tags       []primitive.ObjectID `mongo:"tags"`
	IDs        []primitive.ObjectID `mongo:"_id"`
}

func (mf mongoFilter) BuildStage() []bson.D {
	var match []bson.D
	if elements := mf.buildElements(); len(elements) > 0 {
		match = append(match, mongo.BuildDocument(mongo.StageMatch, elements))
	}
	return match
}

func (mf mongoFilter) buildElements() []bson.E {
	typ := reflect.TypeOf(mf)
	val := reflect.ValueOf(mf)

	var elements []bson.E
	for i := 0; i < typ.NumField(); i++ {
		fieldT := typ.Field(i)
		fieldV := val.Field(i)

		tag := fieldT.Tag.Get(tagMongo)

		switch fieldV.Interface().(type) {
		case string:
			v := fieldV.Interface().(string)
			if v != "" {
				elements = append(elements, mongo.BuildElement(tag, v))
			}
		case null.Bool:
			v := fieldV.Interface().(null.Bool)
			if !v.IsZero() {
				elements = append(elements, mongo.BuildElement(tag, v.Bool))
			}

		case []primitive.ObjectID:
			if v, ok := mongo.BuildArray(fieldV.Interface().([]primitive.ObjectID)); ok {
				elements = append(elements, mongo.BuildElement(tag, mongo.BuildDocument(mongo.OpIn, v)))
			}
		default:
			log.Errorf("Unimplemented type %+v", fieldT.Type)
		}
	}
	return elements
}

func fromFilter(f Filter) mongoFilter {
	return mongoFilter{
		Slug:       f.Slug,
		State:      f.State,
		Style:      f.Style,
		IsFeatured: f.IsFeatured,
		Categories: hexToObjectIDs(f.Categories),
		Tags:       hexToObjectIDs(f.Tags),
		IDs:        hexToObjectIDs(f.IDs),
	}
}

func hexToObjectIDs(hs []string) []primitive.ObjectID {
	var ids []primitive.ObjectID

	for _, v := range hs {
		id, err := primitive.ObjectIDFromHex(v)
		if err != nil {
			// ignore invalid objectID
			continue
		}
		ids = append(ids, id)
	}
	return ids
}

type mongoSort struct {
	PublishedDate query.Order `mongo:"publishedDate"`
	UpdatedAt     query.Order `mongo:"updatedAt"`
}

func (ms mongoSort) BuildStage() []bson.D {
	typ := reflect.TypeOf(ms)
	val := reflect.ValueOf(ms)

	var sortBy []bson.E
	for i := 0; i < typ.NumField(); i++ {
		fieldT := typ.Field(i)
		fieldV := val.Field(i)

		tag := fieldT.Tag.Get(tagMongo)

		switch fieldV.Interface().(type) {
		case query.Order:
			v := fieldV.Interface().(query.Order)
			if v.IsAsc.Bool {
				sortBy = append(sortBy, mongo.BuildElement(tag, mongo.OrderAsc))
			} else {
				sortBy = append(sortBy, mongo.BuildElement(tag, mongo.OrderDesc))
			}
		default:
		}
	}

	return []bson.D{mongo.BuildDocument(mongo.StageSort, sortBy)}
}

func fromSort(s SortBy) mongoSort {
	return mongoSort{
		PublishedDate: s.PublishedDate,
		UpdatedAt:     s.UpdatedAt,
	}
}
