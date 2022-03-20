// Package agent defines a velocity obstacle object which is constructed from a
// line segment.
//
// The line segment obstacle is impermeable from either side.
package agent

import (
	"fmt"

	"github.com/downflux/go-geometry/2d/hyperplane"
	"github.com/downflux/go-geometry/2d/line"
	"github.com/downflux/go-geometry/2d/segment"
	"github.com/downflux/go-geometry/2d/vector"
	"github.com/downflux/go-orca/agent"
	"github.com/downflux/go-orca/internal/vo/line/cache"

	voagent "github.com/downflux/go-orca/internal/vo/agent"
)

type domain int

const (
	collisionLeft domain = iota
	collisionRight
	collisionLine

	collision
)

type VO struct {
	s segment.S
	v vector.V

	c cache.C
}

func New(s segment.S, v vector.V) *VO {
	if !s.Feasible() {
		panic(
			fmt.Sprintf(
				"cannot construct VO object, line segment %v is infeasible",
				s,
			),
		)
	}

	return &VO{
		s: s,
		v: v,
	}
}

type mockAgent struct {
	p vector.V
	v vector.V
}

func (a mockAgent) P() vector.V { return a.p }
func (a mockAgent) V() vector.V { return a.v }

// TODO(minkezhang): add VO.s max speed property
func (a mockAgent) S() float64  { return 0 }
func (a mockAgent) R() float64  { return 0 }
func (a mockAgent) T() vector.V { return a.P() }

// domain returns the side of the truncated cone nearest the relative velocity.
func (vo VO) domain(a agent.A, tau float64) domain {

	d := s(vo.s, a, tau).L().Distance(v(vo.v, a))
	if d <= a.R() {
		return collision
	}
	panic("unimplemented")
}

func (vo VO) ORCA(a agent.A, tau float64) hyperplane.HP {
	c := *cache.New(vo.s, vo.v, a, tau)

	// ld is the distance from the relative velocity between the line
	// obstacle and the agent to the VO cutoff line.
	ld := seg.L().Distance(vel)

	// t is the projected parametric value along the extended line. We need
	// to detect the case where t extends beyond the segment itself, and
	// seg.T() truncates at the segment endpoints.
	t := seg.L().T(vel)

	// Handle the case where the agent collides with the semicircle on the
	// left side of the line segment.
	if t <= seg.TMin() && vector.Magnitude(vector.Sub(vel, seg.L().L(seg.TMin()))) <= a.R() {
		return voagent.New(mockAgent{
			p: vo.s.L().L(vo.s.TMin()),
			v: vo.v,
		}).ORCA(a, tau)
	}

	// Handle the case where the agent collides with the semicircle on the
	// right side of the line segment.
	if t >= seg.TMax() && vector.Magnitude(vector.Sub(vel, seg.L().L(seg.TMax()))) <= a.R() {
		return voagent.New(mockAgent{
			p: vo.s.L().L(vo.s.TMax()),
			v: vo.v,
		}).ORCA(a, tau)
	}

	// Handle the case where the agent collides with the line segment
	// itself.
	if (seg.TMin() > t && t > seg.TMax()) && ld < a.R() {
		return *hyperplane.New(
			seg.L().P(),
			*vector.New(
				seg.L().N().Y(),
				-seg.L().N().X(),
			),
		)
	}

	return hyperplane.HP{}
}

// w returns the perpendicular vector from the line to the relative velocity v.
func (vo VO) w(a agent.A, tau float64) vector.V {
	seg := s(vo.s, a, tau)
	vec := v(vo.v, a)

	p, ok := seg.L().Intersect(*line.New(*vector.New(0, 0), vec))
	if !ok {
		return *vector.New(0, 0)
	}
	return vector.Sub(vec, seg.L().L(seg.T(p)))
}

// l returns the line segment extending from the base of the truncated cone to
// the "left" side of the object line.
func (vo VO) l(a agent.A) vector.V { return vector.V{} }

// r returns the line segment extending from the base of the truncated cone to
// the "right" side of the object line.
func (vo VO) r(a agent.A) vector.V { return vector.V{} }

// v returns the relative velocity between the agent and the obstacle line.
func v(v vector.V, a agent.A) vector.V { return vector.Sub(a.V(), v) }

// s generates a scaled line segment based on the lookahead time and the agent.
func s(s segment.S, a agent.A, tau float64) segment.S {
	return *segment.New(
		*line.New(
			vector.Scale(1/tau, vector.Sub(s.L().P(), a.P())),
			vector.Scale(1/tau, s.L().D()),
		),
		s.TMin(),
		s.TMax(),
	)
}
