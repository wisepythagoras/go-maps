package main

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"math"
	"net/http"

	"github.com/faiface/pixel"
)

func updateTiles(firstX, firstY, z int, startX, startY float64) {
	loading = true

	for j := -1; j < tilesVertical*2+2; j++ {
		tiles[j+1] = make([]*MapTile, tilesHorizontal*2+3)

		for i := -1; i < tilesHorizontal*2+2; i++ {
			x, y := firstX+i, firstY+j

			// DEBUG
			// fmt.Println(url)
			// --

			pic, err := downloadImage(uint32(z), float64(x), float64(y))

			if err != nil {
				fmt.Println(err)
			} else {
				tiles[j+1][i+1] = &MapTile{
					X:      startX + float64(256*i) + (256 / 2),
					Y:      startY + float64(256*(len(tiles)-1-j)) + (256 / 2),
					Sprite: pixel.NewSprite(pic, pixel.R(0, 0, TILE_SIZE, TILE_SIZE)),
				}
			}
		}
	}

	loading = false
}

func deg2rad(d float64) float64 {
	return d * math.Pi / 180.0
}

// https://stackoverflow.com/questions/65494988/get-map-tiles-bounding-box
// https://stackoverflow.com/questions/62908635/leaflet-latlng-coordinates-to-xy-map
func getTileURL(lat, lon float64, zoom int) (uint32, float64, float64) {
	var xtile = (lon + 180) / 360 * float64((int(1) << zoom))
	var ytile = ((1 - math.Log(math.Tan(deg2rad(lat))+1/math.Cos(deg2rad(lat)))/math.Pi) / 2 * float64((int(1) << zoom)))
	return uint32(zoom), xtile, ytile
}

func downloadImage(z uint32, x, y float64) (pixel.Picture, error) {
	imgBytes := tileCache.Get(z, x, y)

	if imgBytes != nil {
		r := bytes.NewReader(imgBytes)
		img, _, err := image.Decode(r)

		if err != nil {
			return nil, err
		}

		return pixel.PictureDataFromImage(img), nil
	}

	url := fmt.Sprintf(tileTemplate, z, int(x), int(y))
	fmt.Println(url)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", USER_AGENT)
	response, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("received code %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	tileCache.Set(z, x, y, body)
	r := bytes.NewReader(body)
	img, _, err := image.Decode(r)

	if err != nil {
		return nil, err
	}

	return pixel.PictureDataFromImage(img), nil
}
