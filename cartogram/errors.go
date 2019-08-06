package cartogram

import (
	"fmt"
)

// SpecVersionError indicates the cartogram version doesn't match voyager version
type SpecVersionError struct {
	ActualVersion, ExpectedVersion int
}

func (s SpecVersionError) Error() string {
	return fmt.Sprintf(
		"spec version mismatch: expected %d, got %d",
		s.ExpectedVersion,
		s.ActualVersion,
	)
}
