package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
)

func main() {
	run()
}

func run() {
	r := gin.Default()
	r.POST("/playerSearch", func(c *gin.Context) {
		var username username
		// var mappedData []byte

		jsonData, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(jsonData))

		err = json.Unmarshal(jsonData, &username)
		if err != nil {
			log.Fatal(err)
		}

		name := strings.Replace(string(username.Username), " ", "+", -1)

		data := getPlayerURL(name)

		fmt.Println(data)

		c.JSON(http.StatusOK, gin.H{
			"players": data,
		})
	})

	r.POST("/playerLookup", func(c *gin.Context) {
		var playerProfile playerProfile

		jsonData, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(jsonData))

		err = json.Unmarshal(jsonData, &playerProfile)
		if err != nil {
			log.Fatal(err)
		}

		data := findPlayer("https://www.dotabuff.com/players/121249435")

		fmt.Println(data)

		c.JSON(http.StatusOK, gin.H{
			"player": data,
		})
	})

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

type player struct {
	Name          string `json:"name"`
	PlayerProfile string `json:"playerProfile"`
}

type playerProfile struct {
	RankName   string `json:"rankName"`
	RankNumber string `json:"rankNumber"`
}

type username struct {
	Username string `json:"username"`
}

type playerSearch struct {
	PlayerURL string `json:"playerURL"`
	Image     string `json:"image"`
	Player    string `json:"player"`
}

const playersURL = "https://www.dotabuff.com/players"

var playerSearchURL = "https://www.dotabuff.com/search?q=%s&commit=Search"

var playerURL = ""

func getPlayers() []player {
	req, err := http.NewRequest(http.MethodGet, playersURL, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36")
	req.Header.Add("Upgrade-Insecure-Requests", "1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close() // nolint: errcheck

	// data, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	return gotReposErrMsg(err)
	// }

	text := parsePlayersPage(resp.Body)

	var player []player

	err = json.Unmarshal(text, &player)
	if err != nil {
		log.Fatal(err)
	}

	return player
}

func getPlayerURL(username string) []playerSearch {
	var playerSearch []playerSearch

	username = strings.Replace(username, " ", "+", -1)

	url := fmt.Sprintf(playerSearchURL, username)

	fmt.Println(url)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36")
	req.Header.Add("Upgrade-Insecure-Requests", "1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close() // nolint: errcheck

	// data, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println(string(data))

	text := parseSearch(resp.Body)

	err = json.Unmarshal(text, &playerSearch)
	if err != nil {
		log.Fatal(err)
	}

	return playerSearch
}

func findPlayer(playerLink string) []playerProfile {
	var playerProfile []playerProfile

	req, err := http.NewRequest(http.MethodGet, playerLink, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36")
	req.Header.Add("Upgrade-Insecure-Requests", "1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close() // nolint: errcheck

	// data, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	return gotReposErrMsg(err)
	// }

	text := parsePlayerProfile(resp.Body)

	err = json.Unmarshal(text, &playerProfile)
	if err != nil {
		log.Fatal(err)
	}

	return playerProfile
}

func remove(slice []map[string]string, s int) []map[string]string {
	return append(slice[:s], slice[s+1:]...)
}

func parseSearch(text io.Reader) (byteData []byte) {
	var data []map[string]string
	var mappedData = map[string]string{}
	var playersURLs []string
	var players []string
	var images []string

	doc, err := goquery.NewDocumentFromReader(text)

	if err != nil {
		log.Fatal(err)
	}

	doc.Find("a.link-type-player").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		// class, _ := s.Attr("class")

		// fmt.Println("HREF!!!!" + href)

		if strings.Contains(href, "/players/") {
			fmt.Println(s.Text())
			playersURLs = append(playersURLs, "https://www.dotabuff.com"+href)
			players = append(players, s.Text())
		}
	})

	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("src")
		// class, _ := s.Attr("class")

		// fmt.Println("SRC!!!!" + src)

		if strings.Contains(src, "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/") {
			images = append(images, src)
		}
	})

	for i, value := range images {
		mappedData = map[string]string{"playerURL": playersURLs[i], "player": players[i], "image": value}
		data = append(data, mappedData)
	}

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf).Encode(data)
	if enc != nil {
		log.Fatal(err)
	}

	bs := buf.Bytes()

	return bs
}

func parsePlayerProfile(text io.Reader) (byteData []byte) {
	var data []map[string]string
	var rankName string
	var rankNumber string
	var mappedData = map[string]string{}

	doc, err := goquery.NewDocumentFromReader(text)

	if err != nil {
		log.Fatal(err)
	}

	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("src")
		class, _ := s.Attr("class")

		if strings.Contains(href, "https://riki.dotabuff.com/c/") && class == "rank-tier-base" && href != "" || class == "rank-tier-pip" {
			if class == "rank-tier-base" {
				rankName = href
			} else if class == "rank-tier-pip" {
				rankNumber = href
			}
		}

		if rankName != "" && rankNumber != "" {
			mappedData = map[string]string{"rankName": rankName, "rankNumber": rankNumber}
			data = append(data, mappedData)
		}
	})

	remove(data, 0)
	remove(data, 0)

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf).Encode(data)
	if enc != nil {
		log.Fatal(err)
	}

	bs := buf.Bytes()

	return bs
}

func parsePlayersPage(text io.Reader) (byteData []byte) {
	var data []map[string]string

	doc, err := goquery.NewDocumentFromReader(text)

	if err != nil {
		log.Fatal(err)
	}

	doc.Find("a.link-type-player").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")

		if strings.Contains(href, "players") {
			mappedData := map[string]string{"name": s.Text(), "playerProfile": href}
			data = append(data, mappedData)
		}
	})

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf).Encode(data)
	if enc != nil {
		log.Fatal(err)
	}

	bs := buf.Bytes()

	return bs
}
