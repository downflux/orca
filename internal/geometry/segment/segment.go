// Package segment defines a truncated cone-like object whose bottom is defined
// by a characteristic line segment (instead of a point). This object also has a
// characteristic "turning radius", which defines the sharpness of the curve
// from the bottom line segment to the edges.
//
//   L \     / R
//      \___/
//        S
//
// As with the case of the point-defined cone, we define tangential lines from
// the origin to the left and right circles of at the ends of the line segment
// S.
package segment

import (
	"fmt"

	"github.com/downflux/go-geometry/2d/hypersphere"
	"github.com/downflux/go-geometry/2d/line"
	"github.com/downflux/go-geometry/2d/segment"
	"github.com/downflux/go-geometry/2d/vector"
	"github.com/downflux/go-orca/internal/geometry/cone"

	ov "github.com/downflux/go-orca/internal/geometry/2d/vector"
)

type S struct {
	// s represents the relative physical line segment of the obstacle.
	s segment.S

	// r is the thickness of the line segment.
	r float64

	cl cone.C
	cr cone.C
}

func New(s segment.S, r float64) *S {
	cTMin, err := cone.New(*hypersphere.New(s.L().L(s.TMin()), r))
	if err != nil {
		panic(
			fmt.Sprintf(
				"could not construct line segment VO object: %v",
				err))
	}
	cTMax, err := cone.New(*hypersphere.New(s.L().L(s.TMax()), r))
	if err != nil {
		panic(
			fmt.Sprintf(
				"could not construct line segment VO object: %v",
				err))
	}

	cl := cTMin
	cr := cTMax

	t := s.L().T(*vector.New(0, 0))
	d := s.L().Distance(*vector.New(0, 0))

	if (
	// The right truncation circle is obstructing the view of the
	// left end of the line segment. Use the right circle to
	// calculate the left tangent leg.
	t >= s.TMax() && d <= r) || (
	// The agent is flipped across the truncation line.
	vector.Determinant(
		s.L().D(),
		s.L().L(s.TMin()),
	) < 0) {
		cl = cTMax
	}
	if (
	// The left truncation circle is obstructing the view of the
	// right end of the line segment. Use the left circle to
	// calculate the right tangent leg.
	t <= s.TMin() && d <= r) || (
	// The agent is flipped across the truncation line.
	vector.Determinant(
		s.L().D(),
		s.L().L(s.TMin()),
	) < 0) {
		cr = cTMin
	}

	return &S{
		s:  s,
		r:  r,
		cl: *cl,
		cr: *cr,
	}
}

// S returns the base of line segment VO. Depending on the setup, this segment
// may differ from the constructor input segment.
func (s S) S() segment.S {
	v := s.s
	// In the oblique case, we need to generate a new segment which
	// preserves the relative orientation between L, S, and R. We take as
	// the segment direction L + R, and set the root of the segment at the
	// base of the cone(s).
	if hypersphere.Within(s.cl.C(), s.cr.C()) {
		v = *segment.New(
			*line.New(
				s.cl.C().P(),
				vector.Add(
					s.L(),
					s.R(),
				),
			),
			0,
			0,
		)
	}
	if !ov.IsNormalOrientation(s.L(), v.L().D()) {
		v = *segment.New(
			*line.New(
				s.s.L().L(s.s.TMax()),
				vector.Scale(-1, v.L().D()),
			),
			v.TMin(),
			v.TMax(),
		)
	}
	return v
}

// IsLeftNegative calculates the relative orientation of L, S, and R.
//
// W do not enforce a convention of directionality for L, S, or R directly --
// that is, S here may point either to the left or right. We do enforce that L,
// S, and R are normally oriented with one another, i.e.
//
//   |L x S| > 0, and
//   |S x R| > 0
//
// This means the line segments defining the cone are actually directed either
// as
//
//     ____/ R  or  L \____
//   L \ S              S / R
//
// We can check for which orientation we are in by checking which end of the S
// corresponds with the base of the left line segment L.
//
// We (rather arbitrarily) define the "left-negative" orientation as the first
// case, and "left-positive" as the second.
//
// In the left-negative orientation, L is pointing downwards, S is right, and R
// is pointing upwards.
func (s S) IsLeftNegative() bool {
	return vector.Within(
		s.CL().C().P(),
		s.S().L().L(s.S().TMin()))
}

// L calculates the left vector of the tangent line from the agent position to
// the base of the truncated line segment.
//
// N.B.: ℓ may be generated by either the left or right truncation circle, but
// is always the left-side tangent line of that circle. The right truncation
// circle will be used if the it obstructs the view of the left truncation
// circle (the oblique case), or if the agent is "flipped" across the truncation
// line.
func (s S) L() vector.V { return s.CL().L() }

func (s S) R() vector.V { return s.CR().R() }
func (s S) CL() cone.C  { return s.cl }
func (s S) CR() cone.C  { return s.cr }
