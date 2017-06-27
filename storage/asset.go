package storage

import (
	"fmt"
	"strings"

	"gopkg.in/mgo.v2/bson"
	"twreporter.org/go-api/models"

	log "github.com/Sirupsen/logrus"
)

func (m *MongoStorage) _StringToPscalCase(str string) (pscalCase string) {
	isToUpper := true
	for _, runeValue := range str {
		if isToUpper {
			pscalCase += strings.ToUpper(string(runeValue))
			isToUpper = false
		} else {
			if runeValue == '_' {
				isToUpper = true
			} else {
				pscalCase += string(runeValue)
			}
		}
	}
	return
}

// GetEmbeddedAsset ...
func (m *MongoStorage) GetEmbeddedAsset(entity models.NewsEntity, embedded []string) {
	if embedded != nil {
		for _, ele := range embedded {
			switch ele {
			case "writters", "photographers", "designers", "engineers":
				assetName := m._StringToPscalCase(ele)
				if ids := entity.GetEmbeddedAsset(assetName + "Origin"); ids != nil {
					if len(ids) > 0 {
						authors, err := m.GetAuthors(ids)
						if err == nil {
							entity.SetEmbeddedAsset(assetName, authors)
						}
					}
				}
				break
			case "hero_image", "leading_image", "leading_image_portrait", "og_image":
				assetName := m._StringToPscalCase(ele)
				if ids := entity.GetEmbeddedAsset(assetName + "Origin"); ids != nil {
					if len(ids) > 0 {
						img, err := m.GetImage(ids[0])
						if err == nil {
							entity.SetEmbeddedAsset(assetName, &img)
						}
					}
				}
				break
			case "leading_video":
				if ids := entity.GetEmbeddedAsset("LeadingVideoOrigin"); ids != nil {
					if len(ids) > 0 {
						video, err := m.GetVideo(ids[0])
						if err == nil {
							entity.SetEmbeddedAsset("LeadingVideo", &video)
						}
					}
				}
				break
			case "categories":
				if ids := entity.GetEmbeddedAsset("CategoriesOrigin"); ids != nil {
					categories, _ := m.GetCategories(ids)
					_categories := make([]models.Category, len(categories))
					for i, v := range categories {
						_categories[i] = v
					}
					entity.SetEmbeddedAsset("Categories", _categories)
				}
				break
			case "tags":
				if ids := entity.GetEmbeddedAsset("TagsOrigin"); ids != nil {
					tags, _ := m.GetTags(ids)
					_tags := make([]models.Tag, len(tags))
					for i, v := range tags {
						_tags[i] = v
					}
					entity.SetEmbeddedAsset("Tags", _tags)
				}
				break
			case "relateds_meta":
				if ids := entity.GetEmbeddedAsset("RelatedsOrigin"); ids != nil {
					relateds, err := m.GetRelatedsMeta(ids)
					if err == nil {
						entity.SetEmbeddedAsset("Relateds", relateds)
					}
				}
				break
			case "topic_meta":
				if ids := entity.GetEmbeddedAsset("TopicOrigin"); ids != nil {
					if len(ids) > 0 {
						t, err := m.GetTopicMeta(ids[0])
						if err == nil {
							entity.SetEmbeddedAsset("Topic", &t)
						}
					}
				}
				break
			default:
				log.Info(fmt.Sprintf("Embedded element (%v) is not supported: ", ele))
			}
		}
	}
	return
}

func (m *MongoStorage) GetTopicMeta(id bson.ObjectId) (models.Topic, error) {

	query := bson.M{
		"_id": id,
	}

	topics, err := m.GetTopics(query, 0, 0, "-publishedDate", []string{"leading_image", "og_image"})

	if err != nil {
		return models.Topic{}, err
	}

	return topics[0], nil
}

// GetRelatedsMeta ...
func (m *MongoStorage) GetRelatedsMeta(ids []bson.ObjectId) ([]models.Post, error) {

	query := bson.M{
		"_id": bson.M{
			"$in": ids,
		},
	}

	posts, err := m.GetMetaOfPosts(query, 0, 0, "-publishedDate", []string{"hero_image", "og_image"})

	if err != nil {
		return nil, err
	}

	return posts, nil
}

// GetCategories ...
func (m *MongoStorage) GetCategories(ids []bson.ObjectId) ([]models.Category, error) {
	var cats []models.Category

	if ids == nil {
		return cats, nil
	}

	query := bson.M{
		"_id": bson.M{
			"$in": ids,
		},
	}

	err := m.GetDocuments(query, 0, 0, "_id", "postcategories", &cats)

	if err != nil {
		return nil, m.NewStorageError(err, "GetCategories", "storage.posts.get_categories")
	}

	return cats, nil
}

// GetTags ...
func (m *MongoStorage) GetTags(ids []bson.ObjectId) ([]models.Tag, error) {
	var tags []models.Tag

	if ids == nil {
		return tags, nil
	}

	query := bson.M{
		"_id": bson.M{
			"$in": ids,
		},
	}

	err := m.GetDocuments(query, 0, 0, "_id", "tags", &tags)

	if err != nil {
		return nil, m.NewStorageError(err, "GetCategories", "storage.posts.get_tags")
	}

	return tags, nil
}

// GetVideo ...
func (m *MongoStorage) GetVideo(id bson.ObjectId) (models.Video, error) {
	var mgoVideo models.MongoVideo

	err := m.GetDocument(id, "videos", &mgoVideo)

	if err != nil {
		return models.Video{}, err
	}

	return mgoVideo.ToVideo(), nil
}

// GetImage ...
func (m *MongoStorage) GetImage(id bson.ObjectId) (models.Image, error) {
	var mgoImg models.MongoImage

	err := m.GetDocument(id, "images", &mgoImg)

	if err != nil {
		return models.Image{}, err
	}

	return mgoImg.ToImage(), nil
}

func (m *MongoStorage) GetAuthors(ids []bson.ObjectId) ([]models.Author, error) {
	var authors []models.Author

	if ids == nil {
		return authors, nil
	}

	query := bson.M{
		"_id": bson.M{
			"$in": ids,
		},
	}

	err := m.GetDocuments(query, 0, 0, "_id", "contacts", &authors)

	if err != nil {
		return authors, err
	}

	return authors, nil
}
