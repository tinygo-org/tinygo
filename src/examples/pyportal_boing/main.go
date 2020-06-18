// Port of Adafruit's "pyportal_boing" demo found here:
// https://github.com/adafruit/Adafruit_ILI9341/blob/master/examples/pyportal_boing
package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/examples/ili9341/pyportal_boing/graphics"
	"tinygo.org/x/drivers/ili9341"
)

const (
	BGCOLOR    = 0xAD75
	GRIDCOLOR  = 0xA815
	BGSHADOW   = 0x5285
	GRIDSHADOW = 0x600C
	RED        = 0xF800
	WHITE      = 0xFFFF

	YBOTTOM = 123  // Ball Y coord at bottom
	YBOUNCE = -3.5 // Upward velocity on ball bounce

	_debug = false
)

var (
	frameBuffer = [(graphics.BALLHEIGHT + 8) * (graphics.BALLWIDTH + 8)]uint16{}

	startTime int64
	frame     int64

	// Ball coordinates are stored floating-point because screen refresh
	// is so quick, whole-pixel movements are just too fast!
	ballx     float32
	bally     float32
	ballvx    float32
	ballvy    float32
	ballframe float32
	balloldx  float32
	balloldy  float32

	// Color table for ball rotation effect
	palette [16]uint16
)

func main() {

	// configure backlight
	backlight.Configure(machine.PinConfig{machine.PinOutput})

	// configure display
	display.Configure(ili9341.Config{})
	print("width, height == ")
	width, height := display.Size()
	println(width, height)

	backlight.High()

	display.SetRotation(ili9341.Rotation270)
	DrawBackground()

	startTime = time.Now().UnixNano()
	frame = 0

	ballx = 20.0
	bally = YBOTTOM // Current ball position
	ballvx = 0.8
	ballvy = YBOUNCE // Ball velocity
	ballframe = 3    // Ball animation frame #
	balloldx = ballx
	balloldy = bally // Prior ball position

	for {

		balloldx = ballx // Save prior position
		balloldy = bally
		ballx += ballvx // Update position
		bally += ballvy
		ballvy += 0.06 // Update Y velocity
		if (ballx <= 15) || (ballx >= graphics.SCREENWIDTH-graphics.BALLWIDTH) {
			ballvx *= -1 // Left/right bounce
		}
		if bally >= YBOTTOM { // Hit ground?
			bally = YBOTTOM  // Clip and
			ballvy = YBOUNCE // bounce up
		}

		// Determine screen area to update.  This is the bounds of the ball's
		// prior and current positions, so the old ball is fully erased and new
		// ball is fully drawn.
		var minx, miny, maxx, maxy, width, height int16

		// Determine bounds of prior and new positions
		minx = int16(ballx)
		if int16(balloldx) < minx {
			minx = int16(balloldx)
		}
		miny = int16(bally)
		if int16(balloldy) < miny {
			miny = int16(balloldy)
		}
		maxx = int16(ballx + graphics.BALLWIDTH - 1)
		if int16(balloldx+graphics.BALLWIDTH-1) > maxx {
			maxx = int16(balloldx + graphics.BALLWIDTH - 1)
		}
		maxy = int16(bally + graphics.BALLHEIGHT - 1)
		if int16(balloldy+graphics.BALLHEIGHT-1) > maxy {
			maxy = int16(balloldy + graphics.BALLHEIGHT - 1)
		}

		width = maxx - minx + 1
		height = maxy - miny + 1

		// Ball animation frame # is incremented opposite the ball's X velocity
		ballframe -= ballvx * 0.5
		if ballframe < 0 {
			ballframe += 14 // Constrain from 0 to 13
		} else if ballframe >= 14 {
			ballframe -= 14
		}

		// Set 7 palette entries to white, 7 to red, based on frame number.
		// This makes the ball spin
		for i := 0; i < 14; i++ {
			if (int(ballframe)+i)%14 < 7 {
				palette[i+2] = WHITE
			} else {
				palette[i+2] = RED
			} // Palette entries 0 and 1 aren't used (clear and shadow, respectively)
		}

		// Only the changed rectangle is drawn into the 'renderbuf' array...
		var c uint16              //, *destPtr;
		bx := minx - int16(ballx) // X relative to ball bitmap (can be negative)
		by := miny - int16(bally) // Y relative to ball bitmap (can be negative)
		bgx := minx               // X relative to background bitmap (>= 0)
		bgy := miny               // Y relative to background bitmap (>= 0)
		var bx1, bgx1 int16       // Loop counters and working vars
		var p uint8               // 'packed' value of 2 ball pixels
		var bufIdx int8 = 0

		//tft.setAddrWindow(minx, miny, width, height)

		for y := 0; y < int(height); y++ { // For each row...
			//destPtr = &renderbuf[bufIdx][0];
			bx1 = bx   // Need to keep the original bx and bgx values,
			bgx1 = bgx // so copies of them are made here (and changed in loop below)
			for x := 0; x < int(width); x++ {
				var bgidx = int(bgy)*(graphics.SCREENWIDTH/8) + int(bgx1/8)
				if (bx1 >= 0) && (bx1 < graphics.BALLWIDTH) && // Is current pixel row/column
					(by >= 0) && (by < graphics.BALLHEIGHT) { // inside the ball bitmap area?
					// Yes, do ball compositing math...
					p = graphics.Ball[int(by*(graphics.BALLWIDTH/2))+int(bx1/2)] // Get packed value (2 pixels)
					if (bx1 & 1) != 0 {
						c = uint16(p & 0xF)
					} else {
						c = uint16(p >> 4)
					} // Unpack high or low nybble
					if c == 0 { // Outside ball - just draw grid
						if graphics.Background[bgidx]&(0x80>>(bgx1&7)) != 0 {
							c = GRIDCOLOR
						} else {
							c = BGCOLOR
						}
					} else if c > 1 { // In ball area...
						c = palette[c]
					} else { // In shadow area...
						if graphics.Background[bgidx]&(0x80>>(bgx1&7)) != 0 {
							c = GRIDSHADOW
						} else {
							c = BGSHADOW
						}
					}
				} else { // Outside ball bitmap, just draw background bitmap...
					if graphics.Background[bgidx]&(0x80>>(bgx1&7)) != 0 {
						c = GRIDCOLOR
					} else {
						c = BGCOLOR
					}
				}
				frameBuffer[y*int(width)+x] = c
				bx1++ // Increment bitmap position counters (X axis)
				bgx1++
			}
			//tft.dmaWait(); // Wait for prior line to complete
			//tft.writePixels(&renderbuf[bufIdx][0], width, false); // Non-blocking write
			bufIdx = 1 - bufIdx
			by++ // Increment bitmap position counters (Y axis)
			bgy++
		}

		display.DrawRGBBitmap(minx, miny, frameBuffer[:width*height], width, height)

		// Show approximate frame rate
		frame++
		if frame&255 == 0 { // Every 256 frames...
			elapsed := (time.Now().UnixNano() - startTime) / int64(time.Second)
			if elapsed > 0 {
				println(frame/elapsed, " fps")
			}
		}
	}
}

func DrawBackground() {
	w, h := display.Size()
	byteWidth := (w + 7) / 8 // Bitmap scanline pad = whole byte
	var b uint8
	for j := int16(0); j < h; j++ {
		for k := int16(0); k < w; k++ {
			if k&7 > 0 {
				b <<= 1
			} else {
				b = graphics.Background[j*byteWidth+k/8]
			}
			if b&0x80 == 0 {
				frameBuffer[k] = BGCOLOR
			} else {
				frameBuffer[k] = GRIDCOLOR
			}
		}
		display.DrawRGBBitmap(0, j, frameBuffer[0:w], w, 1)
	}
}
