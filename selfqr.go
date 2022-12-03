package selfqr

import (
	"bytes"
	"errors"
	"github.com/nfnt/resize"
	"github.com/skip2/go-qrcode"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
)

// New Create a new QR code processor
// `text`: Contents of QR Code
// `size`: Size of QR code
// `level`: Error correction level of two-dimensional code
func New(text string, size int, level qrcode.RecoveryLevel) *PersonaQrCode {
	p := &PersonaQrCode{size: size}
	p.qr, p.Error = qrcode.New(text, level)
	return p
}

type PersonaQrCode struct {
	qr                     *qrcode.QRCode
	img                    *image.RGBA
	size                   int
	foreground, background color.Color
	foreImage              image.Image
	bullEyeForeImageColor  bool
	thick                  int
	Error                  error
}

type (
	BullEyeStyle   int
	CodePointStyle int
)

const (
	// BesRectangle 矩形
	BesRectangle BullEyeStyle = iota
	// BesRoundedRectangle 圆角矩形
	BesRoundedRectangle
	// BesCircular 圆形
	BesCircular
	// BesRhombus 菱形
	BesRhombus
)

const (
	// CpsRectangle 矩形
	CpsRectangle CodePointStyle = iota
	// CpsRoundedRectangle 圆角矩形
	CpsRoundedRectangle
	// CpsCircular 圆形
	CpsCircular
	// CpsRhombus 菱形
	CpsRhombus
	// CpsIsoscelesTop 等边三角形
	CpsIsoscelesTop
	// CpsIsoscelesBottom 尖部向下的等边三角形
	CpsIsoscelesBottom
)

// Result Get the QR code that has been processed. Type is []byte
// `border`: Whether to draw a border area for the QR code. The color is the background color
func (p *PersonaQrCode) Result(border bool) ([]byte, error) {
	if p.Error != nil {
		return nil, p.Error
	}

	if p.img == nil {
		p.qr.DisableBorder = !border
		return p.qr.PNG(p.size)
	}

	if p.size != len(p.qr.Bitmap())*p.thick-1 {
		_img := image.NewRGBA(image.Rect(0, 0, p.size, p.size))
		draw.Draw(_img, _img.Bounds(), resize.Resize(uint(p.size),
			uint(p.size), p.img, resize.Lanczos3), image.Pt(0, 0), draw.Over)
		p.img = _img
	}

	if border {
		_img := image.NewRGBA(image.Rect(0, 0, p.size, p.size))
		for y := 0; y < p.size; y++ {
			for x := 0; x < p.size; x++ {
				_img.Set(x, y, p.background)
			}
		}
		draw.Draw(_img, _img.Bounds().Add(image.Pt(p.thick, p.thick)),
			resize.Resize(uint(p.img.Bounds().Max.X-(p.thick*2)),
				uint(p.img.Bounds().Max.Y-(p.thick*2)), p.img, resize.Lanczos3), image.Pt(0, 0), draw.Over)
		p.img = _img
	}

	var buf bytes.Buffer
	p.Error = png.Encode(&buf, p.img)
	return buf.Bytes(), p.Error

}

// Background color of the QR code
// Translucent colors are not recommended
func (p *PersonaQrCode) Background(col color.Color) {
	p.background = col
}

// Foreground color of the QR code
// Translucent colors are not recommended
func (p *PersonaQrCode) Foreground(col color.Color) {
	p.foreground = col
}

// ForeImage Foreground color values are based on the image
// Note that the selected image should not appear transparent or consistent with the background color
// `bullEye`: Whether to extend the color of the picture to ox eyes. If it is false take the foreground color
// This method should be executed before CodePoint and BullEye
func (p *PersonaQrCode) ForeImage(img image.Image, bullEye bool) {
	bgWidth, bgHeight := img.Bounds().Max.X, img.Bounds().Max.Y
	if p.size != bgWidth || p.size != bgHeight {
		img = resize.Resize(uint(p.size), uint(p.size), img, resize.Bilinear)
	}
	p.foreImage = img
	p.bullEyeForeImageColor = bullEye
}

// CodePoint Draw the two-dimensional code content area
// This method should be executed before BullEye
// 	`s`: The style of the code point
// 	`r`: The size of the code point The value ranges from 0.1 to 1
func (p *PersonaQrCode) CodePoint(s CodePointStyle, r float64) {
	p.hasDefault()

	if r > 1 {
		r = 1
	}
	if r <= 0 {
		r = 0.1
	}

	p.qr.DisableBorder = true
	bitmap := p.qr.Bitmap()

	p.thick = int(math.Round(float64(p.size / len(bitmap))))

	if p.size%len(bitmap) != 0 {
		size := len(bitmap)*p.thick - 1
		p.img = image.NewRGBA(image.Rect(0, 0, size, size))
	}

	for y := 0; y < p.img.Bounds().Max.Y; y++ {
		for x := 0; x < p.img.Bounds().Max.X; x++ {
			p.img.Set(x, y, p.background)
		}
	}

	for y, rows := range bitmap {
		for x, bit := range rows {
			if bit {
				if !((y < 7 && x < 7) || (y < 7 && x > len(rows)-8) || (y > len(bitmap)-8 && x < 7)) {

					var img image.Image
					var foreground color.Color
					if p.foreImage != nil {
						foreground = p.foreImage.At(x*p.thick, y*p.thick)
					} else {
						foreground = p.foreground
					}

					switch s {
					case CpsRectangle:
						img = rectangle(foreground, 2*p.thick, 2*p.thick, 0)
					case CpsRoundedRectangle:
						img = rectangle(foreground, 2*p.thick, 2*p.thick, p.thick/2)
					case CpsCircular:
						img = circular(foreground, 2*p.thick)
					case CpsRhombus:
						img = rhombus(foreground, 2*p.thick)
					case CpsIsoscelesTop:
						img = isosceles(foreground, 2*p.thick, true)
					case CpsIsoscelesBottom:
						img = isosceles(foreground, 2*p.thick, false)
					}

					draw.Draw(p.img, p.img.Bounds(), resize.Resize(uint(r*float64(p.thick)), uint(r*float64(p.thick)),
						img, resize.Bilinear),
						image.Pt(-(x*p.thick)-(int((1-r)*float64(p.thick)/2)), -(y*p.thick)-(int((1-r)*float64(p.thick)/2))), draw.Over)

				}
			}
		}
	}
}

// BullEye Draw the bull's eye area
// 	`os`: Bull eye outside style
// 	`is`: Bull eye inside style
func (p *PersonaQrCode) BullEye(os, is BullEyeStyle) {
	p.hasDefault()
	for _, xy := range [][]int{
		{0, 0},
		{0, p.img.Bounds().Max.Y - (7 * p.thick)},
		{p.img.Bounds().Max.Y - (7 * p.thick), 0},
	} {

		var img = image.NewRGBA(image.Rect(0, 0, p.size, p.size))
		var outside, cover, inner image.Image

		var foreground, background color.Color
		if p.foreImage != nil && p.bullEyeForeImageColor {
			foreground = p.foreImage.At(xy[0], xy[1])
		} else {
			foreground = p.foreground
		}

		if p.background.(color.RGBA).A == 0 {
			background = color.RGBA{R: 255, G: 255, B: 255, A: 255}
		} else {
			background = p.background
		}

		uniform := p.size / 7
		switch os {
		case BesRectangle:
			outside = rectangle(foreground, p.size, p.size, 0)
			cover = rectangle(background, p.size, p.size, 0)
		case BesRoundedRectangle:
			outside = rectangle(foreground, p.size, p.size, int(float64(p.size)*0.1))
			cover = rectangle(background, p.size, p.size, int(float64(p.size)*0.1))
		case BesCircular:
			outside = circular(foreground, p.size/2)
			cover = circular(background, p.size/2)
		case BesRhombus:
			outside = rhombus(foreground, p.size/2)
			cover = rhombus(background, p.size/2)
		}

		draw.Draw(img, img.Bounds(),
			resize.Resize(uint(p.size), uint(p.size), outside, resize.Bilinear), image.Pt(0, 0), draw.Over)

		oSize := uniform * (7 - 2)
		oOffset := (p.size - oSize) / 2
		draw.Draw(img, img.Bounds().Add(image.Pt(oOffset, oOffset)),
			resize.Resize(uint(oSize), uint(oSize), cover, resize.Bilinear), image.Pt(0, 0), draw.Over)

		if p.background.(color.RGBA).A == 0 {
			for y := 0; y < img.Bounds().Max.Y; y++ {
				for x := 0; x < img.Bounds().Max.X; x++ {
					rgba := img.At(x, y).(color.RGBA)
					if rgba.R == 255 && rgba.G == 255 && rgba.B == 255 && rgba.A == 255 {
						img.Set(x, y, color.RGBA{})
					}
				}
			}
		}

		switch is {
		case BesRectangle:
			inner = rectangle(foreground, p.size, p.size, 0)
		case BesRoundedRectangle:
			inner = rectangle(foreground, p.size, p.size, int(float64(p.size)*0.1))
		case BesCircular:
			inner = circular(foreground, p.size/2)
		case BesRhombus:
			inner = rhombus(foreground, p.size/2)
		}

		iSize := uniform * 3
		iOffset := (p.size - iSize) / 2
		draw.Draw(img, img.Bounds().Add(image.Pt(iOffset, iOffset)),
			resize.Resize(uint(iSize), uint(iSize), inner, resize.Lanczos3), image.Pt(0, 0), draw.Over)
		draw.Draw(p.img, p.img.Bounds(),
			resize.Resize(uint(7*p.thick), uint(7*p.thick), img, resize.Bilinear), image.Pt(-xy[0], -xy[1]), draw.Over)
	}
}

// Logo Draw a Logo
// `border`: Whether to draw a border for the logo. the border color is the background color
// At present, the detailed parameters of Logo are not extracted and accepted as variables
func (p *PersonaQrCode) Logo(img image.Image, border bool) {
	if p.qr.Level <= qrcode.Medium {
		p.Error = errors.New("the fault tolerance rate of QR code is too low will lead to unrecognition")
		return
	}
	_img := image.NewRGBA(image.Rect(0, 0, img.Bounds().Max.X, img.Bounds().Max.Y))
	if border {
		for y := 0; y < img.Bounds().Max.X; y++ {
			for x := 0; x < img.Bounds().Max.Y; x++ {
				_img.Set(x, y, p.background)
			}
		}
		draw.Draw(_img, _img.Bounds().Add(image.Pt(
			(img.Bounds().Max.X-int(float64(img.Bounds().Max.X)*0.8))/2,
			(img.Bounds().Max.Y-int(float64(img.Bounds().Max.Y)*0.8))/2),
		), resize.Resize(uint(float64(img.Bounds().Max.X)*0.8), uint(float64(img.Bounds().Max.Y)*0.8),
			img, resize.Lanczos3), image.Point{}, draw.Over)
	} else {
		draw.Draw(_img, _img.Bounds(), img, image.Point{}, draw.Over)
	}

	size := int(float64(p.thick) * 8)

	offset := image.Pt((p.size-size)/2, (p.size-size)/2)
	draw.Draw(p.img, p.img.Bounds().Add(offset),
		resize.Resize(uint(size), uint(size), _img, resize.Lanczos3), image.Pt(0, 0), draw.Over)
}

func (p *PersonaQrCode) hasDefault() {
	if p.img == nil {
		p.img = image.NewRGBA(image.Rect(0, 0, p.size, p.size))
	}
	if p.foreground == nil {
		p.foreground = color.RGBA{A: 255}
	}
	if p.background == nil {
		p.background = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}
}
