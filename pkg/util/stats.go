package util

import (
	"fmt"
	"math"
	"slices"
)

// Float64Data is a slice of float64 values with embedded statistical methods.
type Float64Data []float64

// Sum calculates the total sum of all values in the slice.
//
// Returns:
//   - float64: The sum of the values, or NaN if the slice is empty.
func (f Float64Data) Sum() float64 {
	if len(f) == 0 {
		return math.NaN()
	}
	var sum float64
	for _, v := range f {
		sum += v
	}
	return sum
}

// Mean calculates the average (arithmetic mean) of the values in the slice.
//
// Returns:
//   - float64: The mean value, or NaN if the slice is empty.
func (f Float64Data) Mean() float64 {
	if len(f) == 0 {
		return math.NaN()
	}
	return f.Sum() / float64(len(f))
}

// Min returns the minimum value in the slice.
//
// Returns:
//   - float64: The minimum value, or NaN if the slice is empty.
func (f Float64Data) Min() float64 {
	if len(f) == 0 {
		return math.NaN()
	}
	value := f[0]
	for _, v := range f[1:] {
		if v < value {
			value = v
		}
	}
	return value
}

// Max returns the maximum value in the slice.
//
// Returns:
//   - float64: The maximum value, or NaN if the slice is empty.
func (f Float64Data) Max() float64 {
	if len(f) == 0 {
		return math.NaN()
	}
	value := f[0]
	for _, v := range f[1:] {
		if v > value {
			value = v
		}
	}
	return value
}

// Variance calculates the variance of the values in the slice.
//
// Returns:
//   - float64: The variance, or NaN if the slice is empty.
func (f Float64Data) Variance() float64 {
	if len(f) == 0 {
		return math.NaN()
	}

	mean := f.Mean()
	var variance float64

	for _, v := range f {
		deviation := v - mean
		variance += deviation * deviation
	}

	return variance / float64(len(f))
}

// StdDev calculates the standard deviation of the values in the slice.
//
// Returns:
//   - float64: The standard deviation, or NaN if the slice is empty.
func (f Float64Data) StdDev() float64 {
	variance := f.Variance()
	if math.IsNaN(variance) {
		return math.NaN()
	}
	return math.Sqrt(variance)
}

// Percentile calculates the given percentile of the data using linear interpolation.
//
// Parameters:
//   - percent: A float64 representing the desired percentile (0 to 100).
//
// Returns:
//   - float64: The computed percentile value.
//   - error:   An error if the slice is empty or if the provided percentile is out of range.
func (f Float64Data) Percentile(percent float64) (percentile float64, err error) {
	if len(f) == 0 {
		return math.NaN(), fmt.Errorf("no data")
	}
	if percent < 0 || percent > 100 {
		return math.NaN(), fmt.Errorf("percent out of range")
	}
	if len(f) == 1 {
		return f[0], nil
	}

	// Create a sorted copy of the data
	sorted := slices.Clone(f)
	slices.Sort(sorted)

	// Compute the position in the sorted slice
	pos := (percent / 100) * float64(len(sorted)-1)
	lowerIndex := int(math.Floor(pos))
	upperIndex := int(math.Ceil(pos))

	// If the position is an integer, return the value directly
	if lowerIndex == upperIndex {
		return sorted[lowerIndex], nil
	}

	// Interpolate between the two closest values
	weight := pos - float64(lowerIndex)
	return sorted[lowerIndex]*(1-weight) + sorted[upperIndex]*weight, nil
}
