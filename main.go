package main

import (
	"aggreGATOR/internal/config"
	"fmt"
)

const username = "username"
const dbUrl = "postgres://example"

func main() {
	cfg, err := config.Read()
	if err != nil {
		//fmt.Println("First read error: ", err)
		fmt.Println("Creating a new file.")
	}

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
}
