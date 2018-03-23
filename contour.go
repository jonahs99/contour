package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"os"

	"github.com/fogleman/gg"

	"github.com/jonahs99/vec"
)

type poly struct {
	verts []vec.Vec
	lines []vec.Vec
}

func (p *poly) computeLines() {
	p.lines = make([]vec.Vec, len(p.verts))
	if len(p.verts) == 0 {
		return
	}
	for i := 0; i < len(p.verts); i++ {
		j := (i + 1) % len(p.verts)
		p.lines[i] = p.verts[j]
		p.lines[i].Sub(p.verts[i])
	}
}

func circle(rad float64, n int) poly {
	innerAngle := math.Pi * 2 / float64(n)
	p := poly{}
	for i := 0; i < n; i++ {
		t := innerAngle*float64(i) + innerAngle/2
		v := vec.NewPolar(rad, t)
		v.X += rand.Float64() * 1
		v.Y += rand.Float64() * 1
		p.verts = append(p.verts, v)
	}
	p.computeLines()
	return p
}

func drawPoly(ctx *gg.Context, p poly) {
	n := len(p.verts)
	ctx.MoveTo(p.verts[n-1].X, p.verts[n-1].Y)
	for i := 0; i < n; i++ {
		ctx.LineTo(p.verts[i].X, p.verts[i].Y)
	}

	ctx.SetColor(color.Black)
	ctx.Stroke()
}

func downSample(img image.Image, n int) image.Image {
	bounds := img.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y
	n2 := uint32(n * n)
	out := image.NewRGBA(image.Rect(0, 0, w/n, h/n))
	for y := 0; y < h/n; y++ {
		for x := 0; x < w/n; x++ {
			var r, g, b uint32
			for i := 0; i < n; i++ {
				for j := 0; j < n; j++ {
					sample := img.At(x*n+j, y*n+i)
					sr, sg, sb, _ := sample.RGBA()
					r += sr
					g += sg
					b += sb
				}
			}
			r /= n2 * 256
			g /= n2 * 256
			b /= n2 * 256
			c := color.RGBA{uint8(r), uint8(g), uint8(b), 255}
			out.SetRGBA(x, y, c)
		}
	}
	return out
}

func main() {
	w, h := 600.0, 600.0
	ctx := gg.NewContext(int(w), int(h))
	ctx.Translate(w/2, h/2)
	ctx.Scale(2, 2)

	ctx.SetColor(color.White)
	ctx.Clear()

	contours := []poly{circle(12, 9)}

	targetDist := 5.0

	P, D := 0.016, -0.08

	nContours := 32

	for k := 1; k < nContours; k++ {
		follow := contours[len(contours)-1]

		start := vec.NewXY(12+float64(k)*targetDist*1.2, 0)
		particle := start
		vel := vec.NewXY(0, -2)

		contour := poly{}
		contour.verts = append(contour.verts, particle)

		lastDist := targetDist
		for j := 0; j < 4000; j++ {
			dist, pt := 100.0, vec.Vec{}
			for i := 0; i < len(follow.verts); i++ {
				d, p := PointToSegment(particle, follow.verts[i], follow.lines[i])
				if d < dist {
					dist = d
					pt = p
				}
			}
			correction := 0.0
			correction += (dist - targetDist) * P
			correction += (lastDist - dist) * D
			lastDist = dist
			steer := *pt.Sub(particle).Times(correction)

			vel.Add(steer)
			speed := vec.Mag(vel)
			if speed > 4 {
				vel.Times(4 / speed)
			} else if speed < 2 {
				vel.Times(2 / speed)
			}

			noiseMag := 0.2
			noise := vec.NewPolar(rand.Float64()*noiseMag, rand.Float64()*2*math.Pi)
			vel.Add(noise)

			particle.Add(vel)

			//if j > 10 && vec.Dist(particle, start) < 4 {
			//	break
			//}
			theta := vec.Theta(particle)
			if j > 10 && theta > 0 && theta < 0.04 {
				break
			}

			contour.verts = append(contour.verts, particle)
		}

		contour.computeLines()
		contours = append(contours, contour)

		fmt.Printf("%d contours drawn.\n", k)

	}
	lineWidth := 2.0
	//minWidth := 0.2
	ctx.SetLineWidth(lineWidth)
	for _, c := range contours {
		//ctx.SetLineWidth(float64(nContours-i)/float64(nContours)*lineWidth + minWidth)
		drawPoly(ctx, c)
	}

	img := ctx.Image()
	path := fmt.Sprintf("out-%d.png", 0)
	file, _ := os.Create(path)
	png.Encode(file, img)

	img = downSample(img, 2)
	path = fmt.Sprintf("aa-%d.png", 0)
	file, _ = os.Create(path)
	png.Encode(file, img)

}
