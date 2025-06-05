package main

import (
	"aggreGATOR/internal/config"
	"aggreGATOR/internal/database"
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
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

func handlerLogin(s *state, cmd command) error {
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
	err := s.db.Reset(context.Background())
	if err != nil {
		return err
	}

	fmt.Println("User table reset")
	s.config.SetUser("")
	return nil
}

func handlerGetUsers(s *state, cmd command) error {
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
	feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return err
	}
	fmt.Println(feed)
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
	c.register("login", handlerLogin)
	c.register("register", handlerRegister)
	c.register("users", handlerGetUsers)
	c.register("user", handlerWho)
	c.register("reset", handlerReset)
	c.register("agg", handlerAggregate)
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
