package main

//frame rate independence
//Score
//Game over State - win/loss
//2 player vs playing comp
//Improve AI
//handeling window resize

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

type color struct {
	r, g, b byte
}
type pos struct {
	x, y float32
}
type ball struct {
	pos
	radius float32
	xv, yv float32
	color
}
type paddle struct {
	pos
	w, h  float32
	speed float32
	score int
	color
}

type gameState int

const (
	start gameState = iota
	play
)

var state = start

var nums = [][]byte{
	{
		1, 1, 1,
		1, 0, 1,
		1, 0, 1,
		1, 0, 1,
		1, 1, 1,
	},
	{
		0, 1, 0,
		0, 1, 0,
		0, 1, 0,
		0, 1, 0,
		0, 1, 0,
	},
	{
		1, 1, 1,
		0, 0, 1,
		1, 1, 1,
		1, 0, 0,
		1, 1, 1,
	},
	{
		1, 1, 1,
		0, 0, 1,
		1, 1, 1,
		0, 0, 1,
		1, 1, 1}}

const winWidth, winHeight int = 600, 600

var aiwinner int
var offset int

func lerp(a float32, b float32, pct float32) float32 {
	return a + pct*(b-a)
}

func (paddle *paddle) draw(pixels []byte) {
	startX := paddle.x - paddle.w/2
	startY := paddle.y - paddle.h/2

	for y := 0; y < int(paddle.h); y++ {
		for x := 0; x < int(paddle.w); x++ {
			setPixel(int(startX)+x, int(startY)+y, color{255, 255, 255}, pixels)
		}
	}
	numX := lerp(paddle.x, getCenter().x, 0.2)
	drawNumber(pos{numX, 35}, paddle.color, 10, paddle.score, pixels)
}
func (ball *ball) draw(pixels []byte) {
	//YAGNI - Ya aint gonna need it
	for y := -ball.radius; y < ball.radius; y++ {
		for x := -ball.radius; x < ball.radius; x++ {
			if x*x+y*y < ball.radius*ball.radius {
				setPixel(int(ball.x+x), int(ball.y+y), color{255, 255, 255}, pixels)
			}
		}
	}
}
func drawNumber(pos pos, color color, size int, num int, pixels []byte) {
	startX := int(pos.x) - (size*3)/2
	startY := int(pos.y) - (size*3)/2

	for i, v := range nums[num] {
		if v == 1 {
			for y := startY; y < startY+size; y++ {
				for x := startX; x < startX+size; x++ {
					setPixel(x, y, color, pixels)
				}
			}
		}
		startX += size
		if (i+1)%3 == 0 {
			startY += size
			startX -= size * 3
		}
	}
}
func getCenter() pos {
	return pos{float32(winWidth / 2), float32(winHeight / 2)}
}
func (ball *ball) update(leftPaddle *paddle, rightPaddle *paddle, elapsedTime float32, aiwinner *int, offset *int) {
	ball.x += ball.xv * elapsedTime
	ball.y += ball.yv * elapsedTime

	//handle collisions
	if ball.y-ball.radius < 0 || int(ball.y+ball.radius) > winHeight { //top of screen
		ball.yv = -ball.yv
		*aiwinner = rand.Intn(2)
		*offset = rand.Intn(200) - 100
	}

	if ball.x < 0 {
		rightPaddle.score++
		ball.pos = getCenter()
		state = start
	} else if int(ball.x) > winWidth {
		leftPaddle.score++
		ball.pos = getCenter()
		state = start
	}

	if ball.x-ball.radius < leftPaddle.x+leftPaddle.w/2 {
		if ball.y > leftPaddle.y-leftPaddle.h/2 && ball.y < leftPaddle.y+leftPaddle.h/2 {
			ball.xv = -ball.xv
			ball.x = leftPaddle.x + leftPaddle.w/2.0 + ball.radius
		}
	}
	if ball.x+ball.radius > rightPaddle.x-rightPaddle.w/2 {
		if ball.y > rightPaddle.y-rightPaddle.h/2 && ball.y < rightPaddle.y+rightPaddle.h/2 {
			ball.xv = -ball.xv
			ball.x = rightPaddle.x - rightPaddle.w/2.0 - ball.radius
		}
	}
}

func (paddle *paddle) update(keyState []uint8, elapsedTime float32) {
	if keyState[sdl.SCANCODE_UP] != 0 {
		paddle.y -= paddle.speed * elapsedTime
	}
	if keyState[sdl.SCANCODE_DOWN] != 0 {
		paddle.y += paddle.speed * elapsedTime
	}
}
func (paddle *paddle) aiUpdate(ball *ball, elapsedTime float32, aiwinner *int, offset *int) {
	if *aiwinner == 0 {
		paddle.y = ball.y
	} else {
		paddle.y = ball.y + float32(*offset)
	}

}

func clear(pixels []byte) {
	for i := range pixels {
		pixels[i] = 0
	}
}

func setPixel(x, y int, c color, pixels []byte) {
	index := (y*winWidth + x) * 4
	if index < len(pixels)-4 && index >= 0 {
		pixels[index] = c.r
		pixels[index+1] = c.g
		pixels[index+2] = c.b
	}
}

func main() {
	window, err := sdl.CreateWindow("Testing SDL2", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, int32(winWidth), int32(winHeight), sdl.WINDOW_RESIZABLE)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer window.Destroy() //gets called at the bottom when the function exits, if you have a big game loop that can be helpful

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer renderer.Destroy()

	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(winWidth), int32(winHeight))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tex.Destroy()

	pixels := make([]byte, winWidth*winHeight*4)
	/*
		for y := 0; y < 800; y++ {
			for z := 0; z < 600; z++ {
				setPixel(z, y, color{byte(z % 255), byte(y % 255), 0}, pixels)
			}
		}*/

	play1 := paddle{pos{float32(winWidth - 590), 100}, 20, 100, 400, 0, color{255, 255, 255}}
	play2 := paddle{pos{float32(winWidth - 10), 500}, 20, 100, 400, 0, color{255, 255, 255}}
	ball := ball{pos{200, 200}, 15, 200, 200, color{255, 255, 255}}

	keyState := sdl.GetKeyboardState()

	var frameStart time.Time
	var elapsedTime float32

	//sdl.Delay(9000)
	/* for i := 0; i < 4000; i++ {
		sdl.PumpEvents()
		sdl.Delay(1)
	}
	*/
	for {
		frameStart = time.Now()
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
		}
		guess := &aiwinner
		paddleVariation := &offset
		if state == play {
			ball.update(&play1, &play2, elapsedTime, guess, paddleVariation)
			play1.update(keyState, elapsedTime)
			play2.aiUpdate(&ball, elapsedTime, guess, paddleVariation)

		} else if state == start {
			aiwinner = rand.Intn(2)
			if keyState[sdl.SCANCODE_SPACE] != 0 {
				if play1.score == 3 || play2.score == 3 {
					play1.score = 0
					play2.score = 0
				}
				state = play
			}
		}

		clear(pixels)
		play1.draw(pixels)
		play2.draw(pixels)
		ball.draw(pixels)

		tex.Update(nil, pixels, 600*4)
		renderer.Copy(tex, nil, nil)
		renderer.Present()
		elapsedTime = float32(time.Since(frameStart).Seconds())
		if elapsedTime < .005 {
			sdl.Delay(5 - uint32(elapsedTime/1000.0))
			elapsedTime = float32(time.Since(frameStart).Seconds())
		}
	}
}
