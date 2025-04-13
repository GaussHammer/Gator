package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/GaussHammer/Gator/internal/config"
	"github.com/GaussHammer/Gator/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	dbUrl := "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable"

	cfg, err := config.Read()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		fmt.Println("error connecting to database", err)
	}
	defer db.Close()

	dbQueries := database.New(db)

	var s state
	s.cfg = &cfg
	s.cfg.DBURL = dbUrl
	s.db = dbQueries

	cmds := commands{commandMap: make(map[string]func(*state, command) error)}

	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerGetUsers)
	cmds.register("agg", handlerAggregate)
	cmds.register("addfeed", handlerCreateFeed)
	cmds.register("feeds", handlerFeeds)

	if len(os.Args) < 2 {
		fmt.Println("Error: not enough arguments provided")
		os.Exit(1)
	}
	cmd := command{name: os.Args[1],
		args: os.Args[2:]}

	err = cmds.run(&s, cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	commandMap map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.commandMap[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	handler, exists := c.commandMap[cmd.name]
	if !exists {
		return fmt.Errorf("unknown command: %s", cmd.name)
	}
	return handler(s, cmd)
}
