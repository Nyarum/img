package pixelate

import (
	"github.com/hawx/img/utils"
	"image"
	"image/color"
)

const (
	// Triangle types for Pxl
	BOTH = iota  // Decide base on closeness of colours in each quadrant
	LEFT         // Create only left triangles
	RIGHT        // Create only right triangles
)

func halve(img image.Image, pixelHeight, pixelWidth int) image.Image {
	b := img.Bounds()
	o := image.NewRGBA(image.Rect(0, 0, b.Dx() / 2, b.Dy() / 2))

	for y := 0; y < b.Dy() / 2; y++ {
		for x := 0; x < b.Dx() / 2; x++ {
			tl := img.At(x * 2, y * 2)
			tr := img.At(x * 2 + 1, y * 2)
			br := img.At(x * 2 + 1, y * 2 - 1)
			bl := img.At(x * 2, y * 2 - 1)

			if y % pixelHeight == 0 {
				o.Set(x, y, tl)
			} else if x % pixelWidth == 0 {
				o.Set(x, y, tl)
			} else {
				o.Set(x, y, utils.Average(tl, tr, bl, br))
			}
		}
	}

	return o
}

func pxlDo(img image.Image, triangle, pixelHeight, pixelWidth, scaleFactor int) image.Image {

	b := img.Bounds()

	cols  := b.Dx() / pixelWidth
	rows  := b.Dy() / pixelHeight
	ratio := float64(pixelHeight) / float64(pixelWidth)

	o := image.NewRGBA(image.Rect(0, 0, pixelWidth * cols * scaleFactor, pixelHeight * rows * scaleFactor))

	for col := 0; col < cols; col++ {
		for row := 0; row < rows; row++ {

			to := []color.Color{}
			ri := []color.Color{}
			bo := []color.Color{}
			le := []color.Color{}

			tc := 0; rc := 0; bc := 0; lc := 0

			for y := 0; y < pixelHeight; y++ {
				for x := 0; x < pixelWidth; x++ {

					realY := row * pixelHeight + y
					realX := col * pixelWidth + x

					y_origin := float64(y - pixelHeight / 2)
					x_origin := float64(x - pixelWidth / 2)

					if y_origin > ratio * x_origin && y_origin > ratio * -x_origin {
						tc++
						to = append(to, img.At(realX, realY))

					} else if y_origin < ratio * x_origin && y_origin > ratio * -x_origin {
						rc++
						ri = append(ri, img.At(realX, realY))

					} else if y_origin < ratio * x_origin && y_origin < ratio * -x_origin {
						bc++
						bo = append(bo, img.At(realX, realY))

					} else if y_origin > ratio * x_origin && y_origin < ratio * -x_origin {
						lc++
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

				top_right   := utils.Average(ato, ari)
				bottom_left := utils.Average(abo, ale)

				for y := 0; y < pixelHeight * scaleFactor; y++ {
					for x := 0; x < pixelWidth * scaleFactor; x++ {

						realY := row * pixelHeight * scaleFactor + y
						realX := col * pixelWidth * scaleFactor + x

						y_origin := float64(y - pixelHeight * scaleFactor / 2)
						x_origin := float64(x - pixelWidth * scaleFactor / 2)

						if y_origin > ratio * x_origin {
							o.Set(realX, realY, top_right)
						} else {
							o.Set(realX, realY, bottom_left)
						}
					}
				}

			} else {

				top_left     := utils.Average(ato, ale)
				bottom_right := utils.Average(abo, ari)

				for y := 0; y < pixelHeight * scaleFactor; y++ {
					for x := 0; x < pixelWidth * scaleFactor; x++ {

						realY := row * pixelHeight * scaleFactor + y
						realX := col * pixelWidth * scaleFactor + x

						y_origin := float64(y - pixelHeight * scaleFactor / 2)
						x_origin := float64(x - pixelWidth * scaleFactor / 2)

						if y_origin >= ratio * -x_origin {
							o.Set(realX, realY, top_left)
						} else {
							o.Set(realX, realY, bottom_right)
						}
					}
				}
			}
		}
	}

	return o
}


// Pxl pixelates an Image into right-angled triangles with the dimensions
// given. The triangle direction can be determined by passing the required value
// as triangle; either BOTH, LEFT or RIGHT.
func Pxl(img image.Image, triangle, pixelHeight, pixelWidth int, aliased bool) image.Image {
	if aliased {
		return pxlDo(img, triangle, pixelHeight, pixelWidth, 1)
	}
	return halve(pxlDo(img, triangle, pixelHeight, pixelWidth, 2), pixelHeight, pixelWidth)
}
