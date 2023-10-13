package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	_ "image/png"

	pixel "github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/imdraw"
	"github.com/gopxl/pixel/v2/pixelgl"
	"github.com/paulmach/orb/maptile"
	"golang.org/x/image/colornames"
)

const (
	maxWidth  = 1280
	maxHeight = 768
	// https://tile.openstreetmap.org/{z}/{x}/{y}.png
	// tileTemplate = "https://tile.openstreetmap.org/%d/%d/%d.png"
	// https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}.png
	// https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png
	USER_AGENT = "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/117.0"
	TILE_SIZE  = 256
)

var (
	tiles           [][]*MapTile
	tilesHorizontal int
	tilesVertical   int
	lon             float64 = -73.994808
	lat             float64 = 40.726966
	zoom            int     = 14
	tileCache       *TileCache
	tileTemplate    = "https://a.basemaps.cartocdn.com/dark_all/%d/%d/%d.png"
)

func main() {
	var coords string
	var tempTileTemplate string
	flag.StringVar(&coords, "coords", "", "The coordinates in lon,lat format")
	flag.StringVar(&tempTileTemplate, "tmpl", "", "The tile url")
	flag.Parse()

	if len(coords) > 0 {
		parts := strings.Split(coords, ",")

		if len(parts) != 2 {
			fmt.Printf("Invalid coordinates %q\n", coords)
			return
		}

		var err error
		lon, err = strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)

		if err != nil {
			fmt.Println("Error parsing longitude:", err)
			return
		}

		lat, err = strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)

		if err != nil {
			fmt.Println("Error parsing latitude:", err)
			return
		}
	}

	if len(tempTileTemplate) > 0 {
		tileTemplate = tempTileTemplate
	}

	tilesHorizontal = ((maxWidth / 2) - (TILE_SIZE / 2)) / TILE_SIZE
	tilesVertical = ((maxHeight / 2) - (TILE_SIZE / 2)) / TILE_SIZE
	tiles = make([][]*MapTile, tilesVertical*2+3)

	tileCache = &TileCache{}
	tileCache.Init()

	pixelgl.Run(run)
}

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Go Maps",
		Bounds: pixel.R(0, 0, maxWidth, maxHeight),
		VSync:  true,
	}

	win, err := pixelgl.NewWindow(cfg)

	if err != nil {
		fmt.Println(err)
		return
	}

	imd := imdraw.New(nil)
	bg := imdraw.New(nil)

	tile := maptile.At([2]float64{lon, lat}, maptile.Zoom(zoom))
	z, x, y := tile.Z, tile.X, tile.Y
	_, xf, yf := getTileURL(lat, lon, zoom)
	xOffset, yOffset := -(256*(xf-float64(x)) - 128), 256*(yf-float64(y))-128
	fmt.Println(xOffset, yOffset)

	startTileX, startTileY := int(x)-tilesVertical, int(y)-tilesHorizontal

	mapInstance := Map{
		Win:        win,
		Center:     [2]float64{lon, lat},
		Z:          maptile.Zoom(zoom),
		startTileX: startTileX,
		startTileY: startTileY,
	}

	// Initialize the tiles.
	updateTiles(startTileX, startTileY, int(z), 0, 0)

	for !win.Closed() {
		imd.Clear()
		bg.Clear()

		bg.Color = colornames.Darkgrey
		bg.Push(pixel.V(0, 0), pixel.V(maxWidth, maxHeight))
		bg.Rectangle(0)
		bg.Draw(win)

		// This bit draws the crosshair at the scenter of the screen.
		imd.Color = colornames.Aquamarine
		imd.Push(
			pixel.V(maxWidth/2-6, maxHeight/2),
			pixel.V(maxWidth/2+6, maxHeight/2),
		)
		imd.Line(1)
		imd.Push(
			pixel.V(maxWidth/2, maxHeight/2-6),
			pixel.V(maxWidth/2, maxHeight/2+6),
		)
		imd.Line(1)
		// --

		mapInstance.Update()
		imd.Draw(win)
		win.Update()
	}
}
