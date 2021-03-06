package pixelate

import (
	"image"
	"image/color"
	"image/draw"
	"runtime"

	"github.com/Nyarum/img/utils"
)

// Triangle types for Pxl
type Triangle int

const (
	// Decide base on closeness of colours in each quadrant
	BOTH Triangle = iota

	// Create only left triangles
	LEFT

	// Create only right triangles
	RIGHT
)

func pxlWorker(img image.Image, bounds image.Rectangle, dest draw.Image,
	size utils.Dimension, triangle Triangle, aliased bool, c chan int) {

	ratio := float64(size.H) / float64(size.W)

	inTop := func(x, y float64) bool {
		return (y > ratio*x) && (y > ratio*-x)
	}

	inRight := func(x, y float64) bool {
		return (y < ratio*x) && (y > ratio*-x)
	}

	inBottom := func(x, y float64) bool {
		return (y < ratio*x) && (y < ratio*-x)
	}

	inLeft := func(x, y float64) bool {
		return (y > ratio*x) && (y < ratio*-x)
	}

	to := []color.Color{}
	ri := []color.Color{}
	bo := []color.Color{}
	le := []color.Color{}

	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {

			realY := bounds.Min.Y + y
			realX := bounds.Min.X + x

			yOrigin := float64(y - size.H/2)
			xOrigin := float64(x - size.W/2)

			if inTop(xOrigin, yOrigin) {
				to = append(to, img.At(realX, realY))

			} else if inRight(xOrigin, yOrigin) {
				ri = append(ri, img.At(realX, realY))

			} else if inBottom(xOrigin, yOrigin) {
				bo = append(bo, img.At(realX, realY))

			} else if inLeft(xOrigin, yOrigin) {
				le = append(le, img.At(realX, realY))
			}
		}
	}

	ato := utils.Average(to...)
	ari := utils.Average(ri...)
	abo := utils.Average(bo...)
	ale := utils.Average(le...)

	if (triangle != LEFT) && (triangle == RIGHT ||
		utils.Closeness(ato, ari) > utils.Closeness(ato, ale)) {

		topRight := utils.Average(ato, ari)
		bottomLeft := utils.Average(abo, ale)
		middle := utils.Average(topRight, bottomLeft)

		for y := 0; y < bounds.Dy(); y++ {
			for x := 0; x < bounds.Dx(); x++ {

				realY := bounds.Min.Y + y
				realX := bounds.Min.X + x

				yOrigin := float64(y - size.H/2)
				xOrigin := float64(x - size.W/2)

				if yOrigin > ratio*xOrigin {
					dest.Set(realX, realY, topRight)
				} else if yOrigin == ratio*xOrigin && !aliased {
					dest.Set(realX, realY, middle)
				} else {
					dest.Set(realX, realY, bottomLeft)
				}
			}
		}

	} else {

		topLeft := utils.Average(ato, ale)
		bottomRight := utils.Average(abo, ari)
		middle := utils.Average(topLeft, bottomRight)

		for y := 0; y < bounds.Dy(); y++ {
			for x := 0; x < bounds.Dx(); x++ {

				realY := bounds.Min.Y + y
				realX := bounds.Min.X + x

				yOrigin := float64(y - size.H/2)
				xOrigin := float64(x - size.W/2)

				// Do this one opposite to above so that the diagonals line up when
				// aliased.
				if yOrigin < ratio*-xOrigin {
					dest.Set(realX, realY, bottomRight)
				} else if yOrigin == ratio*-xOrigin && !aliased {
					dest.Set(realX, realY, middle)
				} else {
					dest.Set(realX, realY, topLeft)
				}
			}
		}
	}

	c <- 1
}

func doPxl(img image.Image, size utils.Dimension, triangle Triangle, style Style, aliased bool) image.Image {

	nCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(nCPU)

	var o draw.Image
	b := img.Bounds()
	c := make(chan int, nCPU)
	i := 0 // number of workers created. there may be a better way...

	switch style {
	case CROPPED:
		cols := b.Dx() / size.W
		rows := b.Dy() / size.H

		o = image.NewRGBA(image.Rect(0, 0, size.W*cols, size.H*rows))

		for j, r := range utils.ChopRectangleToSizes(b, size.H, size.W, utils.IGNORE) {
			go pxlWorker(img, r, o, size, triangle, aliased, c)
			i = j
		}

	case FITTED:
		o = image.NewRGBA(img.Bounds())

		for j, r := range utils.ChopRectangleToSizes(img.Bounds(), size.H, size.W, utils.SEPARATE) {
			go pxlWorker(img, r, o, size, triangle, aliased, c)
			i = j
		}
	}

	for j := 0; j < i; j++ {
		<-c
	}

	return o
}

// Pxl pixelates an Image into right-angled triangles with the dimensions
// given. The triangle direction can be determined by passing the required value
// as triangle; either BOTH, LEFT or RIGHT.
func Pxl(img image.Image, size utils.Dimension, triangle Triangle, style Style) image.Image {
	return doPxl(img, size, triangle, style, false)
}

// AliasedPxl does the same as Pxl, but does not smooth diagonal edges of the
// triangles. It is faster, but will produce bad results if size is non-square.
func AliasedPxl(img image.Image, size utils.Dimension, triangle Triangle, style Style) image.Image {
	return doPxl(img, size, triangle, style, true)
}
