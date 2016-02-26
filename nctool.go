package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/serbe/ncp"
)

func (a *App) get() error {
	var (
		err error
		i   int64
	)
	for _, parseurl := range urls {
		topics, err := a.net.ParseForumTree(parseurl)
		if err != nil {
			log.Println("ParseForumTree ", err)
			return err
		}
		for _, topic := range topics {
			_, err := a.getFilmByHref(topic.Href)
			if err == gorm.RecordNotFound {
				film, err := a.net.ParseTopic(topic)
				if err == nil {
					i++
					film = a.checkName(film)
					a.createFilm(film)
				}
			}
		}
	}
	if i > 0 {
		log.Println("Adding", i, "new films")
	} else {
		log.Println("No adding new films")
	}
	return err
}

func (a *App) update() error {
	films, err := a.getWithTorrents()
	if err != nil {
		return err
	}
	for _, film := range films {
		var topic ncp.Topic
		topic.Href = film.Href
		f, err := a.net.ParseTopic(topic)
		if err == nil {
			if f.NNM != film.NNM || f.Seeders != film.Seeders || f.Leechers != f.Leechers || f.Torrent != film.Torrent {
				a.updateFilm(film.ID, f)
			}
		} else {
			return err
		}
	}
	return nil
}

func (a *App) name() error {
	films, err := a.getWithTorrents()
	if err != nil {
		return err
	}
	for _, film := range films {
		if film.Name == strings.ToUpper(film.Name) {
			lowerName := a.getLowerName(film)
			if lowerName != "" {
				a.updateName(film.ID, lowerName)
			}
		}
	}
	return nil
}

func main() {
	args := os.Args
	if contain(args, "help") {
		fmt.Println(`Usage:
	nctool COMMAND

Commands:
	help   показать справку
	get    получить новые фильмы
	update обновление информации фильмов
	name   поиск и исправление имен фильмов`)
		os.Exit(0)
	}
	if containCommand(args) == false {
		fmt.Println(`comand not found: use "nctool help"`)
		os.Exit(1)
	}
	app, err := appInit()
	if err != nil {
		os.Exit(1)
	}
	if contain(args, "get") {
		log.Println("Start getting new films")
		err := app.get()
		log.Println("End getting new films")
		exit(err)
	}
	if contain(args, "update") {
		log.Println("Start update topics")
		err := app.update()
		log.Println("End update topics")
		exit(err)
	}
	if contain(args, "name") {
		log.Println("Start fix names")
		err := app.name()
		log.Println("End fix names")
		exit(err)
	}
}
