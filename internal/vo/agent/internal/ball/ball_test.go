package ball

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/downflux/go-geometry/2d/hyperplane"
	"github.com/downflux/go-geometry/2d/vector"
	"github.com/downflux/go-geometry/epsilon"
	"github.com/downflux/go-orca/internal/vo/agent/internal"
	"github.com/downflux/go-orca/internal/vo/agent/internal/ball/domain"

	mock "github.com/downflux/go-orca/internal/agent/testdata/mock"
	reference "github.com/downflux/go-orca/internal/vo/agent/internal/mock"
)

var (
	_ vo.VO = &VO{}
)

// rn returns a random int between [-100, 100).
func rn() float64 { return rand.Float64()*200 - 100 }

// ra returns an agent with randomized dimensions.
func ra() mock.A {
	return *mock.New(
		mock.O{
			P: *vector.New(rn(), rn()),
			V: *vector.New(rn(), rn()),
			R: math.Abs(rn()),
		},
	)
}

func TestOrientation(t *testing.T) {
	a := *mock.New(mock.O{P: *vector.New(0, 0), V: *vector.New(0, 0), R: 1})
	b := *mock.New(mock.O{P: *vector.New(0, 5), V: *vector.New(1, -1), R: 2})

	t.Run("P", func(t *testing.T) {
		want := *vector.New(0, 5)
		if got := p(a, b, 1); !vector.Within(got, want) {
			t.Errorf("p() = %v, want = %v", got, want)
		}
		if got := p(b, a, 1); !vector.Within(got, vector.Scale(-1, want)) {
			t.Errorf("p() = %v, want = %v", got, vector.Scale(-1, want))
		}
	})
	t.Run("R", func(t *testing.T) {
		want := 3.0
		if got := r(a, b, 1); !epsilon.Within(got, want) {
			t.Errorf("r() = %v, want = %v", got, want)
		}
		if got := r(b, a, 1); !epsilon.Within(got, want) {
			t.Errorf("r() = %v, want = %v", got, want)
		}
	})
	t.Run("V", func(t *testing.T) {
		want := *vector.New(-1, 1)
		if got := v(a, b); !vector.Within(got, want) {
			t.Errorf("v() = %v, want = %v", got, want)
		}
		if got := v(b, a); !vector.Within(got, vector.Scale(-1, want)) {
			t.Errorf("v() = %v, want = %v", got, vector.Scale(-1, want))
		}
	})
	t.Run("W", func(t *testing.T) {
		want := *vector.New(-1, -4)
		if got := w(a, b, 1); !vector.Within(got, want) {
			t.Errorf("w() = %v, want = %v", got, want)
		}
		if got := w(b, a, 1); !vector.Within(got, vector.Scale(-1, want)) {
			t.Errorf("w() = %v, want = %v", got, vector.Scale(-1, want))
		}
	})
}

// TestVOReference asserts a simple RVO2 agent-agent setup will return correct
// values from hand calculations.
func TestVOReference(t *testing.T) {
	a := *mock.New(mock.O{P: *vector.New(0, 0), V: *vector.New(0, 0), R: 1})
	b := *mock.New(mock.O{P: *vector.New(0, 5), V: *vector.New(1, -1), R: 2})

	testConfigs := []struct {
		name   string
		tau    float64
		domain domain.D
		u      vector.V
		a      mock.A
		b      mock.A
		orca   hyperplane.HP
	}{
		{
			name:   "Simple",
			a:      a,
			b:      b,
			tau:    1,
			domain: domain.Circle,
			// These values were determined experimentally.
			u: *vector.New(0.2723931248910011, 1.0895724995640044),
			orca: *hyperplane.New(
				*vector.New(0.13619656244550055, 0.5447862497820022),
				*vector.New(-0.24253562503633297, -0.9701425001453319),
			),
		},
		{
			name:   "LargeTau",
			a:      a,
			b:      b,
			tau:    3,
			domain: domain.Left,
			// These values were determined experimentally.
			u: *vector.New(0.16000000000000003, 0.11999999999999988),
			orca: *hyperplane.New(
				*vector.New(0.08000000000000002, 0.05999999999999994),
				*vector.New(-0.8, -0.6),
			),
		},
		{
			name:   "InverseSimple",
			a:      b,
			b:      a,
			tau:    1,
			domain: domain.Circle,
			// These values were determined experimentally.
			u: vector.Scale(
				-1,
				*vector.New(0.2723931248910011, 1.0895724995640044),
			),
			orca: *hyperplane.New(
				*vector.New(0.8638034375544994, -1.5447862497820022),
				*vector.New(0.24253562503633297, 0.9701425001453319),
			),
		},
		{
			name:   "InverseLargeTau",
			a:      a,
			b:      b,
			tau:    3,
			domain: domain.Left,
			// These values were determined experimentally.
			u: *vector.New(0.16000000000000003, 0.11999999999999988),
			orca: *hyperplane.New(
				*vector.New(0.08000000000000002, 0.05999999999999994),
				*vector.New(-0.8, -0.6),
			),
		},
	}
	for _, c := range testConfigs {
		t.Run(c.name, func(t *testing.T) {
			t.Run("ORCA", func(t *testing.T) {
				got, err := reference.New(c.a, c.b, c.tau).ORCA()
				if err != nil {
					t.Fatalf("ORCA() returned error: %v", err)
				}
				if !hyperplane.Within(got, c.orca) {
					t.Errorf("ORCA() = %v, want = %v", got, c.orca)
				}
			})
		})
	}
}

// TestVOT tests that the tangent line t is being calculated correctly.
func TestVOT(t *testing.T) {
	testConfigs := []struct {
		name string
		a    mock.A
		b    mock.A
		tau  float64
		want vector.V
	}{
		{
			name: "345",
			a:    *mock.New(mock.O{P: *vector.New(0, 0), V: *vector.New(0, 0), R: 1}),
			b:    *mock.New(mock.O{P: *vector.New(0, 5), V: *vector.New(1, -1), R: 2}),
			tau:  1,
			want: *vector.New(-2.4, 3.2),
		},
		{
			name: "345LargeTau",
			a:    *mock.New(mock.O{P: *vector.New(0, 0), V: *vector.New(0, 0), R: 1}),
			b:    *mock.New(mock.O{P: *vector.New(0, 5), V: *vector.New(1, -1), R: 2}),
			tau:  3,
			want: vector.Scale(1.0/3, *vector.New(-2.4, 3.2)),
		},
		{
			name: "Inverse345",
			a:    *mock.New(mock.O{P: *vector.New(0, 5), V: *vector.New(1, -1), R: 2}),
			b:    *mock.New(mock.O{P: *vector.New(0, 0), V: *vector.New(0, 0), R: 1}),
			tau:  1,
			want: vector.Scale(
				-1,
				*vector.New(-2.4, 3.2),
			),
		},
		{
			name: "Inverse345LargeTau",
			a:    *mock.New(mock.O{P: *vector.New(0, 5), V: *vector.New(1, -1), R: 2}),
			b:    *mock.New(mock.O{P: *vector.New(0, 0), V: *vector.New(0, 0), R: 1}),
			tau:  3,
			want: vector.Scale(
				-1,
				vector.Scale(1.0/3, *vector.New(-2.4, 3.2)),
			),
		},
	}

	for _, c := range testConfigs {
		t.Run(c.name, func(t *testing.T) {
			v, err := New(c.a, c.b, c.tau)
			if err != nil {
				t.Fatalf("New() returned error: %v", err)
			}

			if got := v.t(); !vector.Within(got, c.want) {
				t.Errorf("t() = %v, want = %v", got, c.want)
			}
		})
	}
}

// TestVOConformance tests that agent-agent VOs will return u in the correct
// domain using the reference implementation as a sanity check.
func TestVOConformance(t *testing.T) {
	const nTests = 1000
	const delta = 1e-10

	type config struct {
		name string
		a    mock.A
		b    mock.A
		tau  float64
	}

	testConfigs := []config{
		{
			name: "SimpleCase",
			a:    *mock.New(mock.O{P: *vector.New(0, 0), V: *vector.New(0, 0), R: 1}),
			b:    *mock.New(mock.O{P: *vector.New(0, 5), V: *vector.New(1, -1), R: 2}),
			tau:  1,
		},
		{
			name: "SimpleCaseLargeTimeStep",
			a:    *mock.New(mock.O{P: *vector.New(0, 0), V: *vector.New(0, 0), R: 1}),
			b:    *mock.New(mock.O{P: *vector.New(0, 5), V: *vector.New(1, -1), R: 2}),
			tau:  3,
		},
		{
			name: "domain.Collision",
			a:    *mock.New(mock.O{P: *vector.New(0, 0), V: *vector.New(0, 0), R: 1}),
			b:    *mock.New(mock.O{P: *vector.New(0, 3), V: *vector.New(1, -1), R: 2}),
			tau:  1,
		},
	}

	for i := 0; i < nTests; i++ {
		testConfigs = append(testConfigs, config{
			name: fmt.Sprintf("Random-%v", i),
			a:    ra(),
			b:    ra(),
			// A simulation timestep scalar of 0 indicates the
			// simulation will never advance to the next snapshot,
			// which is a meaningless case (and will produce
			// boundary condition errors in our implementation).
			tau: math.Abs(rn()) + delta,
		})
	}

	for _, c := range testConfigs {
		t.Run(c.name, func(t *testing.T) {
			v, err := New(c.a, c.b, c.tau)
			if err != nil {
				t.Fatalf("New() returned a non-nil error: %v", err)
			}

			t.Run("ORCA", func(t *testing.T) {
				want, err := reference.New(c.a, c.b, float64(c.tau)).ORCA()
				if err != nil {
					t.Fatalf("ORCA() returned error: %v", err)
				}
				got, err := v.ORCA()
				if err != nil {
					t.Fatalf("ORCA() returned error: %v", err)
				}

				if !hyperplane.Within(got, want) {
					t.Errorf("ORCA() = %v, want = %v", got, want)
				}
			})
		})
	}
}

// BenchmarkORCA compares the relative performance of VO.check() bewteen the
// official RVO2 spec vs. the custom implementation provided.
func BenchmarkORCA(t *testing.B) {
	testConfigs := []struct {
		name        string
		constructor func(a, b mock.A) vo.VO
	}{
		{
			name:        "VOReference",
			constructor: func(a, b mock.A) vo.VO { return reference.New(a, b, 1) },
		},
		{
			name: "VO",
			constructor: func(a, b mock.A) vo.VO {
				v, _ := New(a, b, 1)
				return v
			},
		},
	}
	for _, c := range testConfigs {
		t.Run(c.name, func(t *testing.B) {
			v := c.constructor(ra(), ra())

			t.ResetTimer()
			for i := 0; i < t.N; i++ {
				v.ORCA()
			}
		})
	}
}

func TestP(t *testing.T) {
	o := mock.O{
		P: *vector.New(0, 1),
		V: *vector.New(1, 2),
		R: 10,
	}
	a := *mock.New(o)
	b := *mock.New(o)

	got := p(a, b, 1)
	if epsilon.Within(vector.Magnitude(got), 0) {
		t.Errorf("p() = %v, want a non-zero result", got)
	}
}
