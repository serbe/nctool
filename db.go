package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/serbe/kpp"
	"github.com/serbe/ncp"
	// pq need to gorm
	_ "github.com/lib/pq"
)

// Film all values
// ID            id
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
// Poster        Ссылка на постер
// UpdatedAt     Дата обновления записи БД
// CreatedAt     Дата создания записи БД
type Film struct {
	ID          int64     `gorm:"column:id"             db:"id"             sql:"AUTO_INCREMENT"`
	Name        string    `gorm:"column:name"           db:"name"           sql:"type:text"`
	EngName     string    `gorm:"column:eng_name"       db:"eng_name"       sql:"type:text"`
	Year        int64     `gorm:"column:year"           db:"year"`
	Genre       string    `gorm:"column:genre"          db:"genre"          sql:"type:text"`
	Country     string    `gorm:"column:country"        db:"country"        sql:"type:text"`
	Director    string    `gorm:"column:director"       db:"director"       sql:"type:text"`
	Producer    string    `gorm:"column:producer"       db:"producer"       sql:"type:text"`
	Actors      string    `gorm:"column:actors"         db:"actors"         sql:"type:text"`
	Description string    `gorm:"column:description"    db:"description"    sql:"type:text"`
	Age         string    `gorm:"column:age"            db:"age"            sql:"type:text"`
	ReleaseDate string    `gorm:"column:release_date"   db:"release_date"   sql:"type:text"`
	RussianDate string    `gorm:"column:russian_date"   db:"russian_date"   sql:"type:text"`
	Duration    string    `gorm:"column:duration"       db:"duration"       sql:"type:text"`
	Kinopoisk   float64   `gorm:"column:kinopoisk"      db:"kinopoisk"`
	IMDb        float64   `gorm:"column:imdb"           db:"imdb"`
	Poster      string    `gorm:"column:poster"         db:"poster"         sql:"type:text"`
	UpdatedAt   time.Time `gorm:"column:updated_at"     db:"updated_at"`
	CreatedAt   time.Time `gorm:"column:created_at"     db:"created_at"`
}

// Torrent all values
// ID            id
// FilmID        Указатель на фильм
// DateCreate    Дата создания раздачи
// Href          Ссылка
// Torrent       Ссылка на torrent
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
// UpdatedAt     Дата обновления записи БД
// CreatedAt     Дата создания записи БД
type Torrent struct {
	ID            int64     `gorm:"column:id"             db:"id"             sql:"AUTO_INCREMENT"`
	FilmID        int64     `gorm:"column:film_id"        db:"film_id"`
	DateCreate    string    `gorm:"column:date_create"    db:"date_create"    sql:"type:text"`
	Href          string    `gorm:"column:href"           db:"href"           sql:"type:text"`
	Torrent       string    `gorm:"column:torrent"        db:"torrent"        sql:"type:text"`
	NNM           float64   `gorm:"column:nnm"            db:"nnm"`
	SubtitlesType string    `gorm:"column:subtitles_type" db:"subtitles_type" sql:"type:text"`
	Subtitles     string    `gorm:"column:subtitles"      db:"subtitles"      sql:"type:text"`
	Video         string    `gorm:"column:video"          db:"video"          sql:"type:text"`
	Quality       string    `gorm:"column:quality"        db:"quality"        sql:"type:text"`
	Resolution    string    `gorm:"column:resolution"     db:"resolution"     sql:"type:text"`
	Audio1        string    `gorm:"column:audio1"         db:"audio1"         sql:"type:text"`
	Audio2        string    `gorm:"column:audio2"         db:"audio2"         sql:"type:text"`
	Audio3        string    `gorm:"column:audio3"         db:"audio3"         sql:"type:text"`
	Translation   string    `gorm:"column:translation"    db:"translation"    sql:"type:text"`
	Size          int64     `gorm:"column:size"           db:"size"`
	Seeders       int64     `gorm:"column:seeders"        db:"seeders"`
	Leechers      int64     `gorm:"column:leechers"       db:"leechers"`
	UpdatedAt     time.Time `gorm:"column:updated_at"     db:"updated_at"`
	CreatedAt     time.Time `gorm:"column:created_at"     db:"created_at"`
}

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
	dbConnect.AutoMigrate(Film{})
	dbConnect.AutoMigrate(Torrent{})
	// dbConnect.LogMode(true)
	inetConnect, err := ncp.Init(conf.Nnm.Login, conf.Nnm.Password)
	if err != nil {
		log.Println("net init ", err)
		return &App{}, err
	}
	return &App{db: dbConnect, net: inetConnect}, nil
}

func (a *App) createFilm(ncf ncp.Film) (int64, error) {
	var (
		film Film
		kp   kpp.KP
	)
	film.Name = ncf.Name
	film.EngName = ncf.EngName
	film.Year = ncf.Year
	film.Genre = ncf.Genre
	film.Country = ncf.Country
	film.Director = ncf.Director
	film.Producer = ncf.Producer
	film.Actors = ncf.Actors
	film.Description = ncf.Description
	film.Age = ncf.Age
	film.ReleaseDate = ncf.ReleaseDate
	film.RussianDate = ncf.RussianDate
	film.Duration = ncf.Duration
	kp, err := a.getRating(film)
	if err == nil {
		film.Kinopoisk = kp.Kinopoisk
		film.IMDb = kp.IMDb
	}
	film.Poster = ncf.Poster
	err = a.db.Model(Film{}).Create(&film).Error
	return film.ID, err
}

func (a *App) createTorrent(ncf ncp.Film) error {
	var (
		tor Torrent
		err error
	)
	tor.FilmID, err = a.getFilmID(ncf)
	if err != nil {
		id, err := a.createFilm(ncf)
		if err != nil {
			return err
		}
		tor.FilmID = id
	}
	tor.DateCreate = ncf.DateCreate
	tor.Href = ncf.Href
	tor.Torrent = ncf.Torrent
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
	return a.db.Model(Torrent{}).Create(&tor).Error
}

func (a *App) getFilms() ([]Film, error) {
	var films []Film
	err := a.db.Model(Film{}).Find(&films).Error
	return films, err
}

func (a *App) getFilmID(ncf ncp.Film) (int64, error) {
	var film Film
	err := a.db.Model(Film{}).Where("name = ? AND year = ?", ncf.Name, ncf.Year).First(&film).Error
	return film.ID, err
}

func (a *App) getTorrentByHref(href string) (Torrent, error) {
	var tor Torrent
	err := a.db.Model(Torrent{}).Where("href = ?", href).First(&tor).Error
	return tor, err
}

func (a *App) updateTorrent(id int64, f ncp.Film) error {
	return a.db.Model(Torrent{}).Where("id = ?", id).UpdateColumns(Torrent{NNM: f.NNM, Seeders: f.Seeders, Leechers: f.Leechers, Torrent: f.Torrent}).Error
}

func (a *App) updateName(id int64, name string) error {
	return a.db.Model(Film{}).Where("id = ?", id).UpdateColumn("name", name).Error
}

func (a *App) updateRating(film Film, kp kpp.KP) error {
	return a.db.Model(Film{}).Where("upper(name) = ? and year = ?", strings.ToUpper(film.Name), film.Year).UpdateColumns(Film{Kinopoisk: kp.Kinopoisk, IMDb: kp.IMDb}).Error
}

func (a *App) getWithDownload() ([]Torrent, error) {
	var (
		torrents []Torrent
	)
	err := a.db.Model(Torrent{}).Where("torrent != ''").Find(&torrents).Error
	return torrents, err
}

func (a *App) getFilmName(ncf ncp.Film) (string, error) {
	var films []Film
	a.db.Model(Film{}).Where("upper(name) = ? and year = ?", strings.ToUpper(ncf.Name), ncf.Year).Find(&films)
	if len(films) > 0 {
		return films[0].Name, nil
	}
	return "", fmt.Errorf("Name not found")
}

func (a *App) getLowerName(film Film) (string, error) {
	var f Film
	err := a.db.Model(Film{}).Where("upper(name) = ? and year = ? and name != ?", strings.ToUpper(film.Name), film.Year, strings.ToUpper(film.Name)).First(&f).Error
	return f.Name, err
}

func (a *App) getNoRating() ([]Film, error) {
	var films []Film
	err := a.db.Model(Film{}).Where("kinopoisk = 0 OR imdb = 0").Find(&films).Error
	return films, err
}

func (a *App) getRating(film Film) (kpp.KP, error) {
	var kp kpp.KP
	kp, err := kpp.GetRating(film.Name, film.EngName, film.Year)
	if err != nil {
		return kp, fmt.Errorf("Rating no found")
	}
	if kp.Kinopoisk == 0 && kp.IMDb == 0 {
		return kp, fmt.Errorf("Rating no found")
	}
	return kp, nil
}
