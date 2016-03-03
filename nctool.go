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
			_, err := a.getTorrentByHref(topic.Href)
			if err == gorm.RecordNotFound {
				film, err := a.net.ParseTopic(topic)
				if err == nil {
					i++
					film = a.checkName(film)
					a.createTorrent(film)
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
	var (
		i        int64
		err      error
		torrents []Torrent
	)

	torrents, err = a.getWithDownload()
	if err != nil {
		return err
	}
	for _, tor := range torrents {
		var topic ncp.Topic
		topic.Href = tor.Href
		f, err := a.net.ParseTopic(topic)
		if err == nil {
			if f.NNM != tor.NNM || f.Seeders != tor.Seeders || f.Leechers != tor.Leechers || f.Torrent != tor.Torrent {
				i++
				a.updateTorrent(tor.ID, f)
			}
		} else {
			return err
		}
	}
	if i > 0 {
		log.Println("Update", i, "films")
	} else {
		log.Println("No films update")
	}
	return nil
}

func (a *App) name() error {
	var i int64
	films, err := a.getFilms()
	if err != nil {
		return err
	}
	for _, film := range films {
		if film.Name == strings.ToUpper(film.Name) {
			lowerName, err := a.getLowerName(film)
			if err == nil {
				i++
				a.updateName(film.ID, lowerName)
			}
		}
	}
	if i > 0 {
		log.Println(i, "name fixed")
	} else {
		log.Println("No fixed names")
	}
	return nil
}

func (a *App) rating() error {
	var (
		i int64
	)
	films, err := a.getNoRating()
	if err != nil {
		return err
	}
	for _, film := range films {
		if film.Kinopoisk == 0 || film.IMDb == 0 {
			kp, err := a.getRating(film)
			if err == nil {
				i++
				_ = a.updateRating(film, kp)
			}
		}
	}
	if i > 0 {
		log.Println(i, "ratings update")
	} else {
		log.Println("No update ratings")
	}
	return nil
}

func main() {
	args := os.Args
	if contain(args, "help") {
		fmt.Println(`Usage:
	nctool COMMAND

Commands:
	help    показать справку
	get     получить новые фильмы
	update  обновление информации фильмов
	name    поиск и исправление имен фильмов
	rating  получение рейтинга Кинопоиска и IMDb`)
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
	if contain(args, "rating") {
		log.Println("Start get ratings")
		err := app.rating()
		log.Println("End get ratings")
		exit(err)
	}
}
