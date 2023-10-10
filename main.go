package main

import (
	"flag"
	"fmt"
	"math"
	"strconv"
	"strings"

	_ "image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/paulmach/orb/maptile"
	"golang.org/x/image/colornames"
)

const (
	maxWidth  = 1280
	maxHeight = 768
	// https://tile.openstreetmap.org/{z}/{x}/{y}.png
	// tileTemplate = "https://tile.openstreetmap.org/%d/%d/%d.png"
	// https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}
	// https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png
	tileTemplate = "https://a.basemaps.cartocdn.com/dark_all/%d/%d/%d.png"
	USER_AGENT   = "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/117.0"
	TILE_SIZE    = 256
)

var (
	// tiles2          [][]*pixel.Sprite
	tiles           [][]*MapTile
	tilesHorizontal int
	tilesVertical   int
	lon             float64 = -73.994808
	lat             float64 = 40.726966
	zoom            int     = 14
	loading                 = false
)

func main() {
	var coords string
	flag.StringVar(&coords, "coords", "", "The coordinates in lon,lat format")
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

	tilesHorizontal = ((maxWidth / 2) - (TILE_SIZE / 2)) / TILE_SIZE
	tilesVertical = ((maxHeight / 2) - (TILE_SIZE / 2)) / TILE_SIZE
	tiles = make([][]*MapTile, tilesVertical*2+3)

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

	// Initialize the tiles.
	updateTiles(startTileX, startTileY, int(z), 0, 0)

	movedI := float64(0)
	var mouseDownVec *pixel.Vec
	mouseIsDown := false
	mouseMove := false
	moveOffsetX, prevMoveOffsetX := float64(0), float64(0)
	moveOffsetY, prevMoveOffsetY := float64(0), float64(0)

	for !win.Closed() {
		d := pixel.V(-movedI, 0)
		_ = d

		imd.Clear()
		bg.Clear()

		bg.Color = colornames.Black
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

		// Handle mouse events.
		tileLenMoveX := -(math.Round(moveOffsetX/256) + 1)
		tileLenMoveY := (math.Round(moveOffsetY/256) + 1)

		if win.Pressed(pixelgl.MouseButton1) && !mouseIsDown {
			v := win.MousePosition()
			mouseDownVec = &v
			mouseIsDown = true
		} else if win.JustReleased(pixelgl.MouseButton1) && mouseIsDown {
			fmt.Println(win.MousePosition())
			mouseIsDown = false
			mouseDownVec = nil

			if mouseMove {
				fmt.Println(tileLenMoveX, tileLenMoveY, moveOffsetX, moveOffsetY)

				// This makes the center jerk around a bit. The reason is that when the map is refetched, the
				// center of the map will match the center of the central tile. Which is not where the user
				// dragged the map to. TODO: Fix this.
				updateTiles(
					startTileX+int(tileLenMoveX),
					startTileY+int(tileLenMoveY),
					int(z),
					-(math.Round(moveOffsetX/256)*256 + 256),
					-(math.Round(moveOffsetY/256)*256 + 256),
				)

				prevMoveOffsetX = moveOffsetX
				prevMoveOffsetY = moveOffsetY

				mouseMove = false
			}
		} else if mouseIsDown {
			newPos := win.MousePosition()
			_ = newPos

			// If the new position is the same as the original, then we have a click.
			// If it's different than the original position, then we're moving.

			if mouseDownVec.X != newPos.X || mouseDownVec.Y != newPos.Y {
				moveOffsetX = prevMoveOffsetX + newPos.X - mouseDownVec.X
				moveOffsetY = prevMoveOffsetY + newPos.Y - mouseDownVec.Y

				if !mouseMove {
					mouseMove = true
				}
			}
		}

		// Handle the scroll event (for zoom in and out).
		yScroll := win.MouseScroll().Y

		if yScroll != 0 {
			zoom += int(yScroll)

			tile := maptile.At([2]float64{lon, lat}, maptile.Zoom(zoom))
			z, x, y = tile.Z, tile.X, tile.Y
			_, xf, yf = getTileURL(lat, lon, zoom)
			xOffset, yOffset = -(256*(xf-float64(x)) - 128), 256*(yf-float64(y))-128

			startTileX, startTileY = int(x)-tilesVertical, int(y)-tilesHorizontal

			updateTiles(
				startTileX,
				startTileY,
				int(z),
				-(math.Round(moveOffsetX/256)*256 + 256),
				-(math.Round(moveOffsetY/256)*256 + 256),
			)
		}
		// --

		// Draw all tiles.
		for j := len(tiles) - 1; j >= 0; j-- {
			for i := 0; i < len(tiles[j]); i++ {
				tile := tiles[j][i]
				X := tile.X + 256 + (xOffset + moveOffsetX)
				Y := tile.Y - 256 + (yOffset + moveOffsetY)
				tileVec := pixel.V(X, Y)

				tile.Sprite.Draw(win, pixel.IM.Moved(tileVec.Sub(d)))
			}
		}
		// --

		imd.Draw(win)
		win.Update()
	}
}
