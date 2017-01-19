package main

import (
	"fmt"
	"log"
	"time"

	"github.com/serbe/kpp"
	"github.com/serbe/ncp"

	"gopkg.in/pg.v5"
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
	ID          int64     `sql:"id"`
	Section     string    `sql:"section"`
	Name        string    `sql:"name"`
	EngName     string    `sql:"eng_name"`
	Year        int       `sql:"year"`
	Genre       []string  `sql:"genre"        pg:",array"`
	Country     []string  `sql:"country"      pg:",array"`
	RawCountry  string    `sql:"raw_country"`
	Director    []string  `sql:"director"     pg:",array"`
	Producer    []string  `sql:"producer"     pg:",array"`
	Actor       []string  `sql:"actor"        pg:",array"`
	Description string    `sql:"description"`
	Age         string    `sql:"age"`
	ReleaseDate string    `sql:"release_date"`
	RussianDate string    `sql:"russian_date"`
	Duration    string    `sql:"duration"`
	Kinopoisk   float64   `sql:"kinopoisk"`
	IMDb        float64   `sql:"imdb"`
	Poster      string    `sql:"poster"`
	PosterURL   string    `sql:"poster_url"`
	CreatedAt   time.Time `sql:"created_at"`
	UpdatedAt   time.Time `sql:"updated_at"`
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
	ID            int64     `sql:"id"`
	MovieID       int64     `sql:"movie_id"`
	DateCreate    string    `sql:"date_create"`
	Href          string    `sql:"href"`
	Torrent       string    `sql:"torrent"`
	Magnet        string    `sql:"magnet"`
	NNM           float64   `sql:"nnm"`
	SubtitlesType string    `sql:"subtitles_type"`
	Subtitles     string    `sql:"subtitles"`
	Video         string    `sql:"video"`
	Quality       string    `sql:"quality"`
	Resolution    string    `sql:"resolution"`
	Audio1        string    `sql:"audio1"`
	Audio2        string    `sql:"audio2"`
	Audio3        string    `sql:"audio3"`
	Translation   string    `sql:"translation"`
	Size          int       `sql:"size"`
	Seeders       int       `sql:"seeders"`
	Leechers      int       `sql:"leechers"`
	CreatedAt     time.Time `sql:"created_at"`
	UpdatedAt     time.Time `sql:"updated_at"`
}

// App struct variables
type App struct {
	db      *pg.DB
	net     *ncp.NC
	hd      string
	px      string
	debug   bool
	debugDB bool
}

var app *App

func appInit() (*App, error) {
	if app == nil {
		app = new(App)
		conf, err := getConfig()
		if err != nil {
			log.Fatal("Error getConfig ", err)
		}
		options := &pg.Options{
			User:     conf.Db.User,
			Password: conf.Db.Password,
			Database: conf.Db.Name,
			// conf.Db.Sslmode,
		}
		db := pg.Connect(options)
		app.db = db
		err = app.createSchema()
		if err != nil {
			log.Fatal("Error createSchema ", err)
		}

		inetConnect, err := ncp.Init(conf.Nnm.Login, conf.Nnm.Password, conf.Address, conf.Proxy, conf.Debug)
		if err != nil {
			log.Println("net init ", err)
			return app, err
		}
		_ = createDir(conf.ImgDir)

		app.net = inetConnect
		app.hd = conf.ImgDir
		app.px = conf.Proxy
		app.debug = conf.Debug
		app.debugDB = conf.DebugDB
	}
	return app, nil
}

func (a *App) logDB(s string) {
	if a.debugDB {
		log.Println(s)
	}
}

func (a *App) createSchema() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS movies (
			id bigserial primary key,
			section text,
			name text,
			eng_name text,
			year int,
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
			poster_url text,
			created_at timestamp,
			updated_at timestamp
        )`,
		`CREATE TABLE IF NOT EXISTS torrents (
			id bigserial primary key,
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
			size int,
			seeders int,
			leechers int,
			created_at timestamp,
			updated_at timestamp
		)`,
	}
	for _, q := range queries {
		a.logDB(q)
		_, err := a.db.Exec(q)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *App) createMovie(ncf ncp.Film) (int64, error) {
	var m Movie
	m.Section = ncf.Section
	m.Name = ncf.Name
	m.EngName = ncf.EngName
	m.Year = ncf.Year
	m.Genre = ncf.Genre
	m.Country = ncf.Country
	m.RawCountry = ncf.RawCountry
	m.Director = ncf.Director
	m.Producer = ncf.Producer
	m.Actor = ncf.Actor
	m.Description = ncf.Description
	m.Age = ncf.Age
	m.ReleaseDate = ncf.ReleaseDate
	m.RussianDate = ncf.RussianDate
	m.Duration = ncf.Duration
	kp, err := a.getRating(m)
	if err == nil {
		m.Kinopoisk = kp.Kinopoisk
		m.IMDb = kp.IMDb
	}
	m.PosterURL = ncf.Poster
	m.Poster, _ = a.getPoster(m.PosterURL)
	err = a.db.Insert(&m)
	return m.ID, err
}

func (a *App) createTorrent(ncf ncp.Film) (int64, error) {
	var (
		t   Torrent
		err error
	)
	t.MovieID, err = a.getMovieID(ncf)
	if err != nil {
		movieID, err := a.createMovie(ncf)
		if err != nil {
			log.Println("createMovie ", err)
			return 0, err
		}
		t.MovieID = movieID
	}
	t.DateCreate = ncf.DateCreate
	t.Href = ncf.Href
	t.Torrent = ncf.Torrent
	t.Magnet = ncf.Magnet
	t.NNM = ncf.NNM
	t.SubtitlesType = ncf.SubtitlesType
	t.Subtitles = ncf.Subtitles
	t.Video = ncf.Video
	t.Quality = ncf.Quality
	t.Resolution = ncf.Resolution
	t.Audio1 = ncf.Audio1
	t.Audio2 = ncf.Audio2
	t.Audio3 = ncf.Audio3
	t.Translation = ncf.Translation
	t.Size = ncf.Size
	t.Seeders = ncf.Seeders
	t.Leechers = ncf.Leechers
	err = a.db.Insert(&t)
	return t.ID, err
}

func (a *App) getMovies() ([]Movie, error) {
	var movies []Movie
	err := a.db.Model(&movies).Select()
	return movies, err
}

func (a *App) getMovieID(ncf ncp.Film) (int64, error) {
	var id int64
	_, err := a.db.QueryOne(&id, `SELECT id FROM movies WHERE UPPER(name) = UPPER(?) AND year = ?`, ncf.Name, ncf.Year)
	return id, err
}

func (a *App) getTorrentByHref(href string) (Torrent, error) {
	var t Torrent
	_, err := a.db.QueryOne(&t, `SELECT * FROM torrents WHERE href = ?`, href)
	return t, err
}

func (a *App) updateTorrent(id int64, f ncp.Film) error {
	var t Torrent
	_, err := a.db.QueryOne(&t, `SELECT * FROM torrents WHERE id = ?`, id)
	if err != nil {
		return err
	}
	t.NNM = f.NNM
	t.Seeders = f.Seeders
	t.Leechers = f.Leechers
	t.Torrent = f.Torrent
	_, err = a.db.Exec(`
		UPDATE
			torrents
		SET
			nnm = ?,
			seeders = ?,
			leechers = ?,
			torrent = ?,
			updated_at = now()
		WHERE
			id = ?
	`, t.NNM, t.Seeders, t.Leechers, t.Torrent, id)
	return err
}

func (a *App) updateName(id int64, name string) error {
	var m Movie
	_, err := a.db.QueryOne(&m, `SELECT * FROM movies WHERE id = ?`, id)
	if err != nil {
		return err
	}
	m.Name = name
	_, err = a.db.Exec(`
		UPDATE
			movies
		SET
			name = ?,
			updated_at = now()
		WHERE
			id = ?
	`, m.Name, id)
	return err
}

func (a *App) updateRating(m Movie, kp kpp.KP) error {
	m.Kinopoisk = kp.Kinopoisk
	m.IMDb = kp.IMDb
	if m.Duration == "" {
		m.Duration = kp.Duration
	}
	_, err := a.db.Exec(`
		UPDATE
			movies
		SET
			kinopoisk = ?,
			imdb = ?,
			duration = ?,
			updated_at = now()
		WHERE
			id = ?
	`, m.Kinopoisk, m.IMDb, m.Duration, m.ID)
	return err
}

func (a *App) updatePoster(m Movie, poster string) error {
	m.Poster = poster
	_, err := a.db.Exec(`
		UPDATE
			movies
		SET
			poster = ?,
			updated_at = now()
		WHERE
			id = ?
	`, m.Poster, m.ID)
	return err
}

func (a *App) updatePosterURL(m Movie, posterURL string) error {
	m.PosterURL = posterURL
	_, err := a.db.Exec(`
		UPDATE
			movies
		SET
			poster_url = ?,
			updated_at = now()
		WHERE
			id = ?
	`, m.PosterURL, m.ID)
	return err
}

func (a *App) getWithDownload() ([]Torrent, error) {
	var torrents []Torrent
	_, err := a.db.Query(&torrents, `SELECT * FROM torrents WHERE magnet != ''`)
	return torrents, err
}

func (a *App) getMovieName(ncf ncp.Film) (string, error) {
	var movies []Movie
	_, err := a.db.Query(&movies, `SELECT * FROM movies WHERE UPPER(name) = UPPER(?) and year = ?"`, ncf.Name, ncf.Year)
	if err != nil {
		return "", err
	}
	if len(movies) > 0 {
		return movies[0].Name, nil
	}
	return "", fmt.Errorf("Name not found")
}

func (a *App) getUpperName(m Movie) (string, error) {
	var s string
	_, err := a.db.QueryOne(&s, `SELECT name FROM movies WHERE UPPER(name) = UPPER(?) and year = ? and name != UPPER(?)`, m.Name, m.Year, m.Name)
	return s, err
}

func (a *App) getNoRatingMovies() ([]Movie, error) {
	var movies []Movie
	_, err := a.db.Query(&movies, `SELECT * FROM movies WHERE kinopoisk IS NULL OR imdb IS NULL`)
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

func (a *App) getTorrentByMovieID(id int64) (Torrent, error) {
	var t Torrent
	_, err := a.db.QueryOne(&t, `SELECT * FROM torrents WHERE movie_id = ?`, id)
	return t, err
}
