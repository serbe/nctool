package main

import (
	"database/sql"

	// pq need to sqlx
	_ "github.com/lib/pq"
)

// Movie - таблица фильмов
// ID            id
// Section       Раздел форума
// Name          Название
// EngName       Английское название
// Year          Год
// Genre         Жанр
// Country       Производство
// RawCountry    Оригинал производство (для отладки)
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
	ID          int64  `db:"id"`
	Section     string `db:"section"`
	Name        string `db:"name"`
	EngName     string `db:"eng_name"`
	Year        int64  `db:"year"`
	Genre       string `db:"genre"`
	Country     []Country
	RawCountry  string `db:"raw_country"`
	Director    []Person
	Producer    []Person
	Actor       []Person
	Description string  `db:"description"`
	Age         string  `db:"age"`
	ReleaseDate string  `db:"release_date"`
	RussianDate string  `db:"russian_date"`
	Duration    string  `db:"duration"`
	Kinopoisk   float64 `db:"kinopoisk"`
	IMDb        float64 `db:"imdb"`
	Poster      string  `db:"poster"`
	PosterURL   string  `db:"poster_url"`
}

// Torrent - таблица торрентов
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
	ID            int64   `db:"id"`
	MovieID       int64   `db:"movie_id"`
	DateCreate    string  `db:"date_create"`
	Href          string  `db:"href"`
	Torrent       string  `db:"torrent"`
	Magnet        string  `db:"magnet"`
	NNM           float64 `db:"nnm"`
	SubtitlesType string  `db:"subtitles_type"`
	Subtitles     string  `db:"subtitles"`
	Video         string  `db:"video"`
	Quality       string  `db:"quality"`
	Resolution    string  `db:"resolution"`
	Audio1        string  `db:"audio1"`
	Audio2        string  `db:"audio2"`
	Audio3        string  `db:"audio3"`
	Translation   string  `db:"translation"`
	Size          int64   `db:"size"`
	Seeders       int64   `db:"seeders"`
	Leechers      int64   `db:"leechers"`
}

// Person - таблица людей (актеры, режиссеры, продюссеры)
// ID            id
// Name          ФИО
// EngName       Английское ФИО
// Year          Год рождения
// Country       Гражданство
// Director      Режиссер фильмов
// Producer      Продюсер фильмов
// Actors        Актер фильмов
// Description   Описание
// Poster        Имя файла постера
type Person struct {
	ID          int64  `db:"id"`
	Name        string `db:"name"`
	EngName     string `db:"eng_name"`
	Year        int64  `db:"year"`
	Genre       string `db:"genre"`
	Country     string `db:"country"`
	Director    []Movie
	Producer    []Movie
	Actor       []Movie
	Description string `db:"description"`
	Poster      string `db:"poster"`
}

// Country - таблица стран
// ID            id
// Name          Название страны
type Country struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

// MovieCountry - таблица сопоставления фильмов странам
// ID            id
// MovieID       id фильмя
// CountryID     id страны
type MovieCountry struct {
	ID        int64 `db:"id"`
	MovieID   int64 `db:"movie_id"`
	CountryID int64 `db:"country_id"`
}

// Director - таблица режиссеров фильмов
// ID            id
// MovieID       id фильмя
// PersonID      id человека
type Director struct {
	ID       int64 `db:"id"`
	MovieID  int64 `db:"movie_id"`
	PersonID int64 `db:"person_id"`
}

// Producer - таблица продюссеров фильмов
// ID            id
// MovieID       id фильмя
// PersonID      id человека
type Producer struct {
	ID       int64 `db:"id"`
	MovieID  int64 `db:"movie_id"`
	PersonID int64 `db:"person_id"`
}

// Actor - таблица актеров фильмов
// ID            id
// MovieID       id фильмя
// PersonID      id человека
type Actor struct {
	ID       int64 `db:"id"`
	MovieID  int64 `db:"movie_id"`
	PersonID int64 `db:"person_id"`
}

func (app *App) createModels() error {
	app.createMoviesModel()
	app.createTorrentsModel()
	app.createCountriesModel()
	app.createMoviesCountriesModel()
	app.createDirectorsModel()
	app.createProducersModel()
	app.createActorsModel()
}

func (app *App) createMoviesModel() {
	sql := `CREATE TABLE IF NOT EXISTS movies ( 
		id serial8,
		section text,
		name text,
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
		poster_url text
	)`
	app.db.MustExec(sql)
}

func (app *App) createTorrentsModel() {
	sql := `CREATE TABLE IF NOT EXISTS torrents (
		id serial8,
		movie_id int8,
		date_create text,
		href text,
		torrent text,
		magnet text,
		nnm float8,
		subtitles_type text,
		subtitles text,
		video text,
		quality text,
		resolution text,
		audio1 text,
		audio2 text,
		audio3 text,
		translation text,
		size int8,
		seeders int8,
		leechers int8
	)`
	app.db.MustExec(sql)
}

func (app *App) createCountriesModel() {
	sql := `CREATE TABLE IF NOT EXISTS countries (
		id serial8,
		name text
	)`
	app.db.MustExec(sql)
}

func (app *App) createMoviesCountriesModel() {
	sql := `CREATE TABLE IF NOT EXISTS movies_countries (
		id serial8,
		movie_id int8,
		country_id int8
	)`
	app.db.MustExec(sql)
}

func (app *App) createDirectorsModel() {
	sql := `CREATE TABLE IF NOT EXISTS directors (
		id serial8,
		movie_id int8,
		person_id int8
	)`
	app.db.MustExec(sql)
}

func (app *App) createProducersModel() {
	sql := `CREATE TABLE IF NOT EXISTS producers (
		id serial8,
		movie_id int8,
		person_id int8
	)`
	app.db.MustExec(sql)
}

func (app *App) createActorsModel() {
	sql := `CREATE TABLE IF NOT EXISTS actors (
		id serial8,
		movie_id int8,
		person_id int8
	)`
	app.db.MustExec(sql)
}
