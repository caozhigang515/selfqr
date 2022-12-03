package selfqr

import (
	"github.com/nfnt/resize"
	"image"
	"image/color"
)

func rectangle(col color.Color, w, h, r int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if (x >= r || y >= r) && (x <= w-r || y <= h-r) && (x >= r || y <= h-r) && (x <= w-r || y >= r) {
				img.Set(x, y, col)
			}
		}
	}

	if r > 0 {
		for xx := 0; xx < r*2; xx++ {
			for yy := 0; yy < r*2; yy++ {
				if ((xx-r)*(xx-r) + (yy-r)*(yy-r)) <= r*r {
					img.Set(xx, yy, col)
					img.Set((w-(2*r))+xx, yy, col)
					img.Set(xx, (h-(2*r))+yy, col)
					img.Set((w-(2*r))+xx, (h-(2*r))+yy, col)
				}
			}
		}
	}
	return img
}

func circular(col color.Color, r int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 2*r, 2*r))
	for xx := 0; xx < r*2; xx++ {
		for yy := 0; yy < r*2; yy++ {
			if ((xx-r)*(xx-r) + (yy-r)*(yy-r)) <= r*r {
				img.Set(xx, yy, col)
			}
		}
	}
	return img
}

func isosceles(col color.Color, h int, top bool) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, h, h/2))
	for y := 0; y < h; y++ {
		for x := 0; x < h; x++ {
			if top {
				if y <= h/2 && x > (h/2)-y && x < (h/2)+y {
					img.Set(x, y, col)
				}
			} else {
				if y > h/2 && x > h/2-(h-y) && x < h/2+(h-y) {
					img.Set(x, y-h/2, col)
				}
			}
		}
	}
	return resize.Resize(uint(h), uint(h), img, resize.Lanczos3)
}

func rhombus(col color.Color, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, h, h))
	for y := 0; y < h; y++ {
		for x := 0; x < h; x++ {
			if y > h/2 {
				if (x > h/2-(h-y)) && x < h/2+(h-y) {
					img.Set(x, y, col)
				}
			} else {
				if x > (h/2)-y && x < (h/2)+y {
					img.Set(x, y, col)
				}
			}
		}
	}

	return img
}
