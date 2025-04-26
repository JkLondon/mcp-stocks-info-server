package models

import (
	"time"
)

// News представляет собой финансовую новость
type News struct {
	ID          string    `json:"id" bson:"_id"`
	Title       string    `json:"title" bson:"title"`
	Description string    `json:"description" bson:"description"`
	Content     string    `json:"content" bson:"content"`
	URL         string    `json:"url" bson:"url"`
	Source      string    `json:"source" bson:"source"`
	PublishedAt time.Time `json:"published_at" bson:"published_at"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	Tags        []string  `json:"tags" bson:"tags"`
	RelatedTo   []string  `json:"related_to" bson:"related_to"` // Связанные тикеры акций
}
