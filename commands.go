package main

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/serbe/ncp"
)

var (
	commands = []string{
		"get",
		"update",
		"name",
		"rating",
		"poster",
	}
)

func (a *App) get() error {
	var (
		err error
		i   int64
	)
	for _, parseurl := range urls {
		topics, err := a.net.ParseForumTree(parseurl, false)
		if err != nil {
			log.Println("ParseForumTree ", parseurl, err)
			return err
		}
		for _, topic := range topics {
			_, err := a.getTorrentByHref(topic.Href)
			if err == gorm.ErrRecordNotFound {
				film, err := a.net.ParseTopic(topic, false)
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
		f, err := a.net.ParseTopic(topic, false)
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
		log.Println("Update", i, "movies")
	} else {
		log.Println("No movies update")
	}
	return nil
}

func (a *App) name() error {
	var i int64
	movies, err := a.getMovies()
	if err != nil {
		return err
	}
	for _, movie := range movies {
		if movie.Name == strings.ToUpper(movie.Name) {
			lowerName, err := a.getLowerName(movie)
			if err == nil {
				i++
				a.updateName(movie.ID, lowerName)
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
	movies, err := a.getNoRating()
	if err != nil {
		return err
	}
	for _, movie := range movies {
		if movie.Kinopoisk == 0 || movie.IMDb == 0 || movie.Duration == "" {
			kp, err := a.getRating(movie)
			if err == nil {
				i++
				_ = a.updateRating(movie, kp)
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

func (a *App) poster() error {
	var (
		i     int64
		files []string
	)
	movies, err := a.getMovies()
	if err != nil {
		return err
	}
	filesInDir, _ := ioutil.ReadDir(a.hd)
	for _, file := range filesInDir {
		files = append(files, file.Name())
	}
	for _, movie := range movies {
		if movie.Poster != "" {
			if existsFile(a.hd+movie.Poster) == false {
				poster, err := a.getPoster(movie.PosterURL)
				if err == nil {
					i++
					_ = a.updatePoster(movie, poster)
				}
			} else {
				files = deleteFromSlice(files, movie.Poster)
			}
		} else {
			poster, err := a.getPoster(movie.PosterURL)
			if err == nil {
				i++
				_ = a.updatePoster(movie, poster)
			}
		}
	}
	for _, file := range files {
		_ = os.Remove(a.hd + file)
	}
	if i > 0 {
		log.Println(i, "posters update")
	} else {
		log.Println("No update posters")
	}
	if len(files) > 0 {
		log.Println("Remove ", len(files), "unused images")
	}
	return nil
}
