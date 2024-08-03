package gitlab

import (
	"context"
)

type Validator struct{}

func (r *Validator) Validate(_ context.Context, _, _, _ string) error {
	return nil
}
