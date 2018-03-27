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
	targetDist := 4.0
	//minTargetDistance := 1.0

	nPoints := 8000

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

	targetSpeed := 2.0

	//lastSpeed := 4.0
	//lastDist := targetDist

	lastControlError := vec.Vec{}
	cumControlError := vec.Vec{}

	particle := points[len(points)-1]
	vel := vec.NewXY(0, 4)

	for i := nSeed; i < nPoints; i++ {
		closestDist := 100.0
		var closestPoint vec.Vec
		for j := 1; j < len(points)-20; j++ {
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

		//speed := vec.Dot(vel, forward)

		radmag := vec.Mag(radial)
		radial.Times((targetDist - closestDist) / radmag)

		forward.Times(targetSpeed).Add(radial)

		// forward is the target
		P, I, D := 0.18, 0.00, 0.01

		controlError := vec.New(forward)
		controlError.Sub(vel)

		steerP := vec.New(controlError)
		steerP.Times(P)

		cumControlError.Add(controlError)
		steerI := vec.New(cumControlError)
		steerI.Times(I)

		steerD := vec.New(controlError)
		steerD.Sub(lastControlError)
		steerD.Times(D)
		lastControlError = controlError

		vel.Add(steerP)
		vel.Add(steerI)
		vel.Add(steerD)

		eps := 0.4
		vel.X += (rand.Float64()*2 - 1) * eps
		vel.Y += (rand.Float64()*2 - 1) * eps

		// Are we crossing the line?
		for j := 1; j < len(points)-20; j++ {
			t1, t2, intersect := vec.Intersect(particle, vel, points[j], lines[j])
			if intersect == 1 && 0 < t1 && t1 < 1 && 0 < t2 && t2 < 1 {
				// Scale back!
				fmt.Println(t1, t2, intersect)
				fmt.Println(j, len(points)-1)
				vel.Times(t1 * 0.5)
			}
		}

		// Advance the particle
		particle.Add(vel)
		plot(particle)

		//targetDist -= (targetDist - minTargetDistance) / float64(nPoints)

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
