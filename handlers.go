package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/GaussHammer/Gator/internal/database"
	"github.com/google/uuid"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("There must be at least one argument")
	}
	name := cmd.args[0]
	_, err := s.db.GetUser(context.Background(), name)
	if err != nil {
		// User doesn't exist
		fmt.Println("User does not exist!")
		os.Exit(1)
	}
	err = s.cfg.SetUser(cmd.args[0])
	if err != nil {
		return fmt.Errorf("Couldn't set the new user %w", err)
	}
	fmt.Println("user has been set")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		fmt.Println("Usage: register <name>")
		return errors.New("missing name argument")
	}
	name := cmd.args[0]
	_, err := s.db.GetUser(context.Background(), name)
	if err == nil {
		fmt.Println("This user already exists!")
		os.Exit(1)
	}
	params := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
	}

	// Now pass the params struct to CreateUser
	user, err := s.db.CreateUser(context.Background(), params)
	if err != nil {
		fmt.Println("Error inserting new user", err)
		os.Exit(1)
	}

	err = s.cfg.SetUser(user.Name)
	if err != nil {
		fmt.Println("could not set the user")
		os.Exit(1)
	}
	fmt.Printf("user %s was created\n", user.Name)
	fmt.Printf("User details: ID=%s, CreatedAt=%v, UpdatedAt=%v\n",
		user.ID, user.CreatedAt, user.UpdatedAt)
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.ResetUsers(context.Background())
	if err != nil {
		fmt.Printf("Error deleting users: %v", err)
		os.Exit(1)
	}
	return nil
}

func handlerGetUsers(s *state, cmd command) error {
	res, err := s.db.GetUsers(context.Background())
	if err != nil {
		fmt.Println("Error getting user names")
		os.Exit(1)
	}
	for i := 0; i < len(res); i++ {
		if res[i] == s.cfg.CurrentUserName {
			fmt.Println("* " + res[i] + " (current)")
		} else {
			fmt.Println("* " + res[i])
		}
	}
	return nil
}

func handlerAggregate(s *state, cmd command, u database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("Please provide the time between requests (e.g 1s, 1m, 1h)")
	}

	timeBetweenReqs, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("invalid duration format: %v", err)
	}

	fmt.Printf("Collecting feeds every: %v\n", timeBetweenReqs)
	ticker := time.NewTicker(timeBetweenReqs)

	for ; ; <-ticker.C {
		if err := scrapeFeeds(s); err != nil {
			fmt.Printf("Error scraping feeds: %v\n", err)
			// Continue even if there's an error
		}
	}
	return nil
}

func handlerCreateFeed(s *state, cmd command, u database.User) error {
	if len(cmd.args) < 2 {
		fmt.Println("Usage: addfeed \"feed name\" \"feed url\"")
		return fmt.Errorf("not enough arguments")
	}

	name := cmd.args[0]
	url := cmd.args[1]

	user := u
	//user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	//if err != nil {
	//	return fmt.Errorf("could not get the current user: %w", err)
	//}

	// Create the feed using the SQLC generated function
	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	})
	if err != nil {
		return fmt.Errorf("error creating feed: %w", err)
	}

	follow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})

	// Print the feed details
	fmt.Println("Feed created successfully!")
	fmt.Printf("ID: %s\n", feed.ID)
	fmt.Printf("Name: %s\n", feed.Name)
	fmt.Printf("URL: %s\n", feed.Url)
	fmt.Printf("User ID: %s\n", feed.UserID)
	fmt.Printf("Created At: %v\n", feed.CreatedAt)
	fmt.Printf("Updated At: %v\n", feed.UpdatedAt)
	fmt.Println(follow.FeedName)
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	results, err := s.db.SelectAllFeeds(context.Background())
	if err != nil {
		return err
	}
	for _, feed := range results {
		fmt.Printf("Name: %s, URL: %s, Creator: %s\n", feed.Name, feed.Url, feed.Name_2)
	}
	return nil
}

func handlerFollow(s *state, cmd command, u database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("Not enough arguments, needs an URL")
	}
	feed, err := s.db.SelectFeedByUrl(context.Background(), cmd.args[0])
	if err != nil {
		return err
	}
	user := u
	feedId := feed.ID
	userId := user.ID

	follow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    userId,
		FeedID:    feedId,
	})
	if err != nil {
		return fmt.Errorf("could not create the feed follow")
	}

	fmt.Printf("feed: %s, user: %s\n", follow.FeedName, follow.UserName)
	return nil
}

func handlerFollowing(s *state, cmd command) error {
	following, err := s.db.GetFeedFollowsForUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("could not bring the feeds that the user follows")
	}
	for _, feed := range following {
		fmt.Printf("User: %s, Feed: %s\n", feed.UserName, feed.FeedName)
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, u database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("Unfollow needs the URL of the site")
	}
	err := s.db.DeleteFeedFollowRecord(context.Background(), database.DeleteFeedFollowRecordParams{
		UserID: u.ID,
		Url:    cmd.args[0],
	})
	if err != nil {
		return err
	}
	return nil
}

func handlerBrowse(s *state, cmd command, u database.User) error {
	var limit int
	if len(cmd.args) < 1 {
		limit = 2
	} else {
		var err error
		limit, err = strconv.Atoi(cmd.args[0])
		if err != nil {
			return fmt.Errorf("limit must be a valid number")
		}
	}
	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: u.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		return fmt.Errorf("couldn't get the posts for the current user")
	}
	fmt.Printf("retrived posts: %d\n", len(posts))
	fmt.Printf("user Id: %s\n", u.ID)
	for _, items := range posts {
		fmt.Println(items.Title)
		fmt.Printf("\n")
		fmt.Println("Url: " + items.Url)
		fmt.Printf("\n")
		fmt.Println("Description: " + items.Description)
		fmt.Printf("\n")
		fmt.Println(items.PublishedAt)

	}
	return nil
}
