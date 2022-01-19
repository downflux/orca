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
)

type domain int

const (
	collision domain = iota
)

type VO struct {
	s segment.S
	v vector.V
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

// domain returns the side of the truncated cone nearest the relative velocity.
func (vo VO) domain(a agent.A, tau float64) domain {
	d := s(vo.s, a, tau).L().Distance(v(vo.v, a))
	if d <= a.R() {
		return collision
	}
	panic("unimplemented")
}

func (vo VO) ORCA(a agent.A, tau float64) hyperplane.HP {
	return hyperplane.HP{}
}

// w returns the perpendicular vector from the line to the relative velocity v.
func (vo VO) w(a agent.A, tau float64) vector.V {
	vs := s(vo.s, a, tau)
	vv := v(vo.v, a)
	return vector.Sub(
		vv,
		vs.L().L(vs.T(vv)),
	)
}

// l returns the line segment extending from the base of the truncated cone to
// the "left" side of the object line.
func (vo VO) l(a agent.A) vector.V { return vector.V{} }

// r returns the line segment extending from the base of the truncated cone to
// the "right" side of the object line.
func (vo VO) r(a agent.A) vector.V { return vector.V{} }

// v returns the relative velocity between the agent and the obstacle line.
func v(v vector.V, a agent.A) vector.V {
	return vector.Sub(a.V(), v)
}

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