package main

import (
	"fmt"
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

func main() {
	w, h := 600.0, 600.0
	ctx := gg.NewContext(int(w), int(h))
	ctx.Translate(w/2, h/2)

	ctx.SetColor(color.White)
	ctx.Clear()

	contours := []poly{circle(12, 9)}

	targetDist := 24.0

	P, D := 0.02, -0.005

	nContours := 30

	for k := 1; k < nContours; k++ {
		follow := contours[len(contours)-1]

		start := vec.NewXY(12+float64(k)*targetDist, 0)
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

			if j > 10 && vec.Dist(particle, start) < 12 {
				break
			}

			contour.verts = append(contour.verts, particle)
		}

		contour.computeLines()
		contours = append(contours, contour)

		fmt.Printf("%d contours drawn.\n", k)

		lineWidth := 2.0
		minWidth := 0.2
		for i, c := range contours {
			ctx.SetLineWidth(float64(nContours-i)/float64(nContours)*lineWidth + minWidth)
			drawPoly(ctx, c)
		}

		img := ctx.Image()
		path := fmt.Sprintf("out-%d.png", k)
		file, _ := os.Create(path)
		png.Encode(file, img)
	}

}
