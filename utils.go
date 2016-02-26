package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/serbe/ncp"
)

var (
	urls = []string{
		"http://nnm-club.me/forum/viewforum.php?f=218",
		"http://nnm-club.me/forum/viewforum.php?f=270",
	}
	commands = []string{
		"get",
		"update",
		"name",
	}
)

// App struct variables
type App struct {
	db  gorm.DB
	net *ncp.NCp
}

type config struct {
	Nnm struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	} `json:"nnmclub"`
	Pq struct {
		User     string `json:"user"`
		Password string `json:"password"`
		Dbname   string `json:"dbname"`
		Sslmode  string `json:"sslmode"`
	} `json:"postgresql"`
}

func getConfig() (config, error) {
	c := config{}
	file, err := ioutil.ReadFile("./config.json")
	if err != nil {
		return c, err
	}
	err = json.Unmarshal(file, &c)
	return c, err
}

func contain(args []string, str string) bool {
	result := false
	for _, item := range args {
		if item == str {
			result = true
			return result
		}
	}
	return result
}

func containCommand(args []string) bool {
	result := false
	for _, item := range commands {
		if contain(args, item) {
			result = true
			return result
		}
	}
	return result
}

func exit(err error) {
	if err == nil {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}

func (a *App) checkName(film ncp.Film) ncp.Film {
	if film.Name != strings.ToUpper(film.Name) {
		return film
	}
	name := a.getFilmName(film)
	if name != "" {
		film.Name = name
		return film
	}
	return film
}
