package controllers

import (
	"html"
	"html/template"
	"strings"

	"github.com/akhilrex/podgrab/db"
)

func PodcastsToDTO(podcasts *[]db.Podcast) *[]PodcastDTO {
	var podcastsDTO []PodcastDTO

	for _, podcast := range *podcasts {
		var podcastItemsDTO []PodcastItemDTO

		for _, podcastItem := range podcast.PodcastItems {
			var podcastItem = PodcastItemDTO{
				Base: db.Base{
					ID: podcastItem.ID,
				},
				PodcastID:   podcastItem.PodcastID,
				Title:       podcastItem.Title,
				Summary:     formatHTML(podcastItem.Summary),
				EpisodeType: podcastItem.EpisodeType,
				Duration:    podcastItem.Duration,
				PubDate:     podcastItem.PubDate,
				FileURL:     podcastItem.FileURL,
				GUID:        podcastItem.GUID,
				Image:       podcastItem.Image,
				LocalImage:  podcastItem.LocalImage,
			}
			podcastItemsDTO = append(podcastItemsDTO, podcastItem)

		}
		podcastDTO := PodcastDTO{
			Base: db.Base{
				ID: podcast.ID,
			},
			Title:            podcast.Title,
			Summary:          formatHTML(podcast.Summary),
			Author:           podcast.Author,
			AuthorURL:        "",
			Image:            podcast.Image,
			URL:              podcast.URL,
			LastEpisode:      podcast.LastEpisode,
			PodcastItems:     podcastItemsDTO,
			Tags:             podcast.Tags,
			AllEpisodesCount: podcast.AllEpisodesCount,
		}
		podcastsDTO = append(podcastsDTO, podcastDTO)
	}
	return &podcastsDTO
}

func PodcastsItemsToDto(podcastItems []db.PodcastItem) *[]PodcastItemDTO {
	var podcastItemsDTO []PodcastItemDTO
	for _, podcastItem := range podcastItems {
		var podcastItem = PodcastItemDTO{
			Base: db.Base{
				ID: podcastItem.ID,
			},
			PodcastID:   podcastItem.PodcastID,
			Title:       podcastItem.Title,
			Summary:     formatHTML(podcastItem.Summary),
			EpisodeType: podcastItem.EpisodeType,
			Duration:    podcastItem.Duration,
			PubDate:     podcastItem.PubDate,
			FileURL:     podcastItem.FileURL,
			GUID:        podcastItem.GUID,
			Image:       podcastItem.Image,
			LocalImage:  podcastItem.LocalImage,
			Podcast:     podcastItem.Podcast,
		}
		podcastItemsDTO = append(podcastItemsDTO, podcastItem)

	}
	return &podcastItemsDTO
}

func formatHTML(htmlString string) template.HTML {
	return template.HTML(
		strings.ReplaceAll(html.UnescapeString(htmlString), "\n", "<br>"),
	)
}
