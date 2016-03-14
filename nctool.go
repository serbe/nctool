package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	args := os.Args
	if contain(args, "help") {
		fmt.Println(`Usage:
	nctool COMMAND

Commands:
	help    показать справку
	get     получить новые фильмы
	update  обновление информации фильмов
	name    поиск и исправление имен фильмов
	rating  получение рейтинга Кинопоиска и IMDb
	poster  получение постеров`)
		os.Exit(0)
	}
	if containCommand(args) == false {
		fmt.Println(`comand not found: use "nctool help"`)
		os.Exit(1)
	}
	app, err := appInit()
	if err != nil {
		os.Exit(1)
	}
	if contain(args, "get") {
		log.Println("Start getting new films")
		err := app.get()
		log.Println("End getting new films")
		exit(err)
	}
	if contain(args, "update") {
		log.Println("Start update topics")
		err := app.update()
		log.Println("End update topics")
		exit(err)
	}
	if contain(args, "name") {
		log.Println("Start fix names")
		err := app.name()
		log.Println("End fix names")
		exit(err)
	}
	if contain(args, "rating") {
		log.Println("Start get ratings")
		err := app.rating()
		log.Println("End get ratings")
		exit(err)
	}
	if contain(args, "poster") {
		log.Println("Start get posters")
		err := app.poster()
		log.Println("End get posters")
		exit(err)
	}
}
