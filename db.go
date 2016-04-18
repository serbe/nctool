package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/serbe/kpp"
	"github.com/serbe/ncp"
	// pq need to sqlx
	_ "github.com/lib/pq"

	sq "github.com/Masterminds/squirrel"
)

// App struct variables
type App struct {
	db    *sqlx.DB
	psql  sq.StatementBuilderType
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
			log.Fatal(err)
		}
		dbConnect, err := sqlx.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s", conf.Pq.User, conf.Pq.Password, conf.Pq.Dbname, conf.Pq.Sslmode))
		if err != nil {
			log.Fatal("db open ", err)
			return app, err
		}
		err = dbConnect.Ping()
		if err != nil {
			log.Fatal("db ping ", err)
			return app, err
		}
		inetConnect, err := ncp.Init(conf.Nnm.Login, conf.Nnm.Password, conf.Address, conf.Px)
		if err != nil {
			log.Println("net init ", err)
			return app, err
		}
		_ = createDir(conf.Hd)
		psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
		app.psql = psql
		app.db = dbConnect
		app.net = inetConnect
		app.hd = conf.Hd
		app.px = conf.Px
		app.debug = conf.Debug
	}
	return app, nil
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
	sql := app.psql.Insert("movies").Columns("section", "name",
		eng_name text,
		year int8,
		genre text,
		raw_country text,
		description text,
		age text,
		release_date text,
		russian_date text,
		duration text,
		kinopoisk float8,
		imdb float8,
		poster text,
		poster_url text)
	err = a.db.Model(Movie{}).Create(&movie).Error
	genres := ncf.Genre
	countries := ncf.Country
	directors := ncf.Director
	producer := ncf.Producer
	actors := ncf.Actor

	fmt.Println(err)
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
	err = a.db.Model(Torrent{}).Create(&tor).Error
	fmt.Println(err)
	return err
}

func (a *App) getMovies() ([]Movie, error) {
	var movies []Movie
	err := a.db.Model(Movie{}).Find(&movies).Error
	return movies, err
}

func (a *App) getMovieID(ncf ncp.Film) (int64, error) {
	var movie Movie
	err := a.db.Model(Movie{}).Where("name = ? AND year = ?", ncf.Name, ncf.Year).First(&movie).Error
	return movie.ID, err
}

func (a *App) getTorrentByHref(href string) (Torrent, error) {
	var tor Torrent
	err := a.db.Model(Torrent{}).Where("href = ?", href).First(&tor).Error
	return tor, err
}

func (a *App) updateTorrent(id int64, f ncp.Film) error {
	return a.db.Model(Torrent{}).Where("id = ?", id).Updates(Torrent{NNM: f.NNM, Seeders: f.Seeders, Leechers: f.Leechers, Torrent: f.Torrent}).Error
}

func (a *App) updateName(id int64, name string) error {
	return a.db.Model(Movie{}).Where("id = ?", id).Updates(Movie{Name: name}).Error
}

func (a *App) updateRating(movie Movie, kp kpp.KP) error {
	var duration string
	if movie.Duration == "" {
		duration = kp.Duration
	}
	return a.db.Model(Movie{}).Where("upper(name) = ? and year = ?", strings.ToUpper(movie.Name), movie.Year).Updates(Movie{Kinopoisk: kp.Kinopoisk, IMDb: kp.IMDb, Duration: duration}).Error
}

func (a *App) updatePoster(movie Movie, poster string) error {
	return a.db.Model(&movie).Where("id = ?", movie.ID).Update("poster", poster).Error
}

func (a *App) updatePosterURL(movie Movie, poster string) error {
	return a.db.Model(&movie).Where("id = ?", movie.ID).Update("poster_url", poster).Error
}

func (a *App) getWithDownload() ([]Torrent, error) {
	var (
		torrents []Torrent
	)
	err := a.db.Model(Torrent{}).Where("magnet != ''").Find(&torrents).Error
	return torrents, err
}

func (a *App) getMovieName(ncf ncp.Film) (string, error) {
	var movies []Movie
	a.db.Model(Movie{}).Where("upper(name) = ? and year = ?", strings.ToUpper(ncf.Name), ncf.Year).Find(&movies)
	if len(movies) > 0 {
		return movies[0].Name, nil
	}
	return "", fmt.Errorf("Name not found")
}

func (a *App) getLowerName(movie Movie) (string, error) {
	var m Movie
	err := a.db.Model(Movie{}).Where("upper(name) = ? and year = ? and name != ?", strings.ToUpper(movie.Name), movie.Year, strings.ToUpper(movie.Name)).First(&m).Error
	return m.Name, err
}

func (a *App) getNoRating() ([]Movie, error) {
	var movies []Movie
	err := a.db.Model(Movie{}).Where("kinopoisk = 0 OR imdb = 0").Find(&movies).Error
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
	err := a.db.Model(Torrent{}).Where("movie_id = ?", id).First(&torrent).Error
	return torrent, err
}
