package main

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/serbe/ncp"
)

var (
	urls = []string{
		"/forum/viewforum.php?f=218", // Зарубежные Новинки (HD*Rip/LQ, DVDRip)
		"/forum/viewforum.php?f=221", // Отечественные Фильмы (HD*Rip/LQ, DVDRip, SATRip, VHSRip)
		"/forum/viewforum.php?f=225", // Зарубежные Фильмы (HD*Rip/LQ, DVDRip, SATRip, VHSRip)
		"/forum/viewforum.php?f=230", // Отечественные Мультфильмы (HD*Rip/LQ, DVDRip, SATRip, VHSRip)
		"/forum/viewforum.php?f=231", // Зарубежные Мультфильмы (HD*Rip/LQ, DVDRip, SATRip, VHSRip)
		"/forum/viewforum.php?f=270", // Отечественные Новинки (HD*Rip/LQ, DVDRip)
		"/forum/viewforum.php?f=319", // Зарубежная Классика (HD*Rip/LQ, DVDRip, SATRip, VHSRip)
		"/forum/viewforum.php?f=320", // Отечественная Классика (HD*Rip/LQ, DVDRip, SATRip, VHSRip)
	}
)

func (a *App) get() error {
	var (
		err    error
		i      int
		topics []ncp.Topic
	)
	for _, parseurl := range urls {
		topics, err = a.net.ParseForumTree(parseurl)
		if err != nil {
			log.Println("ParseForumTree ", parseurl, err)
			return err
		}
		for _, topic := range topics {
			_, err = a.getTorrentByHref(topic.Href)
			if err != nil {
				var film ncp.Film
				film, err = a.net.ParseTopic(topic)
				if err == nil {
					if film.Description != "" {
						i++
						film = a.checkName(film)
						_, err = a.createTorrent(film)
						if err != nil {
							log.Println("createTorrent ", err)
						}
					} else {
						log.Println("empty Description ", film.Href)
					}
				} else {
					log.Println("ParseTopic ", err)
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
		i        int
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
				_ = a.updateTorrent(tor.ID, f)
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
	var i int
	movies, err := a.getMovies()
	if err != nil {
		return err
	}
	for _, movie := range movies {
		if movie.Name == strings.ToUpper(movie.Name) {
			lowerName, err := a.getUpperName(movie)
			if err == nil {
				i++
				_ = a.updateName(movie.ID, lowerName)
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
		i int
	)
	movies, err := a.getNoRatingMovies()
	if err != nil {
		return err
	}
	for _, movie := range movies {
		if movie.Kinopoisk == 0 || movie.IMDb == 0 || movie.Duration == "" {
			kp, err := a.getRating(movie)
			if err == nil {
				i++
				_ = a.updateRating(movie, kp)
				if a.debug {
					log.Println(movie.Name, movie.EngName, movie.Year)
					log.Println(kp)
				}
			} else {
				if a.debug {
					log.Println(movie.Name, movie.EngName, movie.Year)
					log.Println("updateRating: ", err)
				}
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
		i     int
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
		if movie.PosterURL != "" {
			if movie.Poster != "" {
				if !existsFile(a.hd + movie.Poster) {
					poster, err := a.getPoster(movie.PosterURL)
					if err == nil {
						i++
						err = a.updatePoster(movie, poster)
						if err != nil {
							log.Println("updatePoster ", poster, err)
						}
					} else {
						log.Println("getPoster ", poster, err)
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
		} else {
			tor, err := a.getTorrentByMovieID(movie.ID)
			if err == nil {
				var topic ncp.Topic
				topic.Href = tor.Href
				tempFilm, err := a.net.ParseTopic(topic)
				if err == nil {
					if tempFilm.Poster != "" {
						err = a.updatePosterURL(movie, tempFilm.Poster)
						if err == nil {
							poster, err := a.getPoster(tempFilm.Poster)
							if err == nil {
								i++
								_ = a.updatePoster(movie, poster)
							}
						}
					}
				}
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
		log.Println("Remove", len(files), "unused images")
	}
	return nil
}
