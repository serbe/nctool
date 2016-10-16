package main

import (
	"flag"
	"log"
	"os"
)

func main() {
	getFilms := flag.Bool("get", false, "Получить новые фильмы")
	updateFilms := flag.Bool("update", false, "Обновление информации фильмов")
	fixNames := flag.Bool("name", false, "Поиск и исправление имен фильмов")
	getRating := flag.Bool("rating", false, "Получение рейтинга Кинопоиска и IMDb")
	getPosters := flag.Bool("poster", false, "Получение постеров")

	app, err := appInit()
	if err != nil {
		os.Exit(1)
	}
	if *getFilms {
		log.Println("Start getting new films")
		err := app.get()
		log.Println("End getting new films")
		exit(err)
	}
	if *updateFilms {
		log.Println("Start update topics")
		err := app.update()
		log.Println("End update topics")
		exit(err)
	}
	if *fixNames {
		log.Println("Start fix names")
		err := app.name()
		log.Println("End fix names")
		exit(err)
	}
	if *getRating {
		log.Println("Start get ratings")
		err := app.rating()
		log.Println("End get ratings")
		exit(err)
	}
	if *getPosters {
		log.Println("Start get posters")
		err := app.poster()
		log.Println("End get posters")
		exit(err)
	}
}
