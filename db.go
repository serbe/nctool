package main

import (
	"fmt"
	"log"
	// "strconv"
	"strings"
	// "time"

	"github.com/jinzhu/gorm"
	"github.com/serbe/kpp"
	"github.com/serbe/ncp"
	// pq need to gorm
	_ "github.com/lib/pq"
)

func appInit() (*App, error) {
	conf, err := getConfig()
	if err != nil {
		log.Fatal(err)
	}
	dbConnect, err := gorm.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s", conf.Pq.User, conf.Pq.Password, conf.Pq.Dbname, conf.Pq.Sslmode))
	if err != nil {
		log.Fatal("db open ", err)
		return &App{}, err
	}
	dbConnect.DB()
	dbConnect.DB().Ping()
	dbConnect.DB().SetMaxIdleConns(10)
	dbConnect.DB().SetMaxOpenConns(100)
	dbConnect.AutoMigrate(&ncp.Film{})
	// dbConnect.LogMode(true)
	inetConnect, err := ncp.Init(conf.Nnm.Login, conf.Nnm.Password)
	if err != nil {
		log.Println("net init ", err)
		return &App{}, err
	}
	return &App{db: dbConnect, net: inetConnect}, nil
}

func (a *App) createFilm(film ncp.Film) error {
	return a.db.Model(ncp.Film{}).Create(&film).Error
}

func (a *App) getFilmByHref(href string) (ncp.Film, error) {
	var film ncp.Film
	err := a.db.Model(ncp.Film{}).Where("href = ?", href).First(&film).Error
	return film, err
}

func (a *App) updateFilm(id int64, f ncp.Film) error {
	return a.db.Model(ncp.Film{}).Where("id = ?", id).UpdateColumns(ncp.Film{NNM: f.NNM, Seeders: f.Seeders, Leechers: f.Leechers, Torrent: f.Torrent}).Error
}

func (a *App) updateName(id int64, name string) error {
	return a.db.Model(ncp.Film{}).Where("id = ?", id).UpdateColumn("name", name).Error
}

func (a *App) updateRating(film ncp.Film, kinopoisk float64, imdb float64) error {
	return a.db.Model(ncp.Film{}).Where("upper(name) = ? and year = ?", strings.ToUpper(film.Name), film.Year).UpdateColumns(ncp.Film{Kinopoisk: kinopoisk, IMDb: imdb}).Error
}

func (a *App) getWithTorrents() ([]ncp.Film, error) {
	var (
		films []ncp.Film
	)
	err := a.db.Where("torrent != ''").Find(&films).Error
	return films, err
}

func (a *App) getFilmName(film ncp.Film) (string, error) {
	var films []ncp.Film
	a.db.Model(ncp.Film{}).Where("upper(name) = ? and year = ?", strings.ToUpper(film.Name), film.Year).Find(&films)
	if len(films) > 0 {
		return films[0].Name, nil
	}
	return "", fmt.Errorf("Name not found")
}

func (a *App) getLowerName(film ncp.Film) (string, error) {
	var f ncp.Film
	err := a.db.Model(ncp.Film{}).Where("upper(name) = ? and year = ? and name != ?", strings.ToUpper(film.Name), film.Year, strings.ToUpper(film.Name)).First(&f).Error
	return f.Name, err
}

func (a *App) getNoRating() ([]ncp.Film, error) {
	var films []ncp.Film
	err := a.db.Raw("SELECT name, year FROM films WHERE torrent <> '' AND (kinopoisk = 0 OR imdb = 0) GROUP BY name, year;").Scan(&films).Error
	return films, err
}

func (a *App) getRating(film ncp.Film) error {
	var kp kpp.KP
	kp, err := kpp.GetRating(film.Name, film.Year)
	if err != nil {
		return fmt.Errorf("Rating no found")
	}
	if kp.Kinopoisk == 0 && kp.IMDb == 0 {
		return fmt.Errorf("Rating no found")
	}
	return a.updateRating(film, kp.Kinopoisk, kp.IMDb)
}
