package ncp

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// NCp values:
// client http.Client with cookie
type NCp struct {
	client      http.Client
	baseAddress string
}

// Topic from forum
type Topic struct {
	Href    string
	Name    string
	Year    string
	Quality string
	Body    []byte
}

// Film all values
// Section       Раздел форума
// Name          Название
// EngName       Английское название
// Href          Ссылка
// Year          Год
// Genre         Жанр
// Country       Производство
// RawCountry    Производство (оригинальная строка)
// Director      Режиссер
// Producer      Продюсер
// Actors        Актеры
// Description   Описание
// Age           Возраст
// ReleaseDate   Дата мировой премьеры
// RussianDate   Дата премьеры в России
// Duration      Продолжительность
// Quality       Качество видео
// Translation   Перевод
// SubtitlesType Вид субтитров
// Subtitles     Субтитры
// Video         Видео
// Resolution    Разрешение видео
// Audio1        Аудио1
// Audio2        Аудио2
// Audio3        Аудио3
// Kinopoisk     Рейтинг кинопоиска
// Imdb          Рейтинг IMDb
// NNM           Рейтинг nnm-club
// Sound         Звук
// Size          Размер
// DateCreate    Дата создания раздачи
// Torrent       Ссылка на torrent
// Magnet        Ссылка на magnet
// Poster        Ссылка на постер
// Seeders       Количество раздающих
// Leechers      Количество скачивающих
type Film struct {
	Section       string
	Name          string
	EngName       string
	Href          string
	Year          int
	Genre         []string
	Country       []string
	RawCountry    string
	Director      []string
	Producer      []string
	Actor         []string
	Description   string
	Age           string
	ReleaseDate   string
	RussianDate   string
	Duration      string
	Quality       string
	Translation   string
	SubtitlesType string
	Subtitles     string
	Video         string
	Resolution    string
	Audio1        string
	Audio2        string
	Audio3        string
	Kinopoisk     float64
	IMDb          float64
	NNM           float64
	Size          int
	DateCreate    string
	Torrent       string
	Magnet        string
	Poster        string
	Seeders       int
	Leechers      int
}

// Init nnmc with login password
func Init(login string, password string, baseAddress string, proxy string) (*NCp, error) {
	var client http.Client
	if proxy != "" {
		os.Setenv("HTTP_PROXY", proxy)
	}
	cookieJar, _ := cookiejar.New(nil)
	client.Jar = cookieJar
	urlPost := baseAddress + "/forum/login.php"
	form := url.Values{}
	form.Set("username", login)
	form.Add("password", password)
	form.Add("redirect", "")
	form.Add("login", "âõîä")
	req, _ := http.NewRequest("POST", urlPost, bytes.NewBufferString(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(form.Encode())))
	_, err := client.Do(req)
	return &NCp{client: client, baseAddress: baseAddress}, err
}

// getHTML get body from href
func getHTML(href string, n *NCp, debug bool) ([]byte, error) {
	resp, err := n.client.Get(href)
	if err != nil {
		log.Println("client Get error:", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("ioutil.ReadAll error:", err)
	}
	buffer := bytes.NewBufferString("")
	for _, char := range body {
		var ch = toUtf(char)
		fmt.Fprintf(buffer, "%c", ch)
	}
	doc := buffer.Bytes()
	doc = replaceAll(doc, "&nbsp;", " ")
	doc = replaceAll(doc, "&amp;", "&")
	doc = replaceAll(doc, "&quot;", "'")
	doc = replaceAll(doc, "<br />", " ")
	doc = replaceAll(doc, "</span>:", ":</span>")
	doc = replaceAll(doc, "  ", " ")
	doc = removeTag(doc, `<span style="text-decoration:.+?">(.+?)</span>`)
	doc = removeTag(doc, `<span style="color:.+?">(.+?)</span>`)

	if debug == true {
		u, err := url.Parse(href)
		if err == nil {
			q := u.Query()
			t := q.Get("t")
			f := q.Get("f")
			if t != "" {
				ioutil.WriteFile(t+".html", doc, 0600)
			} else {
				if f != "" {
					ioutil.WriteFile(f+".html", doc, 0600)
				} else {
					ioutil.WriteFile(u.Path+".html", doc, 0600)
				}
			}
		}
	}
	return doc, nil
}

// ParseForumTree get topics from forumTree
func (n *NCp) ParseForumTree(href string, debug bool) ([]Topic, error) {
	var (
		topics []Topic
		reTree = regexp.MustCompile(`<a href="viewtopic.php\?t=(\d+).*?"class="topictitle">(.+?)\s\((\d{4})\)\s(.+?)</a>`)
	)
	body, err := getHTML(n.baseAddress+href, n, debug)
	if err != nil {
		return topics, err
	}
	if reTree.Match(body) == false {
		return topics, fmt.Errorf("No topic in body")
	}
	findResult := reTree.FindAllSubmatch(body, -1)
	for _, v := range findResult {
		var t Topic
		t.Href = string(v[1])
		t.Name = string(v[2])
		t.Year = string(v[3])
		t.Quality = string(v[4])
		topics = append(topics, t)
	}
	return topics, nil
}

// ParseTopic get film from topic
func (n *NCp) ParseTopic(topic Topic, debug bool) (Film, error) {
	var (
		film Film
	)
	name := strings.Split(topic.Name, " / ")
	switch len(name) {
	case 1:
		film.Name = strings.Trim(name[0], " ")
	case 2:
		film.Name = strings.Trim(name[0], " ")
		film.EngName = strings.Trim(name[1], " ")
	case 3:
		film.Name = strings.Trim(name[0], " ")
		film.EngName = strings.Trim(name[1], " ")
	case 4:
		film.Name = strings.Trim(name[0], " ")
		film.EngName = strings.Trim(name[2], " ")
	}
	film.Name = strings.Replace(film.Name, ":", ": ", -1)
	film.Name = strings.Replace(film.Name, "  ", " ", -1)
	film.EngName = strings.Replace(film.EngName, ":", ": ", -1)
	film.EngName = strings.Replace(film.EngName, "  ", " ", -1)
	film.Href = topic.Href
	if year, err := strconv.Atoi(topic.Year); err == nil {
		film.Year = year
	}
	body, err := getHTML(n.baseAddress+"/forum/viewtopic.php?t="+film.Href, n, debug)
	if err != nil {
		return film, err
	}
	topic.Body = body
	film.Section = topic.getSection()
	film.Country, film.RawCountry = topic.getCountry()
	film.Genre = topic.getGenre()
	film.Director = topic.getDirector()
	film.Producer = topic.getProducer()
	film.Actor = topic.getActor()
	film.Description = topic.getDescription()
	film.Age = topic.getAge()
	film.ReleaseDate = topic.getReleaseDate()
	film.RussianDate = topic.getRussianDate()
	film.Duration = topic.getDuration()
	film.Quality = topic.getQuality()
	film.Translation = topic.getTranslation()
	film.SubtitlesType = topic.getSubtitlesType()
	film.Subtitles = topic.getSubtitles()
	film.Video = topic.getVideo()
	film.Resolution = getResolution(film.Video)
	film.Audio1 = topic.getAudio1()
	film.Audio2 = topic.getAudio2()
	film.Audio3 = topic.getAudio3()
	film.Torrent = topic.getTorrent()
	film.Magnet = topic.getMagnet()
	film.DateCreate = topic.getDate()
	film.Size = topic.getSize()
	film.NNM = topic.getRating()
	film.Poster = topic.getPoster()
	film.Seeders = topic.getSeeds()
	film.Leechers = topic.getLeechs()
	return film, nil
}

func replaceAll(body []byte, from string, to string) []byte {
	var reStr = regexp.MustCompile(from)
	if reStr.Match(body) == false {
		return body
	}
	return reStr.ReplaceAll(body, []byte(to))
}

func replaceDate(s string) string {
	s = strings.Replace(s, " Янв ", ".01.", -1)
	s = strings.Replace(s, " января ", ".01.", -1)
	s = strings.Replace(s, " Фев ", ".02.", -1)
	s = strings.Replace(s, " февраля ", ".02.", -1)
	s = strings.Replace(s, " Мар ", ".03.", -1)
	s = strings.Replace(s, " марта ", ".03.", -1)
	s = strings.Replace(s, " Апр ", ".04.", -1)
	s = strings.Replace(s, " апреля ", ".04.", -1)
	s = strings.Replace(s, " Май ", ".05.", -1)
	s = strings.Replace(s, " мая ", ".05.", -1)
	s = strings.Replace(s, " Июн ", ".06.", -1)
	s = strings.Replace(s, " июня ", ".06.", -1)
	s = strings.Replace(s, " Июл ", ".07.", -1)
	s = strings.Replace(s, " июля ", ".07.", -1)
	s = strings.Replace(s, " Авг ", ".08.", -1)
	s = strings.Replace(s, " августа ", ".08.", -1)
	s = strings.Replace(s, " Сен ", ".09.", -1)
	s = strings.Replace(s, " сентября ", ".09.", -1)
	s = strings.Replace(s, " Окт ", ".10.", -1)
	s = strings.Replace(s, " октября ", ".10.", -1)
	s = strings.Replace(s, " Ноя ", ".11.", -1)
	s = strings.Replace(s, " ноября ", ".11.", -1)
	s = strings.Replace(s, " Дек ", ".12.", -1)
	s = strings.Replace(s, " декабря ", ".12.", -1)
	split := strings.Split(s, ".")
	if len(split[0]) == 1 {
		s = "0" + s
	}
	return s
}

func caseInsensitiveContains(s, substr string) bool {
	s, substr = strings.ToUpper(s), strings.ToUpper(substr)
	return strings.Contains(s, substr)
}

func cleanStr(str string) string {
	var reSpan = regexp.MustCompile("<span .*?>")
	str = reSpan.ReplaceAllString(str, "")
	str = strings.Trim(str, " ")
	return str
}

func removeTag(body []byte, tag string) []byte {
	var reTag = regexp.MustCompile(tag)
	if reTag.Match(body) == false {
		return body
	}
	tags := reTag.FindAllSubmatch(body, -1)
	for _, item := range tags {
		body = bytes.Replace(body, item[0], item[1], 1)
	}
	return body
}

func stringToStruct(in string) []string {
	var out []string
	itemsArray := strings.Split(in, ",")
	for _, item := range itemsArray {
		trimItem := strings.Trim(item, " ")
		if item != "" {
			out = append(out, trimItem)
		}
	}
	return out
}
