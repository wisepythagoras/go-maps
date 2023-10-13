package main

import (
	"fmt"
	"math"

	pixel "github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/pixelgl"
	"github.com/paulmach/orb/maptile"
)

type Map struct {
	Win                          *pixelgl.Window
	Center                       [2]float64
	MoveOffsetX, prevMoveOffsetX float64
	MoveOffsetY, prevMoveOffsetY float64
	XOffset, YOffset             float64
	Z                            maptile.Zoom
	mouseIsDown, mouseMove       bool
	startTileX, startTileY       int
	mouseDownVec                 *pixel.Vec
}

func (m *Map) LatLonToXY(lon, lat float64) (float64, float64) {
	C := (256 / (2 * math.Pi)) * math.Pow(2, float64(m.Z))

	x := C * ((lon * (math.Pi / 180)) + math.Pi)
	y := C * (math.Pi - math.Log(math.Tan((math.Pi/4)+(lat*(math.Pi/180))/2)))

	return x, y
}

func (m *Map) XYToLatLon(x, y float64) (float64, float64) {
	xcenter, ycenter := m.LatLonToXY(m.Center[0], m.Center[1])

	xPoint := xcenter - (maxWidth/2 - x)
	yPoint := ycenter - (maxHeight/2 - y)
	fmt.Println(xPoint, yPoint)

	C := (256 / (2 * math.Pi)) * math.Pow(2, float64(m.Z))
	M := (xPoint / C) - math.Pi
	N := -(yPoint / C) + math.Pi

	lonPoint := M * (180 / math.Pi)
	latPoint := ((math.Atan(math.Pow(math.E, N)) - (math.Pi / 4)) * 2) * (180 / math.Pi)

	return lonPoint, latPoint
}

func (m *Map) Update() {
	// Handle mouse events.
	tileLenMoveX := -(math.Round(m.MoveOffsetX/256) + 1)
	tileLenMoveY := (math.Round(m.MoveOffsetY/256) + 1)

	if m.Win.Pressed(pixelgl.MouseButton1) && !m.mouseIsDown {
		v := m.Win.MousePosition()
		m.mouseDownVec = &v
		m.mouseIsDown = true
	} else if m.Win.JustReleased(pixelgl.MouseButton1) && m.mouseIsDown {
		fmt.Println(m.Win.MousePosition())
		m.mouseIsDown = false
		m.mouseDownVec = nil

		if m.mouseMove {
			updateTiles(
				m.startTileX+int(tileLenMoveX),
				m.startTileY+int(tileLenMoveY),
				int(zoom),
				-(math.Round(m.MoveOffsetX/256)*256 + 256),
				-(math.Round(m.MoveOffsetY/256)*256 + 256),
			)

			m.prevMoveOffsetX = m.MoveOffsetX
			m.prevMoveOffsetY = m.MoveOffsetY

			m.mouseMove = false
		} else {
			posVec := m.Win.MousePosition()
			lon, lat := m.XYToLatLon(posVec.X, maxHeight-posVec.Y)
			fmt.Println(lat, lon)
		}
	} else if m.mouseIsDown {
		newPos := m.Win.MousePosition()
		_ = newPos

		// If the new position is the same as the original, then we have a click.
		// If it's different than the original position, then we're moving.

		if m.mouseDownVec.X != newPos.X || m.mouseDownVec.Y != newPos.Y {
			m.MoveOffsetX = m.prevMoveOffsetX + newPos.X - m.mouseDownVec.X
			m.MoveOffsetY = m.prevMoveOffsetY + newPos.Y - m.mouseDownVec.Y

			if !m.mouseMove {
				m.mouseMove = true
			}
		}
	}

	// Handle the scroll event (for zoom in and out).
	yScroll := m.Win.MouseScroll().Y

	if yScroll != 0 {
		zoom += int(yScroll)

		tile := maptile.At([2]float64{lon, lat}, maptile.Zoom(zoom))
		_, x, y := tile.Z, tile.X, tile.Y
		_, xf, yf := getTileURL(lat, lon, zoom)
		m.XOffset, m.YOffset = -(256*(xf-float64(x)) - 128), 256*(yf-float64(y))-128

		m.startTileX, m.startTileY = int(x)-tilesVertical, int(y)-tilesHorizontal

		updateTiles(
			m.startTileX,
			m.startTileY,
			int(zoom),
			-(math.Round(m.MoveOffsetX/256)*256 + 256),
			-(math.Round(m.MoveOffsetY/256)*256 + 256),
		)
	}
	// --

	// Draw all tiles.
	for j := len(tiles) - 1; j >= 0; j-- {
		for i := 0; i < len(tiles[j]); i++ {
			tile := tiles[j][i]

			if tile != nil {
				X := tile.X + 256 + (m.XOffset + m.MoveOffsetX)
				Y := tile.Y - 256 + (m.YOffset + m.MoveOffsetY)
				tileVec := pixel.V(X, Y)

				tile.Sprite.Draw(m.Win, pixel.IM.Moved(tileVec))
			}
		}
	}
}
