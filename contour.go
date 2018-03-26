package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"

	svg "github.com/ajstarks/svgo"

	"github.com/jonahs99/vec"
)

func main() {
	targetDist := 7.0
	minTargetDistance := 1.0
	//P, D := 0.016, -0.08

	nPoints := 100000

	points := make([]vec.Vec, 0, nPoints)
	lines := make([]vec.Vec, 0, nPoints-1)

	plot := func(v vec.Vec) {
		points = append(points, v)
		if len(points) > 1 {
			v.Sub(points[len(points)-2])
			lines = append(lines, v)
		}
	}

	// The "seed"
	nSeed := 40
	for i := 0; i < nSeed; i++ {
		frac := float64(i) / float64(nSeed)
		p := vec.NewPolar(10+targetDist*frac, 2*math.Pi*frac)
		plot(p)
	}

	// The path

	targetSpeed := 3.0

	//lastSpeed := 4.0
	//lastDist := targetDist

	particle := points[len(points)-1]
	vel := vec.NewXY(0, 4)
	for i := nSeed; i < nPoints; i++ {
		closestDist := 100.0
		var closestPoint vec.Vec
		for j := 0; j < len(lines)-20; j++ {
			dist, pt := PointToSegment(particle, points[j], lines[j])
			if dist < closestDist {
				closestDist = dist
				closestPoint = pt
			}
		}

		radial := vec.New(particle)
		radial.Sub(closestPoint)
		forward := vec.Norm(radial)
		forward.Div(vec.Mag(forward))

		/*speed := vec.Dot(vel, forward)

		distError := targetDist - closestDist
		speedError := targetSpeed - speed

		P := 0.04
		D := -0.01

		radial.Times(distError*P + (closestDist-lastDist)*D)
		forward.Times(speedError*P + (speed-lastSpeed)*D)
		steer := *radial.Add(forward)

		steerMag := vec.Mag(steer)
		if steerMag > 0.4 {
			steer.Times(0.4 / steerMag)
		}

		lastDist = closestDist
		lastSpeed = speed

		vel.Add(steer)*/

		radmag := vec.Mag(radial)
		radial.Times((targetDist - closestDist) / radmag)

		forward.Times(targetSpeed).Add(radial)

		forward.Times(0.6)
		vel.Times(0.4)
		vel.Add(forward)

		eps := 0.02
		vel.X += (rand.Float64()*2 - 1) * eps
		vel.Y += (rand.Float64()*2 - 1) * eps

		particle.Add(vel)
		plot(particle)

		targetDist -= (targetDist - minTargetDistance) / float64(nPoints)

		if i%1000 == 999 {
			fmt.Printf("%d points done\n", i+1)
		}
	}

	// Write output

	path := fmt.Sprintf("out.svg")
	file, err := os.Create(path)
	if err != nil {
		fmt.Printf("Could not create the output file: '%v'\n", err)
		os.Exit(-1)
	}

	w, h := 1000, 1000

	canvas := svg.New(file)
	canvas.Start(w, h)
	canvas.Translate(w/2, h/2)
	x, y := make([]int, 0, len(points)), make([]int, 0, len(points))
	for i := 0; i < len(points); i++ {
		x = append(x, int(points[i].X))
		y = append(y, int(points[i].Y))
	}
	canvas.Polyline(x, y, "fill: none; stroke: black; stroke-width: 1.2; stroke-linejoin: round")

	canvas.Gend()
	canvas.End()
}
