package main

import (
	"github.com/jonahs99/vec"
)

// ScalarProject returns the distance along v1 that v2 projects onto
func ScalarProject(v1, v2 vec.Vec) float64 {
	return vec.Dot(v1, v2) / vec.Mag2(v1)
}

// PointToSegment returns the distance to a line segment and the closest point on the segment
func PointToSegment(v, o, d vec.Vec) (float64, vec.Vec) {
	v.Sub(o)
	t := ScalarProject(d, v)
	if t < 0 {
		return vec.Mag(v), o
	} else if t > 1 {
		v.Sub(d)
		return vec.Mag(v), *o.Add(d)
	}
	d.Times(t)
	v.Sub(d)
	return vec.Mag(v), *o.Add(d)
}
