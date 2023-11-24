package controllers

import (
	"html/template"
	"time"

	"github.com/akhilrex/podgrab/db"
)

// Podcast is
type PodcastDTO struct {
	db.Base
	Title string

	Summary template.HTML `gorm:"type:text"`

	Author    string
	AuthorURL string

	Image string

	URL string

	LastEpisode *time.Time

	PodcastItems []PodcastItemDTO

	Tags []*db.Tag `gorm:"many2many:podcast_tags;"`

	AllEpisodesCount int `gorm:"-"`
}

// PodcastItem is
type PodcastItemDTO struct {
	db.Base
	PodcastID string
	Podcast   db.Podcast
	Title     string
	Summary   template.HTML `gorm:"type:text"`

	EpisodeType string

	Duration int

	PubDate time.Time

	FileURL string

	GUID  string
	Image string

	BookmarkDate time.Time

	LocalImage string
}
