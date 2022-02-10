package ball

import (
	"math"
	"math/rand"

	"github.com/downflux/go-geometry/2d/hyperplane"
	"github.com/downflux/go-geometry/2d/vector"
	"github.com/downflux/go-orca/agent"
	"github.com/downflux/go-orca/internal/vo/agent/internal/ball/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	mock "github.com/downflux/go-orca/internal/agent/testdata/mock"
)

const (
	minTau = 1e-3
)

// Reference implements the official RVO2 spec. See
// https://gamma.cs.unc.edu/RVO2/ for more information.
type Reference struct {
	a   mock.A
	b   mock.A
	tau float64
}

func New(a mock.A, b mock.A, tau float64) *Reference {
	return &Reference{
		a:   a,
		b:   b,
		tau: tau,
	}
}

func (vo Reference) ORCA() (hyperplane.HP, error) {
	u, err := vo.u()
	if err != nil {
		return hyperplane.HP{}, err
	}

	var n vector.V

	switch d := vo.check(); d {
	case domain.Collision:
		fallthrough
	case domain.Circle:
		tw := vo.w()
		if d == domain.Collision {
			tw = w(vo.a, vo.b, minTau)
		}
		n = vector.Unit(tw)
	case domain.Right:
		fallthrough
	case domain.Left:
		l := vo.l()
		// Rotate anti-clockwise by π / 2 towards the "outside" of the
		// VO cone.
		n = vector.Unit(*vector.New(-l.Y(), l.X()))
	default:
		return hyperplane.HP{}, status.Errorf(codes.Internal, "invalid domain %v", d)
	}
	return *hyperplane.New(
		vector.Add(vo.a.V(), vector.Scale(0.5, u)),
		n,
	), nil
}

func (vo Reference) n() (vector.V, error) {
	orca, err := vo.ORCA()
	if err != nil {
		return vector.V{}, err
	}
	return orca.N(), nil
}

func (vo Reference) u() (vector.V, error) {
	switch d := vo.check(); d {
	case domain.Collision:
		fallthrough
	case domain.Circle:
		tr := vo.r()
		tw := vo.w()

		if d == domain.Collision {
			tr = r(vo.a, vo.b, minTau)
			tw = w(vo.a, vo.b, minTau)
		}

		return vector.Scale(tr-vector.Magnitude(tw), vector.Unit(tw)), nil
	case domain.Right:
		fallthrough
	case domain.Left:
		l := vo.l()
		return vector.Sub(vector.Scale(vector.Dot(vo.v(), l), l), vo.v()), nil
	default:
		return vector.V{}, status.Errorf(codes.Internal, "invalid domain %v", d)
	}
}
func (vo Reference) r() float64  { return r(vo.a, vo.b, vo.tau) }
func (vo Reference) p() vector.V { return p(vo.a, vo.b, vo.tau) }
func (vo Reference) w() vector.V { return w(vo.a, vo.b, vo.tau) }
func (vo Reference) v() vector.V { return v(vo.a, vo.b) }

// t calculates the unnormalized vector of the tangent line from the base of p
// to the edge of the truncation circle. This corresponds to line.direction in
// the RVO2 implementation. Returns the left or right vector based on the
// projected side of u onto the VO.
func (vo Reference) t() vector.V {
	tp := p(vo.a, vo.b, 1)
	tr := r(vo.a, vo.b, 1)
	l := math.Sqrt(vector.SquaredMagnitude(tp) - math.Pow(tr, 2))
	return vector.Scale(1/vector.SquaredMagnitude(tp), *vector.New(
		tp.X()*l-tp.Y()*tr,
		tp.X()*tr+tp.Y()*l,
	))
}

// l calculates the domain-aware leg of the tangent line.
func (vo Reference) l() vector.V {
	t := vo.t()
	if vo.check() == domain.Right {
		tp := p(vo.a, vo.b, 1)
		tr := r(vo.a, vo.b, 1)
		l := math.Sqrt(vector.SquaredMagnitude(tp) - math.Pow(tr, 2))
		t = vector.Scale(-1/vector.SquaredMagnitude(tp), *vector.New(
			tp.X()*l+tp.Y()*tr,
			-tp.X()*tr+tp.Y()*l,
		))
	}
	return t
}

func (vo Reference) check() domain.D {
	if vector.SquaredMagnitude(vo.p()) <= math.Pow(vo.r(), 2) {
		return domain.Collision
	}

	wp := vector.Dot(vo.w(), vo.p())
	if wp < 0 && math.Pow(wp, 2) > vector.SquaredMagnitude(vo.w())*math.Pow(vo.r(), 2) {
		return domain.Circle
	}

	if vector.Determinant(vo.p(), vo.w()) > 0 {
		return domain.Left
	}

	return domain.Right
}

// v is a utility function calculating the relative velocities between two
// agents.
//
// Note that the relative velocity here is oriented from b.V to a.V.
func v(a agent.A, b agent.A) vector.V { return vector.Sub(a.V(), b.V()) }

// r is a utility function calculating the radius of the truncated VO circle.
func r(a agent.A, b agent.A, tau float64) float64 { return (a.R() + b.R()) / tau }

// p is a utility function calculating the relative position vector between two
// agents, scaled to the center of the truncated circle.
//
// Note the relative position is oriented from a.P to b.P.
func p(a agent.A, b agent.A, tau float64) vector.V {
	// Check for the degenerate case -- if two agents are too close, return
	// some sensical non-zero answer.
	if vector.Within(a.P(), b.P()) {
		return vector.Scale(
			1/tau,
			vector.Unit(
				*vector.New(
					rand.Float64(),
					rand.Float64(),
				),
			),
		)
	}
	return vector.Scale(1/tau, vector.Sub(b.P(), a.P()))
}

// w is a utility function calculating the relative velocity between a and b,
// centered on the truncation circle.
func w(a agent.A, b agent.A, tau float64) vector.V {
	return vector.Sub(v(a, b), p(a, b, tau))
}
