package ncp

import (
	"regexp"
	"strconv"
	"strings"
)

func (t *Topic) getSection() string {
	var (
		reSection = regexp.MustCompile(`<a href="viewforum.php\?f=\d+?" class="nav">(.+?)</a>`)
		section   string
	)
	if reSection.Match(t.Body) == true {
		section = string(reSection.FindSubmatch(t.Body)[1])
		section = cleanStr(section)
	}
	return section
}

func (t *Topic) getRating() float64 {
	var (
		reRating = regexp.MustCompile(`>(\d,\d|\d)<\/span>.+?\(Голосов:`)
		rating   float64
	)
	if reRating.Match(t.Body) == true {
		str := string(reRating.FindSubmatch(t.Body)[1])
		str = strings.Replace(str, ",", ".", -1)
		rating, _ = strconv.ParseFloat(str, 64)
	}
	return rating
}

func (t *Topic) getSize() int {
	var (
		reSize = regexp.MustCompile(`Размер блока: \d.+?B"> (\d{1,2},\d{1,2}|\d{3,4}|\d{1,2})\s`)
		size   int
	)
	if reSize.Match(t.Body) == true {
		str := string(reSize.FindSubmatch(t.Body)[1])
		str = strings.Replace(str, ",", ".", -1)
		if s64, err := strconv.ParseFloat(str, 64); err == nil {
			if s64 < 100 {
				s64 = s64 * 1000
			}
			size = int(s64)
		}
	}
	return size
}

func (t *Topic) getTorrent() string {
	var (
		reTor   = regexp.MustCompile(`<a href="download\.php\?id=(\d{5,7})" rel="nofollow">Скачать<`)
		torrent string
	)
	if reTor.Match(t.Body) == true {
		findTor := reTor.FindSubmatch(t.Body)
		torrent = string(findTor[1])
		// http://nnm-club.me/forum/download.php?id=
	}
	return torrent
}

func (t *Topic) getMagnet() string {
	var (
		reMag  = regexp.MustCompile(`href="magnet:\?xt=urn:btih:(.+?)(?:"|&)`)
		magnet string
	)
	if reMag.Match(t.Body) == true {
		findMag := reMag.FindSubmatch(t.Body)
		magnet = string(findMag[1])
		// magnet:?xt=urn:btih:
	}
	return magnet
}

func (t *Topic) getPoster() string {
	var (
		rePos = regexp.MustCompile(`"postImg postImgAligned img-right" title="http:\/\/assets\..+?\/forum\/image\.php\?link=(.+?(?:jpg|jpeg|png))"`)
		image string
	)
	if rePos.Match(t.Body) == true {
		image = string(rePos.FindSubmatch(t.Body)[1])
	}
	return image
}

func (t *Topic) getDate() string {
	var (
		reDate = regexp.MustCompile(`> (\d{1,2} .{3} \d{4}).{9}<`)
		date   string
	)
	if reDate.Match(t.Body) == true {
		date = replaceDate(string(reDate.FindSubmatch(t.Body)[1]))
	}
	return date
}

func (t *Topic) getSeeds() int {
	var (
		reSs  = regexp.MustCompile(`<span class="seed">\[ <b>(\d{1,5})\s`)
		seeds int
	)
	if reSs.Match(t.Body) == true {
		ss := reSs.FindSubmatch(t.Body)
		seeds, _ = strconv.Atoi(string(ss[1]))
	}
	return seeds
}

func (t *Topic) getLeechs() int {
	var (
		reLs   = regexp.MustCompile(`<span class="leech">\[ <b>(\d{1,5})\s`)
		leechs int
	)
	if reLs.Match(t.Body) == true {
		ls := reLs.FindSubmatch(t.Body)
		leechs, _ = strconv.Atoi(string(ls[1]))
	}
	return leechs
}

func getResolution(str string) string {
	var (
		reRes      = regexp.MustCompile(`(\d{3,4}x\d{3,4}|\d{3,4}X\d{3,4}|\d{3,4}х\d{3,4}|\d{3,4}Х\d{3,4})`)
		resolution string
	)
	if reRes.MatchString(str) == true {
		resolution = reRes.FindString(str)
	}
	return resolution
}

func (t *Topic) getCountry() ([]string, string) {
	var (
		reCountry  = regexp.MustCompile(`<span style="font-weight: bold">Производство:\s*<\/span>(.+?)<`)
		rawCountry string
		country    []string
	)
	if reCountry.Match(t.Body) == true {
		rawCountry = string(reCountry.FindSubmatch(t.Body)[1])
		rawCountry = cleanStr(rawCountry)
	}
	lowerRawCountry := strings.ToLower(rawCountry)
	for _, item := range counriesList {
		i := strings.Index(lowerRawCountry, strings.ToLower(item))
		if i != -1 {
			country = append(country, item)
			lowerRawCountry = lowerRawCountry[:i] + lowerRawCountry[i+len(item):]
		}
	}
	return country, rawCountry
}

func (t *Topic) getGenre() []string {
	var (
		reGenre = regexp.MustCompile(`<span style="font-weight: bold">Жанр:\s*<\/span>(.+?)<`)
		genre   []string
	)
	if reGenre.Match(t.Body) == true {
		str := string(reGenre.FindSubmatch(t.Body)[1])
		str = strings.ToLower(cleanStr(str))
		str = strings.Trim(str, ".")
		genre = stringToStruct(str)
	}
	return genre
}

func (t *Topic) getDirector() []string {
	var (
		reDirector = regexp.MustCompile(`<span style="font-weight: bold">Режиссер:\s*<\/span>(.+?)<`)
		director   []string
	)
	if reDirector.Match(t.Body) == true {
		str := string(reDirector.FindSubmatch(t.Body)[1])
		str = cleanStr(str)
		str = strings.Trim(str, ".")
		director = stringToStruct(str)
	}
	return director
}

func (t *Topic) getProducer() []string {
	var (
		reProducer = regexp.MustCompile(`<span style="font-weight: bold">Продюсер:\s*<\/span>(.+?)<`)
		producer   []string
	)
	if reProducer.Match(t.Body) == true {
		str := string(reProducer.FindSubmatch(t.Body)[1])
		str = cleanStr(str)
		str = strings.Trim(str, ".")
		producer = stringToStruct(str)
	}
	return producer
}

func (t *Topic) getActor() []string {
	var (
		reActor = regexp.MustCompile(`<span style="font-weight: bold">Актеры:\s*<\/span>(.+?)<`)
		actor   []string
	)
	if reActor.Match(t.Body) == true {
		str := string(reActor.FindSubmatch(t.Body)[1])
		str = cleanStr(str)
		str = strings.Trim(str, ".")
		actor = stringToStruct(str)
	}
	return actor
}

func (t *Topic) getDescription() string {
	var (
		reDescription = regexp.MustCompile(`<span style="font-weight: bold">(?:Описание фильма|Описание):\s*<\/span>(.+?)<`)
		description   string
	)
	if reDescription.Match(t.Body) == true {
		description = string(reDescription.FindSubmatch(t.Body)[1])
		description = cleanStr(description)
	}
	return description
}

func (t *Topic) getAge() string {
	var (
		reAge = regexp.MustCompile(`<span style="font-weight: bold">Возраст:\s*<\/span>(.+?)<`)
		age   string
	)
	if reAge.Match(t.Body) == true {
		age = string(reAge.FindSubmatch(t.Body)[1])
		age = cleanStr(age)
	}
	return age
}

func (t *Topic) getReleaseDate() string {
	var (
		reReleaseDate = regexp.MustCompile(`<span style="font-weight: bold">Дата мировой премьеры:\s*<\/span>(.+?)<`)
		releaseDate   string
	)
	if reReleaseDate.Match(t.Body) == true {
		releaseDate = string(reReleaseDate.FindSubmatch(t.Body)[1])
		releaseDate = cleanStr(releaseDate)
		releaseDate = replaceDate(releaseDate)
	}
	return releaseDate
}

func (t *Topic) getRussianDate() string {
	var (
		reRussianDate = regexp.MustCompile(`<span style="font-weight: bold">(?:Дата премьеры в России|Дата Российской премьеры|Дата российской премьеры):\s*<\/span>(.+?)<`)
		russianDate   string
	)
	if reRussianDate.Match(t.Body) == true {
		russianDate = string(reRussianDate.FindSubmatch(t.Body)[1])
		russianDate = cleanStr(russianDate)
		russianDate = replaceDate(russianDate)
	}
	return russianDate
}

func (t *Topic) getDuration() string {
	var (
		reDuration = regexp.MustCompile(`<span style="font-weight: bold">Продолжительность:\s*<\/span>(.+?)<`)
		duration   string
	)
	if reDuration.Match(t.Body) == true {
		duration = string(reDuration.FindSubmatch(t.Body)[1])
		duration = cleanStr(duration)
	} else {
		reDuration = regexp.MustCompile(`\sПродолжительность\s+?&#58; (\d{1,2}) ч\. (\d{1,2}) м\.`)
		if reDuration.Match(t.Body) == true {
			submatch := reDuration.FindSubmatch(t.Body)
			hour := string(submatch[1])
			minute := string(submatch[2])
			if len(hour) == 1 {
				hour = "0" + hour
			}
			if len(minute) == 1 {
				minute = "0" + minute
			}
			duration = hour + ":" + minute + ":00"
		}
	}
	if len(duration) < 5 {
		duration = ""
	}
	return duration
}

func (t *Topic) getQuality() string {
	var (
		reQuality = regexp.MustCompile(`<span style="font-weight: bold">(?:Качество видео|Качество):\s*<\/span>(.+?)<`)
		quality   string
	)
	if reQuality.Match(t.Body) == true {
		quality = string(reQuality.FindSubmatch(t.Body)[1])
		quality = cleanStr(quality)
	}
	return quality
}

func (t *Topic) getTranslation() string {
	var (
		reTranslation = regexp.MustCompile(`<span style="font-weight: bold">Перевод:\s*<\/span>(.+?)<`)
		translation   string
	)
	translation = "Не требуется"
	if reTranslation.Match(t.Body) == true {
		translation = string(reTranslation.FindSubmatch(t.Body)[1])
		translation = cleanStr(translation)
	}
	return translation
}

func (t *Topic) getSubtitlesType() string {
	var (
		reSubtitlesType = regexp.MustCompile(`<span style="font-weight: bold">Вид субтитров:\s*<\/span>(.+?)<`)
		subtitlesType   string
	)
	if reSubtitlesType.Match(t.Body) == true {
		subtitlesType = string(reSubtitlesType.FindSubmatch(t.Body)[1])
		subtitlesType = cleanStr(subtitlesType)
	}
	return subtitlesType
}

func (t *Topic) getSubtitles() string {
	var (
		reSubtitles = regexp.MustCompile(`<span style="font-weight: bold">Субтитры:\s*<\/span>(.+?)<`)
		subtitles   string
	)
	if reSubtitles.Match(t.Body) == true {
		subtitles = string(reSubtitles.FindSubmatch(t.Body)[1])
		subtitles = cleanStr(subtitles)
	}
	return subtitles
}

func (t *Topic) getVideo() string {
	var (
		reVideo = regexp.MustCompile(`<span style="font-weight: bold">Видео:\s*<\/span>(.+?)<`)
		video   string
	)
	if reVideo.Match(t.Body) == true {
		video = string(reVideo.FindSubmatch(t.Body)[1])
		video = cleanStr(video)
	}
	return video
}

func (t *Topic) getAudio1() string {
	var (
		reAudio = regexp.MustCompile(`<span style="font-weight: bold">(?:Аудио\s?:\s*|Аудио\s?.?1.?:\s*)<\/span>(.+?)<`)
		audio   string
	)
	if reAudio.Match(t.Body) == true {
		audio = string(reAudio.FindSubmatch(t.Body)[1])
		audio = cleanStr(audio)
	}
	return audio
}

func (t *Topic) getAudio2() string {
	var (
		reAudio = regexp.MustCompile(`<span style="font-weight: bold">Аудио\s?.?2.?:\s*<\/span>(.+?)<`)
		audio   string
	)
	if reAudio.Match(t.Body) == true {
		audio = string(reAudio.FindSubmatch(t.Body)[1])
		audio = cleanStr(audio)
	}
	return audio
}

func (t *Topic) getAudio3() string {
	var (
		reAudio = regexp.MustCompile(`<span style="font-weight: bold">Аудио\s?.?3.?:\s*<\/span>(.+?)<`)
		audio   string
	)
	if reAudio.Match(t.Body) == true {
		audio = string(reAudio.FindSubmatch(t.Body)[1])
		audio = cleanStr(audio)
	}
	return audio
}
