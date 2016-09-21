package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nfnt/resize"
	"github.com/serbe/ncp"
)

type config struct {
	Nnm struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	} `json:"nnmclub"`
	PqCfg struct {
		User     string `json:"user"`
		Password string `json:"password"`
		Dbname   string `json:"dbname"`
		Sslmode  bool   `json:"sslmode"`
	} `json:"postgresql"`
	Address string `json:"address"`
	HhhpDir string `json:"httpdir"`
	Proxy   string `json:"proxy"`
	Debug   bool   `json:"debug"`
	DebugDB bool   `json:"debugdb"`
}

func getConfig() (config, error) {
	c := config{}
	file, err := ioutil.ReadFile("./config.json")
	if err != nil {
		return c, err
	}
	err = json.Unmarshal(file, &c)
	return c, err
}

func contain(args []string, str string) bool {
	result := false
	for _, item := range args {
		if item == str {
			result = true
			return result
		}
	}
	return result
}

func containCommand(args []string) bool {
	result := false
	for _, item := range commands {
		if contain(args, item) {
			result = true
			return result
		}
	}
	return result
}

func exit(err error) {
	if err == nil {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}

func (a *App) checkName(ncf ncp.Film) ncp.Film {
	if ncf.Name != strings.ToUpper(ncf.Name) {
		return ncf
	}
	name, err := a.getMovieName(ncf)
	if err == nil {
		ncf.Name = name
		return ncf
	}
	return ncf
}

func (a *App) getPoster(url string) (string, error) {
	var (
		img    image.Image
		poster string
	)
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(url)
	if err != nil {
		return poster, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return poster, err
	}
	ext := url[len(url)-3:]
	if ext == "jpg" {
		img, err = jpeg.Decode(bytes.NewReader(body))
		if err != nil {
			return poster, err
		}
	} else if ext == "png" {
		img, err = png.Decode(bytes.NewReader(body))
		if err != nil {
			return poster, err
		}
	} else {
		return poster, fmt.Errorf("Not supportet extension")
	}
	m := resize.Resize(150, 0, img, resize.Lanczos3)
	outName := strings.Replace(url, "/", "", -1)
	outName = strings.Replace(outName, ":", "", -1)
	if len(outName) < 20 {
		outName = outName[:len(outName)-4]
	} else {
		outName = outName[len(outName)-20 : len(outName)-4]
	}
	poster = outName + ".jpg"
	out, err := os.Create(a.hd + poster)
	if err != nil {
		return poster, err
	}
	defer out.Close()
	jpeg.Encode(out, m, nil)
	return poster, nil
}

func existsFile(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func findStringInSlice(list []string, s string) int {
	for i, b := range list {
		if b == s {
			return i
		}
	}
	return -1
}

func deleteFromSlice(list []string, s string) []string {
	sis := findStringInSlice(list, s)
	list = append(list[:sis], list[sis+1:]...)
	return list
}

func createDir(path string) error {
	if existsFile(path) == true {
		return nil
	}
	return os.Mkdir(path, 0777)
}
