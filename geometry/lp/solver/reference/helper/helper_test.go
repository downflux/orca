package helper

import (
	"fmt"
	"math"
	"testing"

	"github.com/downflux/orca/geometry/plane"
	"github.com/downflux/orca/geometry/vector"

	agent "github.com/downflux/orca/agent/reference"
)

const (
	tolerance = 1e-7
)

func TestAdd(t *testing.T) {
	a := *agent.New(agent.O{T: *vector.New(0, 0.9), S: 1})

	type config struct {
		name       string
		h          H
		constraint plane.HP
		success    bool
		want       vector.V
	}

	var testConfigs []config

	// The constraint lies tangent to the maximal speed circle, with the
	// preferred velocity T lying inside the constraint.
	//
	// N.B.: The half-plane HP(P, N) has permissible values oriented in the
	// same direction as N.
	for _, o := range []struct {
		name string
		p    vector.V
	}{
		{name: "0", p: *vector.New(1, 0)},
		{name: "45", p: vector.Unit(*vector.New(1, 1))},
		{name: "90", p: *vector.New(0, 1)},
		{name: "135", p: vector.Unit(*vector.New(-1, 1))},
		{name: "180", p: *vector.New(-1, 0)},
		{name: "225", p: vector.Unit(*vector.New(-1, -1))},
		{name: "270", p: *vector.New(-1, 0)},
		{name: "315", p: vector.Unit(*vector.New(1, -1))},
	} {
		// The constraint lies tangent to the maximal speed circle, with the
		// preferred velocity T lying inside the constraint.
		//
		// N.B.: The half-plane HP(P, N) has permissible values oriented in the
		// same direction as N.
		testConfigs = append(testConfigs, config{
			name:       fmt.Sprintf("SingleViableConstraint/Tangent/In/%v", o.name),
			h:          H{a: a},
			constraint: *plane.New(o.p, vector.Scale(-1, o.p)),
			success:    true,
			want:       o.p,
		})
		// The constraint lies tangent to the maximal speed circle, with
		// the preferred velocity T lying outside the constraint.
		//
		// Since we're still looking for the minimal distance between
		// the preferred velocity and the constraint, the calculated
		// velocity vector should still be the same.
		testConfigs = append(testConfigs, config{
			name:       fmt.Sprintf("SingleViableConstraint/Tangent/Out/%v", o.name),
			h:          H{a: a},
			constraint: *plane.New(o.p, o.p),
			success:    true,
			want:       o.p,
		})
	}

	// It doesn't matter if the constraint points into or away from the
	// circle -- if there is no intersection between the two, then the two
	// constraints invalidate one another, and we will achieve an infeasible
	// solution.
	testConfigs = append(
		testConfigs,
		config{
			name:       "SingleNotViableConstraint/In",
			h:          H{a: a},
			constraint: *plane.New(*vector.New(a.S()+1, 0), *vector.New(-1, 0)),
			success:    false,
			want:       vector.V{},
		},
		config{
			name:       "SingleNotViableConstraint/Out",
			h:          H{a: a},
			constraint: *plane.New(*vector.New(a.S()+1, 0), *vector.New(1, 0)),
			success:    false,
			want:       vector.V{},
		},
	)

	// Given that the preferred velocity is pointing straight up, and a
	// viable constraint intersects the circle horizontally, the optimal
	// solution is just the Y-intersect of the constraint line.
	//
	// Also tests P-invariance -- that is, if our choice of P does not
	// matter as long as it lies on the constraint line.
	testConfigs = append(
		testConfigs,
		config{
			name:       "SingleViableConstraint/Intersection/CenterP",
			h:          H{a: a},
			constraint: *plane.New(*vector.New(0, 0.5), *vector.New(0, -1)),
			success:    true,
			want:       *vector.New(0, 0.5),
		},
		config{
			name: "SingleViableConstraint/Intersection/LeftP",
			h:    H{a: a},
			// P is tangent to the circular constraint in top left quadrant.
			constraint: *plane.New(*vector.New(a.S()*-math.Sqrt(3)/2, 0.5), *vector.New(0, -1)),
			success:    true,
			want:       *vector.New(0, 0.5),
		},
		config{
			name:       "SingleViableConstraint/Intersection/RightP",
			h:          H{a: a},
			constraint: *plane.New(*vector.New(a.S()*math.Sqrt(3)/2, 0.5), *vector.New(0, -1)),
			success:    true,
			want:       *vector.New(0, 0.5),
		},
	)

	testConfigs = append(
		testConfigs,
		config{
			name: "SimpleRefinedConstraint/Improve",
			h: H{
				a: a,
				cs: []plane.HP{
					*plane.New(*vector.New(0, 1), *vector.New(0, -1)),
				},
			},
			constraint: *plane.New(*vector.New(0, 0.5), *vector.New(0, -1)),
			success:    true,
			want:       *vector.New(0, 0.5),
		},
		// When the new constraint generates a vector less optimal than
		// the ones generated previously, this implementation returns an
		// infeasible error. This feasibility check is done in
		// linearProgram2 in the RVO2 implementation.
		config{
			name: "SimpleRefinedConstraint/NoImprove",
			h: H{
				a: a,
				cs: []plane.HP{
					*plane.New(*vector.New(0, 0.5), *vector.New(0, -1)),
				},
			},
			constraint: *plane.New(*vector.New(0, 1), *vector.New(0, -1)),
			success:    false,
			want:       *vector.New(0, 0),
		},
	)

	for _, c := range testConfigs {
		t.Run(c.name, func(t *testing.T) {
			if got, success := c.h.Add(c.constraint); success != c.success || !vector.Within(got, c.want, tolerance) {
				t.Fatalf("Add() = %v, %v, want = %v, %v", got, success, c.want, c.success)
			}
		})
	}
}