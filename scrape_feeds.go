package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/GaussHammer/Gator/internal/database"
	"github.com/google/uuid"
)

func scrapeFeeds(s *state) error {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("Error fetching next feed")
	}
	now := time.Now()
	err = s.db.MarkFeedFetched(context.Background(), database.MarkFeedFetchedParams{
		UpdatedAt: now,
		ID:        feed.ID,
	})
	if err != nil {
		return err
	}

	posts, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		return err
	}

	for _, item := range posts.Channel.Item {
		pubDate, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			pubDate = time.Now()
		}
		err = s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       item.Title,
			Url:         item.Link,
			Description: item.Description,
			PublishedAt: pubDate,
			FeedID:      feed.ID,
		})
		if err != nil {
			// Check for duplicate URL error
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") ||
				strings.Contains(err.Error(), "UNIQUE constraint failed") {
				// Just ignore these errors and continue
				continue
			}
			// For other errors, log them
			log.Printf("Error creating post: %v", err)
		}
	}
	return nil
}
