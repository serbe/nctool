package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/serbe/ncp"
	// pq need to gorm
	_ "github.com/lib/pq"
)

func appInit() (*App, error) {
	conf, err := getConfig()
	if err != nil {
		panic(err)
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
	return a.db.Create(&film).Error
}

func (a *App) getFilmByHref(href string) (ncp.Film, error) {
	var film ncp.Film
	err := a.db.Where("href = ?", href).First(&film).Error
	return film, err
}

func (a *App) updateFilm(id int64, f ncp.Film) error {
	return a.db.Where("id = ?", id).UpdateColumns(ncp.Film{NNM: f.NNM, Seeders: f.Seeders, Leechers: f.Leechers, Torrent: f.Torrent}).Error
}

func (a *App) getWithTorrents() ([]ncp.Film, error) {
	var (
		films []ncp.Film
	)
	err := a.db.Where("torrent != ''").Find(&films).Error
	return films, err
}

func (a *App) getFilmName(film ncp.Film) string {
	var films []ncp.Film
	a.db.Where("upper(name) = ? and year = ?", strings.ToUpper(film.Name), film.Year).Find(&films)
	if len(films) > 0 {
		return films[0].Name
	}
	return ""
}
