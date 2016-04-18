package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/serbe/kpp"
	"github.com/serbe/ncp"
	"gopkg.in/pg.v4"
)

// Movie all values
// TableName     Название таблицы
// ID            id
// Section       Раздел форума
// Name          Название
// EngName       Английское название
// Year          Год
// Genre         Жанр
// Country       Производство
// Director      Режиссер
// Producer      Продюсер
// Actors        Актеры
// Description   Описание
// Age           Возраст
// ReleaseDate   Дата мировой премьеры
// RussianDate   Дата премьеры в России
// Duration      Продолжительность
// Kinopoisk     Рейтинг кинопоиска
// Imdb          Рейтинг IMDb
// Poster        Имя файла постера
// PosterURL     Сетевая ссылка на постер
type Movie struct {
	ID          int64    `sql:"id"`
	Section     string   `sql:"section"`
	Name        string   `sql:"name"`
	EngName     string   `sql:"eng_name"`
	Year        int64    `sql:"year"`
	Genre       []string `sql:"genre"        pg:",array" `
	Country     []string `sql:"country"      pg:",array"`
	RawCountry  string   `sql:"raw_country"`
	Director    []string `sql:"director"     pg:",array"`
	Producer    []string `sql:"producer"     pg:",array"`
	Actor       []string `sql:"actor"        pg:",array"`
	Description string   `sql:"description"`
	Age         string   `sql:"age"`
	ReleaseDate string   `sql:"release_date"`
	RussianDate string   `sql:"russian_date"`
	Duration    string   `sql:"duration"`
	Kinopoisk   float64  `sql:"kinopoisk"`
	IMDb        float64  `sql:"imdb"`
	Poster      string   `sql:"poster"`
	PosterURL   string   `sql:"poster_url"`
}

// Torrent all values
// TableName     Название таблицы
// ID            id
// FilmID        Указатель на фильм
// DateCreate    Дата создания раздачи
// Href          Ссылка
// Torrent       Ссылка на torrent
// Magnet        Ссылка на magnet
// NNM           Рейтинг nnm-club
// SubtitlesType Вид субтитров
// Subtitles     Субтитры
// Video         Видео
// Quality       Качество видео
// Resolution    Разрешение видео
// Audio1        Аудио1
// Audio2        Аудио2
// Audio3        Аудио3
// Translation   Перевод
// Size          Размер
// Seeders       Количество раздающих
// Leechers      Количество скачивающих
type Torrent struct {
	ID            int64   `sql:"id"`
	MovieID       int64   `sql:"movie_id"`
	DateCreate    string  `sql:"date_create"`
	Href          string  `sql:"href"`
	Torrent       string  `sql:"torrent"`
	Magnet        string  `sql:"magnet"`
	NNM           float64 `sql:"nnm"`
	SubtitlesType string  `sql:"subtitles_type"`
	Subtitles     string  `sql:"subtitles"`
	Video         string  `sql:"video"`
	Quality       string  `sql:"quality"`
	Resolution    string  `sql:"resolution"`
	Audio1        string  `sql:"audio1"`
	Audio2        string  `sql:"audio2"`
	Audio3        string  `sql:"audio3"`
	Translation   string  `sql:"translation"`
	Size          int64   `sql:"size"`
	Seeders       int64   `sql:"seeders"`
	Leechers      int64   `sql:"leechers"`
}

// App struct variables
type App struct {
	db    *pg.DB
	net   *ncp.NCp
	hd    string
	px    string
	debug bool
}

var app *App

func appInit() (*App, error) {
	if app == nil {
		app = new(App)
		conf, err := getConfig()
		if err != nil {
			log.Fatal("Error getConfig ", err)
		}
		db := pg.Connect(&pg.Options{
			Database: conf.Pq.Dbname,
			User:     conf.Pq.User,
			Password: conf.Pq.Password,
			SSL:      conf.Pq.Sslmode,
		})
		// err = db.Close()
		// if err != nil {
		// 	log.Fatal(err)
		// }
		err = createSchema(db)
		if err != nil {
			log.Fatal("Error createSchema ", err)
		}

		app.db = db

		inetConnect, err := ncp.Init(conf.Nnm.Login, conf.Nnm.Password, conf.Address, conf.Px)
		if err != nil {
			log.Println("net init ", err)
			return app, err
		}
		_ = createDir(conf.Hd)

		app.net = inetConnect
		app.hd = conf.Hd
		app.px = conf.Px
		app.debug = conf.Debug
	}
	return app, nil
}

func createSchema(db *pg.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS movies (
			id bigserial,
			section text,
			name text,
			eng_name text,
			year bigint,
			genre text[],
			country text[],
			raw_country text,
			director text[],
			producer text[],
			actor text[],
			description text,
			age text,
			release_date text,
			russian_date text,
			duration text,
			kinopoisk double precision,
			imdb double precision,
			poster text,
			poster_url text
        )`,
		`CREATE TABLE IF NOT EXISTS torrents (
			id bigserial,
			movie_id bigint,
			date_create text,
			href text,
			torrent text,
			magnet text,
			nnm numeric,
			subtitles_type text,
			subtitles text,
			video text,
			quality text,
			resolution text,
			audio1 text,
			audio2 text,
			audio3 text,
			translation text,
			size bigint,
			seeders bigint,
			leechers bigint
		)`,
	}
	for _, q := range queries {
		_, err := db.Exec(q)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *App) createMovie(ncf ncp.Film) (int64, error) {
	var (
		movie Movie
		kp    kpp.KP
	)
	movie.Section = ncf.Section
	movie.Name = ncf.Name
	movie.EngName = ncf.EngName
	movie.Year = ncf.Year
	movie.Genre = ncf.Genre
	movie.Country = ncf.Country
	movie.RawCountry = ncf.RawCountry
	movie.Director = ncf.Director
	movie.Producer = ncf.Producer
	movie.Actor = ncf.Actor
	movie.Description = ncf.Description
	movie.Age = ncf.Age
	movie.ReleaseDate = ncf.ReleaseDate
	movie.RussianDate = ncf.RussianDate
	movie.Duration = ncf.Duration
	kp, err := a.getRating(movie)
	if err == nil {
		movie.Kinopoisk = kp.Kinopoisk
		movie.IMDb = kp.IMDb
	}
	movie.PosterURL = ncf.Poster
	movie.Poster, _ = a.getPoster(movie.PosterURL)
	err = a.db.Create(&movie)
	return movie.ID, err
}

func (a *App) createTorrent(ncf ncp.Film) error {
	var (
		tor Torrent
		err error
	)
	tor.MovieID, err = a.getMovieID(ncf)
	if err != nil {
		id, err := a.createMovie(ncf)
		if err != nil {
			return err
		}
		tor.MovieID = id
	}
	tor.DateCreate = ncf.DateCreate
	tor.Href = ncf.Href
	tor.Torrent = ncf.Torrent
	tor.Magnet = ncf.Magnet
	tor.NNM = ncf.NNM
	tor.SubtitlesType = ncf.SubtitlesType
	tor.Subtitles = ncf.Subtitles
	tor.Video = ncf.Video
	tor.Quality = ncf.Quality
	tor.Resolution = ncf.Resolution
	tor.Audio1 = ncf.Audio1
	tor.Audio2 = ncf.Audio2
	tor.Audio3 = ncf.Audio3
	tor.Translation = ncf.Translation
	tor.Size = ncf.Size
	tor.Seeders = ncf.Seeders
	tor.Leechers = ncf.Leechers
	err = a.db.Create(&tor)
	return err
}

func (a *App) getMovies() ([]Movie, error) {
	var movies []Movie
	err := a.db.Model(&movies).Select()
	return movies, err
}

func (a *App) getMovieID(ncf ncp.Film) (int64, error) {
	var movie Movie
	err := a.db.Model(&movie).Where("name = ? AND year = ?", ncf.Name, ncf.Year).First()
	return movie.ID, err
}

func (a *App) getTorrentByHref(href string) (Torrent, error) {
	var tor Torrent
	err := a.db.Model(&tor).Where("href = ?", href).First()
	return tor, err
}

func (a *App) updateTorrent(id int64, f ncp.Film) error {
	return a.db.Update(Torrent{ID: id, NNM: f.NNM, Seeders: f.Seeders, Leechers: f.Leechers, Torrent: f.Torrent})
}

func (a *App) updateName(id int64, name string) error {
	return a.db.Update(Movie{ID: id, Name: name})
}

func (a *App) updateRating(movie Movie, kp kpp.KP) error {
	var duration string
	if movie.Duration == "" {
		duration = kp.Duration
	} else {
		duration = movie.Duration
	}
	return a.db.Update(Movie{ID: movie.ID, Kinopoisk: kp.Kinopoisk, IMDb: kp.IMDb, Duration: duration})
}

func (a *App) updatePoster(movie Movie, poster string) error {
	return a.db.Update(Movie{ID: movie.ID, Poster: poster})
}

func (a *App) updatePosterURL(movie Movie, poster string) error {
	return a.db.Update(Movie{ID: movie.ID, PosterURL: poster})
}

func (a *App) getWithDownload() ([]Torrent, error) {
	var (
		torrents []Torrent
	)
	err := a.db.Model(&torrents).Where("magnet != ''").Select()
	return torrents, err
}

func (a *App) getMovieName(ncf ncp.Film) (string, error) {
	var movies []Movie
	a.db.Model(&movies).Where("upper(name) = ? and year = ?", strings.ToUpper(ncf.Name), ncf.Year).Select()
	if len(movies) > 0 {
		return movies[0].Name, nil
	}
	return "", fmt.Errorf("Name not found")
}

func (a *App) getLowerName(movie Movie) (string, error) {
	var m Movie
	err := a.db.Model(&m).Where("upper(name) = ? and year = ? and name != ?", strings.ToUpper(movie.Name), movie.Year, strings.ToUpper(movie.Name)).First()
	return m.Name, err
}

func (a *App) getNoRating() ([]Movie, error) {
	var movies []Movie
	err := a.db.Model(&movies).Where("kinopoisk = 0 OR imdb = 0").Select()
	return movies, err
}

func (a *App) getRating(movie Movie) (kpp.KP, error) {
	var kp kpp.KP
	kp, err := kpp.GetRating(movie.Name, movie.EngName, movie.Year)
	if err != nil {
		return kp, fmt.Errorf("Rating no found")
	}
	if kp.Kinopoisk == 0 && kp.IMDb == 0 {
		return kp, fmt.Errorf("Rating no found")
	}
	return kp, nil
}

func (a *App) getFilmByMovieID(id int64) (Torrent, error) {
	var torrent Torrent
	err := a.db.Model(&torrent).Where("movie_id = ?", id).First()
	return torrent, err
}
