package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"os"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

func loadImage(filename string) image.Image {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	pic, err := jpeg.Decode(file)
	if err != nil {
		panic(err)
	}
	return pic
}

func saveImage(filename string, im image.Image) {
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	jpeg.Encode(file, im, &jpeg.Options{Quality: 100})
}

func newCanvas(w, h int) *image.RGBA {
	rect := image.Rect(0, 0, w, h)
	canvas := image.NewRGBA(rect)
	for i := range canvas.Pix {
		if i%250 == 0 {
			canvas.Pix[i] = 0
		} else {
			canvas.Pix[i] = 255
		}
	}
	return canvas
}

func fontFace(filename string, size float64) font.Face {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	f, err := freetype.ParseFont(b)
	if err != nil {
		panic(err)
	}
	return truetype.NewFace(f, &truetype.Options{Size: size})
}

func drawBackground(canvas draw.Image, rect image.Rectangle) {
	bg := image.NewUniform(color.RGBA{0, 0, 0, 100})
	draw.Draw(canvas, rect, bg, image.ZP, draw.Over)
}

func paragraphHeight(text []string, ff font.Face, lineSpacing int) int {
	return ff.Metrics().Ascent.Floor() +
		(ff.Metrics().Height.Floor()+lineSpacing)*len(text) - lineSpacing
}

func wordWrap(text string, width int, ff font.Face) []string {
	lineWidth := fixed.I(0)
	rs := []rune(text)
	line := ""
	lines := make([]string, 0)
	for i := 0; i < len(rs); i++ {
		r := rs[i]
		advance, ok := ff.GlyphAdvance(r)
		if !ok {
			// skipping unknown character
			continue
		}

		if lineWidth+advance < fixed.I(width) {
			line += string(r)
			lineWidth += advance
			if r == '\n' { // handle line breakers
				line = line[:len(line)-1]
				lines = append(lines, line)
				line = ""
				lineWidth = fixed.I(0)
			}
			if i == len(rs)-1 { // last loop
				lines = append(lines, line)
			}
		} else {
			lines = append(lines, line)
			line = ""
			lineWidth = fixed.I(0)
			i--
		}
	}
	return lines
}

func drawTextWordWrap(canvas draw.Image, lines []string,
	ff font.Face, lineSpacing, x, y int) {
	point := fixed.Point26_6{
		// X offset
		X: fixed.I(x),
		// Y offset of glyph
		// This value is accepted by font.Drawer as the Y value of baseline,
		// so Ascent value must be added
		Y: ff.Metrics().Ascent + fixed.I(y),
	}
	drawer := &font.Drawer{
		Src: image.NewUniform(color.Black),
		Dst: canvas,
		// Note that this is the baseline location
		Dot:  point,
		Face: ff,
	}

	for _, line := range lines {
		drawer.DrawString(line)
		point.Y += ff.Metrics().Height
		point.Y += fixed.I(lineSpacing)
		drawer.Dot = point
	}
}

func main() {
	const w, h = 500, 500
	const spacing = 2
	const leftMargin, rightMargin = 10, 10
	const text = `瓦亚格岛
	
瓦亚格岛是印度尼西亚西巴布亚省拉贾安帕特群岛的一部分。这些无人居住的小岛很受潜水者和浮潜者的欢迎，他们渴望探索周围巨大而多样的珊瑚礁系统。
瓦亚格岛是珊瑚礁三角区的一部分，虽然它只覆盖了地球上1.6%的海洋区域，但却包含了地球上所有已知的珊瑚物种的76%。`
	ff := fontFace("C:/Windows/Fonts/simsun.ttc", 24)
	canvas := newCanvas(w, h)
	lines := wordWrap(text, w/2, ff)
	drawBackground(canvas, image.Rect(leftMargin, rightMargin, w/2,
		paragraphHeight(lines, ff, spacing)))
	drawTextWordWrap(canvas, lines, ff, spacing, leftMargin, rightMargin)

	saveImage("output.jpg", canvas)
}
