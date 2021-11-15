package solver

import (
	"testing"

	"github.com/downflux/go-geometry/nd/hyperplane"
	"github.com/downflux/go-geometry/nd/vector"
	"github.com/downflux/go-orca/internal/solver/constraint"

	v2d "github.com/downflux/go-geometry/2d/vector"
)

func TestOptimize(t *testing.T) {
	type config struct {
		name    string
		cs      []constraint.C
		v       v2d.V
		success bool
		want    v2d.V
	}

	testConfigs := []config{
		{
			name:    "NoConstraints",
			cs:      nil,
			v:       *v2d.New(1, 2),
			success: true,
			want:    *v2d.New(1, 2),
		},

		// The target minimization vector is already within the single
		// constraint.
		{
			name: "SingleConstraint/WithinConstraint",
			cs: []constraint.C{
				*constraint.New(
					*hyperplane.New(
						*vector.New(0, 1),
						*vector.New(0, 1),
					),
				),
			},
			v:       *v2d.New(0, 2),
			success: true,
			want:    *v2d.New(0, 2),
		},

		// The target minimization vector is outside the constraint, and
		// so must be recalculated.
		{
			name: "SingleConstraint/OutsideConstraint",
			cs: []constraint.C{
				*constraint.New(
					*hyperplane.New(
						*vector.New(0, 1),
						*vector.New(0, 1),
					),
				),
			},
			v:       *v2d.New(0, -1),
			success: true,
			want:    *v2d.New(0, 1),
		},
	}

	testConfigs = append(
		testConfigs,
		func() []config {
			c := *constraint.New(
				*hyperplane.New(
					*vector.New(0, 1),
					*vector.New(0, 1),
				),
			)

			d := *constraint.New(
				*hyperplane.New(
					*vector.New(0, 2),
					*vector.New(0, 1),
				),
			)

			return []config{
				{
					name:    "ParalleConstraints/SuccessivelyConstrain",
					cs:      []constraint.C{c, d},
					v:       *v2d.New(0, -1),
					success: true,
					want:    *v2d.New(0, 2),
				},

				// Ensure that relaxing a parallel constraint
				// will not cause the function to throw an
				// infeasibility error.
				{
					name:    "ParalleConstraints/RelaxConstraintStillFeasible",
					cs:      []constraint.C{d, c},
					v:       *v2d.New(0, -1),
					success: true,
					want:    *v2d.New(0, 2),
				},
			}
		}()...,
	)

	for _, c := range testConfigs {
		t.Run(c.name, func(t *testing.T) {
			if got, ok := optimize(c.v, c.cs); ok != c.success || !v2d.Within(c.want, got) {
				t.Errorf("optimize() = %v, %v, want = %v, %v", got, ok, c.want, c.success)
			}
		})
	}
}
