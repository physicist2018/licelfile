package licelformat

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- NewLicelProfile ---

func TestNewLicelProfile_ValidAnalog(t *testing.T) {
	line := " 1 0 1 16380 1 0000 7.50 00355.o 0 0 00 000 12 002001 0.500 BT0"
	pr, err := NewLicelProfile(line)
	require.NoError(t, err)

	assert.True(t, pr.Active)
	assert.False(t, pr.Photon)
	assert.Equal(t, 1, pr.LaserType)
	assert.Equal(t, 16380, pr.NDataPoints)
	assert.Equal(t, [3]int{1, 0, 0}, pr.Reserved)
	assert.Equal(t, 0, pr.HighVoltage)
	assert.Equal(t, 7.50, pr.BinWidth)
	assert.Equal(t, 355.0, pr.Wavelength)
	assert.Equal(t, "o", pr.Polarization)
	assert.Equal(t, 0, pr.BinShift)
	assert.Equal(t, 0, pr.DecBinShift)
	assert.Equal(t, 12, pr.AdcBits)
	assert.Equal(t, 2001, pr.NShots)
	assert.Equal(t, 0.5, pr.DiscrLevel)
	assert.Equal(t, "BT", pr.DeviceID)
	assert.Equal(t, 0, pr.NCrate)
}

func TestNewLicelProfile_ValidPhoton(t *testing.T) {
	line := " 1 1 1 16380 1 0000 7.50 00355.o 0 0 00 000 00 002001 3.1746 BC0"
	pr, err := NewLicelProfile(line)
	require.NoError(t, err)

	assert.True(t, pr.Active)
	assert.True(t, pr.Photon)
	assert.Equal(t, 1, pr.LaserType)
	assert.Equal(t, 16380, pr.NDataPoints)
	assert.Equal(t, 0, pr.AdcBits)
	assert.Equal(t, 2001, pr.NShots)
	assert.InDelta(t, 3.1746, pr.DiscrLevel, 1e-6)
	assert.Equal(t, "BC", pr.DeviceID)
	assert.Equal(t, 0, pr.NCrate)
}

func TestNewLicelProfile_TooFewFields(t *testing.T) {
	_, err := NewLicelProfile(" 1 0 1")
	assert.Error(t, err)
}

func TestNewLicelProfile_InvalidWavelength(t *testing.T) {
	line := " 1 0 1 16380 1 0000 7.50 abc.o 0 0 00 000 12 002001 0.500 BT0"
	_, err := NewLicelProfile(line)
	assert.Error(t, err)
}

func TestNewLicelProfile_InvalidNumbers(t *testing.T) {
	line := " X 0 1 16380 1 0000 7.50 00355.o 0 0 00 000 12 002001 0.500 BT0"
	_, err := NewLicelProfile(line)
	assert.Error(t, err)
}

// --- Metadata ---

func TestLicelProfile_Metadata_Analog(t *testing.T) {
	pr := LicelProfile{
		Active:       true,
		Photon:       false,
		LaserType:    1,
		NDataPoints:  16380,
		Reserved:     [3]int{1, 0, 0},
		HighVoltage:  0,
		BinWidth:     7.50,
		Wavelength:   355,
		Polarization: "o",
		BinShift:     0,
		DecBinShift:  0,
		AdcBits:      12,
		NShots:       2001,
		DiscrLevel:   0.5,
		DeviceID:     "BT",
		NCrate:       0,
	}

	s := pr.Metadata()
	assert.Contains(t, s, "0.500")
	assert.Contains(t, s, "BT0")
	assert.Contains(t, s, "355.o")
	assert.Contains(t, s, "\r\n")
}

func TestLicelProfile_Metadata_Photon(t *testing.T) {
	pr := LicelProfile{
		Active:       true,
		Photon:       true,
		LaserType:    1,
		NDataPoints:  16380,
		Reserved:     [3]int{1, 0, 0},
		HighVoltage:  0,
		BinWidth:     7.50,
		Wavelength:   355,
		Polarization: "o",
		BinShift:     0,
		DecBinShift:  0,
		AdcBits:      0,
		NShots:       2001,
		DiscrLevel:   3.1746,
		DeviceID:     "BC",
		NCrate:       0,
	}

	s := pr.Metadata()
	assert.Contains(t, s, "3.1746")
	assert.Contains(t, s, "BC0")
}

// --- ProfileRaw ---

func TestLicelProfile_ProfileRaw(t *testing.T) {
	pr := LicelProfile{Photon: false, AdcBits: 12, NShots: 2001, DiscrLevel: 0.5, Data: []float64{17.27, 17.28, 17.29}}
	data, err := pr.ProfileRaw()
	require.NoError(t, err)
	assert.Len(t, data, 12)
}

// --- btoi ---

func TestBtoi(t *testing.T) {
	assert.Equal(t, 1, btoi(true))
	assert.Equal(t, 0, btoi(false))
}
