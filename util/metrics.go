package util

import (
	"math"

	"github.com/obicons/avis/entities"
)

func Distance(p1, p2 entities.Position) float64 {
	return math.Sqrt(
		math.Pow(p1.X-p2.X, 2) +
			math.Pow(p1.Y-p2.Y, 2) +
			math.Pow(p1.Z-p2.Z, 2),
	)
}
