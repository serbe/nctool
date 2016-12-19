package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/serbe/kpp"
	"github.com/serbe/ncp"

	"github.com/lib/pq"
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
	Genre       []string  `sql:"genre"`
	Country     []string  `sql:"country"`
	RawCountry  string    `sql:"raw_country"`
	Director    []string  `sql:"director"`
	Producer    []string  `sql:"producer"`
	Actor       []string  `sql:"actor"`
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
	db      *sql.DB
	net     *ncp.NCp
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
		options := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s",
			conf.Db.User,
			conf.Db.Password,
			conf.Db.Name,
			conf.Db.Sslmode,
		)
		db, err := sql.Open("postgres", options)
		if err != nil {
			log.Fatal(err)
		}

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
	var (
		m  Movie
		id int64
		kp kpp.KP
	)
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
	err = a.db.QueryRow(`
		INSERT INTO movies
			(section, name, eng_name, year, genre, country, raw_country, director, producer, actor, description, age, release_date, russian_date, duration, kinopoisk, imdb, poster, poster_url, created_at)
		VALUES
			($1,      $2,   $3,       $4,   $5,    $6,      $7,          $8,       $9,       $10,   $11,         $12, $13,          $14,          $15,      $16,       $17,  $18,    $19,        now())
		RETURNING
			id
	`,
		m.Section, m.Name, m.EngName, m.Year, pq.Array(m.Genre), pq.Array(m.Country), m.RawCountry, pq.Array(m.Director), pq.Array(m.Producer), pq.Array(m.Actor), m.Description, m.Age, m.ReleaseDate, m.RussianDate, m.Duration, m.Kinopoisk, m.IMDb, m.Poster, m.PosterURL,
	).Scan(&id)
	return id, err
}

func (a *App) createTorrent(ncf ncp.Film) (int64, error) {
	var (
		t   Torrent
		id  int64
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
	err = a.db.QueryRow(`
		INSERT INTO torrents
			(movie_id, date_create, href, torrent, magnet, nnm, subtitles_type, subtitles, video, quality, resolution, audio1, audio2, audio3, translation, size, seeders, leechers, created_at)
		VALUES
			($1,       $2,          $3,   $4,      $5,     $6,  $7,             $8,        $9,    $10,     $11,        $12,    $13,    $14,    $15,         $16,  $17,     $18,      now())
		RETURNING
			id
	`,
		t.MovieID, t.DateCreate, t.Href, t.Torrent, t.Magnet, t.NNM, t.SubtitlesType, t.Subtitles, t.Video, t.Quality, t.Resolution, t.Audio1, t.Audio2, t.Audio3, t.Translation, t.Size, t.Seeders, t.Leechers,
	).Scan(&id)
	return id, err
}

func (a *App) getMovies() ([]Movie, error) {
	rows, err := a.db.Query(`SELECT * FROM movies`)
	if err != nil {
		return nil, err
	}
	return scanMovies(rows)
}

func (a *App) getMovieID(ncf ncp.Film) (int64, error) {
	var id int64
	err := a.db.QueryRow(`SELECT id FROM movies WHERE UPPER(name) = UPPER($1) AND year = $2`, ncf.Name, ncf.Year).Scan(&id)
	return id, err
}

func (a *App) getTorrentByHref(href string) (Torrent, error) {
	row := a.db.QueryRow(`SELECT * FROM torrents WHERE href = $1`, href)
	return scanTorrent(row)
}

func (a *App) updateTorrent(id int64, f ncp.Film) error {
	row := a.db.QueryRow(`SELECT * FROM torrents WHERE id = $1`, id)
	t, err := scanTorrent(row)
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
			nnm = $2,
			seeders = $3,
			leechers = $4,
			torrent = $5,
			updated_at = now()
		WHERE
			id = $1
	`, id, t.NNM, t.Seeders, t.Leechers, t.Torrent)
	return err
}

func (a *App) updateName(id int64, name string) error {
	row := a.db.QueryRow(`SELECT * FROM movies WHERE id = $1`, id)
	m, err := scanMovie(row)
	if err != nil {
		return err
	}
	m.Name = name
	_, err = a.db.Exec(`
		UPDATE
			moviess
		SET
			name = $2,
			updated_at = now()
		WHERE
			id = $1
	`, id, m.Name)
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
			moviess
		SET
			kinopoisk = $2,
			imdb = $3,
			duration = $4,
			updated_at = now()
		WHERE
			id = $1
	`, m.ID, m.Kinopoisk, m.IMDb, m.Duration)
	return err
}

func (a *App) updatePoster(m Movie, poster string) error {
	m.Poster = poster
	_, err := a.db.Exec(`
		UPDATE
			moviess
		SET
			poster = $2,
			updated_at = now()
		WHERE
			id = $1
	`, m.ID, m.Poster)
	return err
}

func (a *App) updatePosterURL(m Movie, posterURL string) error {
	m.PosterURL = posterURL
	_, err := a.db.Exec(`
		UPDATE
			moviess
		SET
			poster_url = $2,
			updated_at = now()
		WHERE
			id = $1
	`, m.ID, m.PosterURL)
	return err
}

func (a *App) getWithDownload() ([]Torrent, error) {
	rows, err := a.db.Query(`SELECT * FROM torrents WHERE magnet != ''`)
	if err != nil {
		return nil, err
	}
	return scanTorrents(rows)
}

func (a *App) getMovieName(ncf ncp.Film) (string, error) {
	rows, err := a.db.Query(`SELECT * FROM movies WHERE UPPER(name) = UPPER($1) and year = $2"`, ncf.Name, ncf.Year)
	if err != nil {
		return "", err
	}
	movies, err := scanMovies(rows)
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
	err := a.db.QueryRow(`SELECT name FROM movies WHERE UPPER(name) = UPPER($1) and year = $2 and name != UPPER($1)`, m.Name, m.Year).Scan(&s)
	return s, err
}

func (a *App) getNoRating() ([]Movie, error) {
	rows, err := a.db.Query(`SELECT * FROM movies WHERE kinopoisk = 0 OR imdb = 0`)
	if err != nil {
		return nil, err
	}
	return scanMovies(rows)
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
	row := a.db.QueryRow(`SELECT * FROM torrents WHERE movie_id = $1`, id)
	return scanTorrent(row)
}
