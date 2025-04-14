package main

import (
	"context"
	"fmt"
	"time"

	"github.com/GaussHammer/Gator/internal/database"
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
		fmt.Println(item.Title)
	}
	return nil
}
