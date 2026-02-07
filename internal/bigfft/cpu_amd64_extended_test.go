//go:build amd64

package bigfft

import (
	"testing"
)

// ─────────────────────────────────────────────────────────────────────────────
// CPU Feature Extended Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestHasBMI2(t *testing.T) {
	t.Parallel()
	hasBMI2 := HasBMI2()
	features := GetCPUFeatures()
	if hasBMI2 != features.BMI2 {
		t.Errorf("HasBMI2() = %v, but GetCPUFeatures().BMI2 = %v", hasBMI2, features.BMI2)
	}
	t.Logf("BMI2 available: %v", hasBMI2)
}

func TestHasADX(t *testing.T) {
	t.Parallel()
	hasADX := HasADX()
	features := GetCPUFeatures()
	if hasADX != features.ADX {
		t.Errorf("HasADX() = %v, but GetCPUFeatures().ADX = %v", hasADX, features.ADX)
	}
	t.Logf("ADX available: %v", hasADX)
}

func TestCPUFeaturesString(t *testing.T) {
	t.Parallel()
	features := GetCPUFeatures()
	str := features.String()
	if str == "" {
		t.Error("CPUFeatures.String() returned empty string")
	}
	t.Logf("CPU Features string: %s", str)
}

func TestSIMDLevelString(t *testing.T) {
	t.Parallel()
	levels := []SIMDLevel{SIMDNone, SIMDAVX2, SIMDAVX512, SIMDLevel(99)} // Including unknown
	for _, level := range levels {
		t.Run(level.String(), func(t *testing.T) {
			t.Parallel()
			str := level.String()
			if str == "" {
				t.Errorf("SIMDLevel(%d).String() returned empty string", level)
			}
			t.Logf("SIMDLevel %d: %s", level, str)
		})
	}
}
