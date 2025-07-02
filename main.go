package main

import (
	"aggreGATOR/internal/config"
	"aggreGATOR/internal/database"
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

const username = "username"
const dbURL = "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable"

type state struct {
	db     *database.Queries
	config *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	callbacks map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	callback, exists := c.callbacks[cmd.name]
	if !exists {
		return fmt.Errorf(`no command with a name of "%s" exists`, cmd.name)
	}
	err := callback(s, cmd)
	if err != nil {
		return fmt.Errorf(`error running command "%s": %w`, cmd.name, err)
	}

	return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.callbacks[name] = f
}

func handlerLogin(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("login expects at least a single argument")
	}

	_, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("given user does not exists: %w", err)
	}

	err = s.config.SetUser(cmd.args[0])
	if err != nil {
		return err
	}

	fmt.Println("User has been set to:", cmd.args[0])
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("register expects at least a single argument")
	}
	now := time.Now()
	user, err := s.db.CreateUser(context.Background(),
		database.CreateUserParams{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
			Name:      cmd.args[0],
		})
	if err != nil {
		return fmt.Errorf("user alredy exists: %w", err)
	}

	s.config.SetUser(user.Name)

	fmt.Println("A new User has been created with the name:", cmd.args[0])
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.ResetUsers(context.Background())
	if err != nil {
		return err
	}

	fmt.Println("User table reset")
	s.config.SetUser("")

	err = s.db.ResetFeedFollows(context.Background())
	if err != nil {
		return err
	}

	fmt.Println("FeedFollows table reset")
	return nil
}

func handlerGetUsers(s *state, cmd command, user database.User) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}

	for _, user := range users {
		fmt.Print("*", user.Name)
		if user.Name == s.config.Current_user_name {
			fmt.Print(" (current)")
		}
		fmt.Println()
	}
	return nil
}

func handlerAggregate(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("aggregate expects at least a single argument")
	}

	timeBetweenReqs, err := time.ParseDuration(cmd.args[0])
	if len(cmd.args) == 0 {
		return fmt.Errorf("arument given to aggregate command is suspect: %w", err)
	}

	fmt.Printf("Collecting feeds every: %s\n", cmd.args[0])

	ticker := time.NewTicker(timeBetweenReqs)
	for ; ; <-ticker.C {
		err = scrapeFeeds(s)
		if err != nil {
			fmt.Printf("error scraping feed: %w\n\n", err)
		}
	}

	return nil
}

func handlerWho(s *state, cmd command) error {
	if s.config.Current_user_name == "" {
		fmt.Println("Not logged in as any user.")
		return nil
	}
	fmt.Println("Loged in as:", s.config.Current_user_name)
	return nil
}

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	// Prepare and send the http request
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("Failed to make request: ", err)
	}
	req.Header.Set("User-Agent", "gator")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("Failed get: ", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &RSSFeed{}, fmt.Errorf("Failed status: ", err)
	}
	// Unmarshall what we got
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("Failed ReadAll: ", err)
	}

	var ret RSSFeed
	err = xml.Unmarshal([]byte(body), &ret)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("Failed Unmarshal: ", err)
	}
	ret.Channel.Title = html.UnescapeString(ret.Channel.Title)
	ret.Channel.Description = html.UnescapeString(ret.Channel.Description)
	return &ret, nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) <= 1 {
		return fmt.Errorf("addfeed expects at least a two arguments")
	}

	user, err := s.db.GetUser(context.Background(), s.config.Current_user_name)
	if err != nil {
		return fmt.Errorf("error geting current user: %w", err)
	}

	now := time.Now()
	feed, err := s.db.CreateFeeds(context.Background(),
		database.CreateFeedsParams{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
			Name:      cmd.args[0],
			Url:       cmd.args[1],
			UserID: uuid.NullUUID{
				UUID:  user.ID,
				Valid: true,
			},
		})

	if err != nil {
		return fmt.Errorf("create feed error: %w", err)
	}

	fmt.Printf(`A new Feed has been created with the name: "%s" and url: "%s"`, feed.Name, feed.Url)
	fmt.Println()

	_, err = s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		UserID: uuid.NullUUID{
			UUID:  user.ID,
			Valid: true,
		},
		FeedID: uuid.NullUUID{
			UUID:  feed.ID,
			Valid: true,
		},
	})
	if err != nil {
		return fmt.Errorf("error creating feed follow: %w", err)
	}
	fmt.Printf(`A new FeedFollow has been created with the name: "%s" and url: "%s" for user: "%s"`, feed.Name, feed.Url, user.Name)
	fmt.Println()

	return nil
}

func handlerGetFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("error geting feeds: %w", err)
	}

	fmt.Println("Feed name, Feed url, User name")
	for i := range feeds {
		fmt.Printf("* %s, %s, %s\n", feeds[i].FeedName, feeds[i].FeedUrl, feeds[i].UserName)
	}

	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("follow expects at least a single argument")
	}

	now := time.Now()
	feed, err := s.db.GetFeedByURL(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("error on get feed by url: %w", err)
	}

	_, err = s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		UserID: uuid.NullUUID{
			UUID:  user.ID,
			Valid: true,
		},
		FeedID: uuid.NullUUID{
			UUID:  feed.ID,
			Valid: true,
		},
	})
	if err != nil {
		return fmt.Errorf("error creating feed follow: %w", err)
	}
	fmt.Printf(`A new FeedFollow has been created with the name: "%s" and url: "%s" for user: "%s"`, feed.Name, feed.Url, user.Name)
	fmt.Println()

	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	follows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("error geting current users follows: %w", err)
	}

	if len(follows) == 0 {
		fmt.Println("You do not follow any feed right now")
	} else {
		fmt.Println("You are following:")
	}

	for _, feed := range follows {
		fmt.Printf("%s, %s\n", feed.FeedName, feed.FeedUrl)
	}

	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("follow expects at least a single argument")
	}

	feedfollow, err := s.db.GetFeedFollowsForUserByURL(context.Background(), database.GetFeedFollowsForUserByURLParams{
		ID:  user.ID,
		Url: cmd.args[0],
	})

	if err != nil { // disgusting
		feedfollow_2, err := s.db.GetFeedFollowsForUserByName(context.Background(), database.GetFeedFollowsForUserByNameParams{
			ID:   user.ID,
			Name: cmd.args[0],
		})
		if err != nil {
			return fmt.Errorf("error getting FollowsForUser: %w", err)
		}

		err = s.db.DeleteFeedById(context.Background(), feedfollow_2.FeedID)
		if err != nil {
			return fmt.Errorf("error deleteing feed record: %w", err)
		}
	} else {
		err = s.db.DeleteFeedById(context.Background(), feedfollow.FeedID)

		if err != nil {
			return fmt.Errorf("error deleteing feed record: %w", err)
		}
	}

	return nil
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	newHandler := func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.config.Current_user_name)
		if err != nil {
			return fmt.Errorf("error getting current user: %w", err)
		}

		return handler(s, cmd, user)
	}

	return newHandler
}

func TruncateTooLongString(s string, maxLength int) string {
	if len(s) > maxLength {
		return s[:maxLength]
	}
	return s
}

func scrapeFeeds(s *state) error {
	sqlFeed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("error getting oldest feed: %w", err)
	}

	err = s.db.MarkFeedFetched(context.Background(), database.MarkFeedFetchedParams{
		ID: sqlFeed.ID,
		LastFetchedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
	})

	if err != nil {
		return fmt.Errorf("error marking oldest feed as fetched: %w", err)
	}

	rssFeed, err := fetchFeed(context.Background(), sqlFeed.Url)

	if err != nil {
		return fmt.Errorf("error fetching RSS Feed: %w", err)
	}

	now := time.Now()
	fmt.Printf("\n\nFeed: %s\n", rssFeed.Channel.Title)
	for num, item := range rssFeed.Channel.Item {
		fmt.Printf("Item num: %d/%d | ", num+1, len(rssFeed.Channel.Item))
		parsedTime, timeErr := time.Parse(time.RFC1123, item.PubDate)
		// just ignoring the pub date in case of errors on given item
		if timeErr != nil {
			fmt.Printf("Error ocured on TIME parse of Title: %s\n", item.Title)
		}

		itemTitle := TruncateTooLongString(item.Title, 255)
		itemDescription := TruncateTooLongString(item.Description, 255)
		if len(item.Link) > 255 {
			fmt.Printf("Cant add post, too long url Title: %s\n", item.Title)
			continue
		}
		_, postErr := s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
			Title:     itemTitle,
			Url:       item.Link,
			Description: sql.NullString{
				String: itemDescription,
				Valid:  itemDescription != "",
			},
			PublishedAt: sql.NullTime{
				Time:  parsedTime,
				Valid: timeErr != nil || item.PubDate != "",
			},
			FeedID: uuid.NullUUID{
				UUID:  sqlFeed.ID,
				Valid: true,
			},
		})

		if postErr != nil {
			var pqErr *pq.Error
			if errors.As(postErr, &pqErr) && pqErr.Code == "23505" {
				fmt.Printf("Post already exists. Ignoring. Title: %s\n", item.Title)
				continue
			}
			fmt.Printf("Error ocured on item of Title: %s\n", item.Title)
			fmt.Println(postErr)
		} else {
			fmt.Printf("Post Created for Title: %s\n", item.Title)
		}

	}

	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	var limit int32 = int32(2)
	if len(cmd.args) == 1 {
		num_, err := strconv.Atoi(cmd.args[0])
		if err != nil {
			return fmt.Errorf("invalid limit, must be number")
		}

		limit = int32(num_)
	}

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: uuid.NullUUID{
			UUID:  user.ID,
			Valid: true,
		},
		Limit:  limit,
		Offset: 0,
	})

	if err != nil {
		return fmt.Errorf("Error geting posts for user: %w", err)
	}

	fmt.Println("Posts: ")
	for _, post := range posts {
		fmt.Printf("Date: %s, Title: %s\nDescription:\n", post.PublishedAt.Time.Format("2006-01-02"), post.Title)
		fmt.Println(post.Description.String)
		fmt.Println()
	}

	return nil
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println("No config file found.")
		fmt.Println("Creating a new file.")
	}
	var s *state
	s = new(state)
	s.config = &cfg

	db, err := sql.Open("postgres", s.config.Db_url)
	if err != nil {
		fmt.Println("Could not open database.")
		os.Exit(1)
	}

	dbQueries := database.New(db)
	s.db = dbQueries

	var c commands
	c.callbacks = make(map[string]func(*state, command) error)
	c.register("login", middlewareLoggedIn(handlerLogin))
	c.register("register", handlerRegister)
	c.register("users", middlewareLoggedIn(handlerGetUsers))
	c.register("user", handlerWho)
	c.register("reset", handlerReset)
	c.register("agg", handlerAggregate)
	c.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	c.register("feeds", handlerGetFeeds)
	c.register("follow", middlewareLoggedIn(handlerFollow))
	c.register("following", middlewareLoggedIn(handlerFollowing))
	c.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	c.register("browse", middlewareLoggedIn(handlerBrowse))
	c.register("help", func(*state, command) error { return fmt.Errorf("No. https://www.youtube.com/watch?v=gWm2NzNLc_A") })
	if len(os.Args) == 1 {
		fmt.Println(`No command was specified. Try "help".`)
		os.Exit(1)
	}

	name := os.Args[1]
	var args []string
	if len(os.Args) == 2 {
		args = []string{}
	} else {
		args = os.Args[2:]
	}

	cmd := command{
		name,
		args,
	}
	err = c.run(s, cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	/*

		cfg, err = cfg.SetUser(username)
		if err != nil {
			fmt.Println("SetUser error: ", err)
			return
		}

		cfg, err = cfg.SetDbUrl(dbUrl)
		if err != nil {
			fmt.Println("SetDbUrl error: ", err)
			return
		}

		cfg, err = config.Read()
		if err != nil {
			fmt.Println("Second read error: ", err)
			return
		}

		cfg_str, err := cfg.AsByte()
		if err != nil {
			fmt.Println("AsByte error: ", err)
			return
		}

		fmt.Print(string(cfg_str))
	*/
}
