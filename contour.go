package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"os"

	svg "github.com/ajstarks/svgo"

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

func drawPoly(canvas *svg.SVG, p poly) {
	n := len(p.verts)

	x, y := make([]int, n), make([]int, n)

	for i := 0; i < n; i++ {
		x[i] = int(p.verts[i].X)
		y[i] = int(p.verts[i].Y)
	}

	canvas.Polyline(x, y, "fill: none; stroke: black; stroke-width: 1;")
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
	w, h := 1200, 1200

	contours := []poly{circle(12, 9)}

	targetDist := 6.0

	P, D := 0.016, -0.08

	nContours := 72

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

	path := fmt.Sprintf("out.svg")
	file, err := os.Create(path)
	if err != nil {
		fmt.Printf("Could not create the output file: '%v'\n", err)
		os.Exit(-1)
	}

	canvas := svg.New(file)
	canvas.Start(w, h)
	canvas.Translate(w/2, h/2)

	for _, c := range contours {
		drawPoly(canvas, c)
	}

	canvas.Gend()
	canvas.End()

}
