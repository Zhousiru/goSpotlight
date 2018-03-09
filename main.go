package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/buger/jsonparser"
)

const (
	API_URL     = "https://arc.msn.com/v3/Delivery/Cache?&pid=279978&fmt=json&ctry=US&lc=en-US&pl=en-US"
	REQUEST_NUM = 50
)

func download(i int, url string, image_type string) {
	res, err := http.Get(url)
	statusCode := res.StatusCode
	if err != nil {
		log.Printf("Request %d: Error: %s", i, err)
		return
	}
	if statusCode != http.StatusOK {
		log.Printf("Request %d: Error: Image URL returned %d", i, statusCode)
		return
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("Request %d: Error: %s", i, err)
	}

	name := string([]byte(url)[(len(url) - 15):(len(url) - 9)])
	path := fmt.Sprintf("./images/%s/", image_type)
	if _, err := os.Stat(path); err != nil {
		err = os.MkdirAll(path, 0666)
		if err != nil {
			log.Printf("Request %d: Error: %s", i, err)
		}
	}
	fImage, err := os.Create(path + name + ".jpg")
	if err != nil {
		log.Printf("Request %d: Error: %s", i, err)
	}

	defer fImage.Close()

	_, err = fImage.Write(data)
	if err != nil {
		log.Printf("Request %d: Error: %s", i, err)
	}

	fImage.Sync()

	return
}

func main() {
	fmt.Println("[Spotlight Spider v0.0.1]")

	for i := 0; i < REQUEST_NUM; i++ {
		log.Printf("Request %d: Getting API response...", i)

		res, err := http.Get(API_URL)
		statusCode := res.StatusCode
		if err != nil {
			log.Fatalf("Request %d: Error: %s", i, err)
		}
		if statusCode != http.StatusOK {
			log.Fatalf("Request %d: Error: API returned %d", i, statusCode)
		}

		content, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatalf("Request %d: Error: %s", i, err)
		}

		json_itemorder, _, _, err := jsonparser.Get(content, "batchrsp", "itemorder")
		if err != nil {
			log.Fatalf("Request %d: Error: %s", i, err)
		}

		itemsNum, err := strconv.Atoi(string(json_itemorder[len(json_itemorder)-2]))
		if err != nil {
			log.Fatalf("Request %d: Error: %s", i, err)
		}

		itemsNum++

		log.Printf("Request %d: %d item(s) have been found, getting image URL(s) from item(s)...", i, itemsNum)

		for j := 0; j < itemsNum; j++ {
			itemRaw, err := jsonparser.GetString(content, "batchrsp", "items", "["+strconv.Itoa(j)+"]", "item")
			if err != nil {
				log.Printf("Request %d: Error: %s", i, err)
				continue
			}

			item := []byte(strings.Replace(itemRaw, "\\\"", "\"", -1))

			imageHUrlRaw, err := jsonparser.GetString(item, "ad", "image_fullscreen_001_landscape", "u")
			imageVUrlRaw, err := jsonparser.GetString(item, "ad", "image_fullscreen_001_portrait", "u")

			imageHUrl := strings.Replace(imageHUrlRaw, "\\/", "/", -1)
			imageVUrl := strings.Replace(imageVUrlRaw, "\\/", "/", -1)

			log.Printf("Request %d: Downloading...", i)
			download(i, imageHUrl, "horizontal")
			download(i, imageVUrl, "vertical")
		}
	}
}
