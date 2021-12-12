package orca

import (
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"testing"

	"github.com/downflux/go-geometry/2d/hypersphere"
	"github.com/downflux/go-geometry/nd/vector"
	"github.com/downflux/go-kd/kd"
	"github.com/downflux/go-kd/point"
	"github.com/downflux/go-orca/agent"
	"github.com/google/go-cmp/cmp"

	v2d "github.com/downflux/go-geometry/2d/vector"
	mock "github.com/downflux/go-orca/internal/agent/testdata/mock"
)

var _ point.P = P{}

// r calculates a "good enough" neighbor vision radius to be passed into
// orca.Step.
func r(s float64, tau float64) float64 { return tau * s }

func rn() float64 { return rand.Float64()*200 - 100 }
func rv() v2d.V   { return *v2d.New(rn(), rn()) }
func ra() mock.A {
	return *mock.New(
		mock.O{
			P: rv(),
			V: v2d.Scale(rand.Float64()*.5, v2d.Unit(rv())),
			T: rv(),
			// Ensure the agent's target vector is inside the
			// bounding circle.
			S: rn() + 100,
		},
	)
}
func rt(n int) *kd.T {
	// Generating large number of points in tests will mess with data
	// collection figures. We should ignore these allocs.
	runtime.MemProfileRate = 0
	defer func() { runtime.MemProfileRate = 512 * 1024 }()

	ps := make([]point.P, 0, n)
	for i := 0; i < n; i++ {
		ps = append(ps, *New(ra()))
	}
	t, _ := kd.New(ps)
	return t
}

func TestStep(t *testing.T) {
	type config struct {
		name   string
		agents []agent.A
		tau    float64
		f      func(a agent.A) bool

		want []Mutation
	}

	testConfigs := []config{
		func() config {
			a := mock.New(
				mock.O{
					P: *v2d.New(1, 2),
					V: *v2d.New(2, 3),
					T: *v2d.New(3, 4),
					S: 10,
				},
			)

			return config{
				name:   "PointerEqualityComparison",
				agents: []agent.A{a},
				tau:    1e-2,
				f:      func(agent.A) bool { return true },
				want: []Mutation{
					Mutation{
						A: a,
						V: vector.V(a.T()),
					},
				},
			}
		}(),
	}

	for _, c := range testConfigs {
		t.Run(c.name, func(t *testing.T) {
			var ps []point.P
			for _, a := range c.agents {
				ps = append(ps, *New(a))
			}

			tr, err := kd.New(ps)
			if err != nil {
				t.Fatalf("New() = _, %v, want = _, %v", err, nil)
			}

			s := math.Inf(-1)
			for _, a := range kd.Data(tr) {
				s = math.Max(s, a.(P).a.S())
			}

			got, err := Step(O{
				T:   tr,
				Tau: c.tau,
				F:   c.f,
				R: func(tau float64) float64 { return r(s, tau) },
			})
			if err != nil {
				t.Errorf("Step() = _, %v, want = _, %v", got, nil)
			}
			if diff := cmp.Diff(
				c.want,
				got,
				cmp.AllowUnexported(
					mock.A{},
					hypersphere.C{},
				),
			); diff != "" {
				t.Errorf("Step() mismatch (-want +got):\n%v", diff)
			}
		})
	}
}

// TODO(minkezhang): Change tau value to coincide with existing altered tau
// assumptions.
func BenchmarkStep(b *testing.B) {
	type config struct {
		name string
		t    *kd.T
	}

	testConfigs := []config{}
	for i := 0; i < 7; i++ {
		n := int(math.Pow(10, float64(i)))
		testConfigs = append(testConfigs, config{
			name: fmt.Sprintf("N=%v", n),
			t:    rt(n),
		})
	}

	for _, c := range testConfigs {
		s := math.Inf(-1)
		for _, a := range kd.Data(c.t) {
			s = math.Max(s, a.(P).a.S())
		}

		b.Run(c.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := Step(O{
					T:   c.t,
					Tau: 1e-2,
					F:   func(a agent.A) bool { return true },
					R: func(tau float64) float64 { return r(s, tau) },
				}); err != nil {
					b.Errorf("Step() = _, %v, want = _, %v", err, nil)
				}
			}
		})
	}
}
