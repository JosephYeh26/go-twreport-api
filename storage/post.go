package storage

import (
	"twreporter.org/go-api/models"
)

// _GetPosts finds the posts according to query string and also get the embedded assets
func (m *MongoStorage) _GetPosts(qs interface{}, limit int, offset int, sort string, embedded []string) ([]models.Post, error) {
	var posts []models.Post
	err := m.GetDocuments(qs, limit, offset, sort, "posts", &posts)

	if err != nil {
		return posts, err
	}

	for index := range posts {
		m.GetEmbeddedAsset(&posts[index], embedded)
	}

	return posts, nil
}

// GetMetaOfPosts is a type-specific functions implementing the method defined in the NewsStorage.
// It finds the posts according to query string and only return the metadata of posts.
func (m *MongoStorage) GetMetaOfPosts(qs interface{}, limit int, offset int, sort string, embedded []string) ([]models.Post, error) {
	if embedded == nil {
		embedded = []string{"hero_image", "categories", "tags", "topic_meta", "og_image"}
	}

	posts, err := m._GetPosts(qs, limit, offset, sort, embedded)

	if err != nil {
		return posts, err
	}

	// remove content because of size
	for index := range posts {
		posts[index].Content = nil
	}

	return posts, nil
}

// GetFullPosts is a type-specific functions implementing the method defined in the NewsStorage.
// It finds the posts according to query string.
func (m *MongoStorage) GetFullPosts(qs interface{}, limit int, offset int, sort string, embedded []string) ([]models.Post, error) {
	if embedded == nil {
		embedded = []string{"hero_image", "leading_video", "categories", "tags", "topic_meta", "og_image", "writters", "photographers", "designers", "engineers", "relateds_meta"}
	}

	return m._GetPosts(qs, limit, offset, sort, embedded)
}
