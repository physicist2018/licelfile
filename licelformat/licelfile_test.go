package licelformat

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- str2Bool / str2Int / str2Float ---

func TestStr2Bool(t *testing.T) {
	tests := []struct {
		input  string
		expect bool
		err    bool
	}{
		{"1", true, false},
		{"0", false, false},
		{"t", true, false},
		{"f", false, false},
		{"true", true, false},
		{"false", false, false},
		{"", false, true},
		{"invalid", false, true},
	}
	for _, tt := range tests {
		v, err := str2Bool(tt.input)
		if tt.err {
			assert.Error(t, err, "input: %q", tt.input)
		} else {
			require.NoError(t, err, "input: %q", tt.input)
			assert.Equal(t, tt.expect, v, "input: %q", tt.input)
		}
	}
}

func TestStr2Int(t *testing.T) {
	v, err := str2Int("42")
	require.NoError(t, err)
	assert.Equal(t, 42, v)

	_, err = str2Int("abc")
	assert.Error(t, err)
}

func TestStr2Float(t *testing.T) {
	v, err := str2Float("3.14")
	require.NoError(t, err)
	assert.Equal(t, 3.14, v)

	_, err = str2Float("abc")
	assert.Error(t, err)
}

// --- bytesToFloat64Array ---

func TestBytesToFloat64Array(t *testing.T) {
	// raw bytes: int32 1, 2, 3 (little-endian)
	raw := []byte{1, 0, 0, 0, 2, 0, 0, 0, 3, 0, 0, 0}
	result := bytesToFloat64Array(raw)
	assert.Equal(t, []float64{1, 2, 3}, result)
}

// --- scaleFactor ---

func TestScaleFactor_Photon(t *testing.T) {
	pr := LicelProfile{Photon: true, NShots: 2000}
	scale := pr.scaleFactor()
	assert.InDelta(t, 1.0/(2000.0*0.05), scale, 1e-10)
}

func TestScaleFactor_Analog(t *testing.T) {
	pr := LicelProfile{Photon: false, AdcBits: 12, NShots: 2001, DiscrLevel: 0.5}
	adcScale := 1 << pr.AdcBits
	expected := 0.5 * 1000.0 / float64(adcScale*2001)
	assert.InDelta(t, expected, pr.scaleFactor(), 1e-10)
}

// --- parseTime ---

func TestParseTime(t *testing.T) {
	ts, err := parseTime("10/02/2020 19:22:35")
	require.NoError(t, err)
	assert.Equal(t, 2020, ts.Year())
	assert.Equal(t, time.February, ts.Month())
	assert.Equal(t, 10, ts.Day())
	assert.Equal(t, 19, ts.Hour())
	assert.Equal(t, 22, ts.Minute())
	assert.Equal(t, 35, ts.Second())
	assert.Equal(t, time.Local, ts.Location())
}

func TestParseTime_Invalid(t *testing.T) {
	_, err := parseTime("not a date")
	assert.Error(t, err)
}

func TestParseTime_WrongOrder(t *testing.T) {
	// MM/DD/YYYY instead of DD/MM/YYYY
	_, err := parseTime("30/02/2020 19:22:35")
	assert.Error(t, err)
}

// --- Save → round-trip test ---

func TestLicelFile_Save_Roundtrip(t *testing.T) {
	// Load a real LICEL file
	testFile := filepath.Join("..", "testdata", "b2021019.223500")

	lf, err := LoadLicelFile(testFile)
	require.NoError(t, err, "loading test file")

	// Save to temporary directory
	tmpDir := t.TempDir()
	savedPath := filepath.Join(tmpDir, "roundtrip.dat")

	err = lf.Save(savedPath)
	require.NoError(t, err, "saving file")

	// Read back
	lf2, err := LoadLicelFile(savedPath)
	require.NoError(t, err, "reloading saved file")

	assert.Equal(t, lf.MeasurementSite, lf2.MeasurementSite)
	assert.Equal(t, lf.MeasurementStartTime, lf2.MeasurementStartTime)
	assert.Equal(t, lf.MeasurementStopTime, lf2.MeasurementStopTime)
	assert.Equal(t, lf.AltitudeAboveSeaLevel, lf2.AltitudeAboveSeaLevel)
	assert.Equal(t, lf.Longitude, lf2.Longitude)
	assert.Equal(t, lf.Latitude, lf2.Latitude)
	assert.Equal(t, lf.Zenith, lf2.Zenith)
	assert.Equal(t, lf.Laser1NShots, lf2.Laser1NShots)
	assert.Equal(t, lf.Laser1Freq, lf2.Laser1Freq)
	assert.Equal(t, lf.Laser2NShots, lf2.Laser2NShots)
	assert.Equal(t, lf.Laser2Freq, lf2.Laser2Freq)
	assert.Equal(t, lf.NDatasets, lf2.NDatasets)
	assert.Equal(t, lf.Laser3NShots, lf2.Laser3NShots)
	assert.Equal(t, lf.Laser3Freq, lf2.Laser3Freq)

	require.Len(t, lf2.Profiles, len(lf.Profiles))
	for i := range lf.Profiles {
		assert.Equal(t, lf.Profiles[i].Active, lf2.Profiles[i].Active, "profile %d Active", i)
		assert.Equal(t, lf.Profiles[i].Photon, lf2.Profiles[i].Photon, "profile %d Photon", i)
		assert.Equal(t, lf.Profiles[i].LaserType, lf2.Profiles[i].LaserType, "profile %d LaserType", i)
		assert.Equal(t, lf.Profiles[i].NDataPoints, lf2.Profiles[i].NDataPoints, "profile %d NDataPoints", i)
		assert.Equal(t, lf.Profiles[i].Reserved, lf2.Profiles[i].Reserved, "profile %d Reserved", i)
		assert.Equal(t, lf.Profiles[i].HighVoltage, lf2.Profiles[i].HighVoltage, "profile %d HighVoltage", i)
		assert.Equal(t, lf.Profiles[i].BinWidth, lf2.Profiles[i].BinWidth, "profile %d BinWidth", i)
		assert.Equal(t, lf.Profiles[i].Wavelength, lf2.Profiles[i].Wavelength, "profile %d Wavelength", i)
		assert.Equal(t, lf.Profiles[i].Polarization, lf2.Profiles[i].Polarization, "profile %d Polarization", i)
		assert.Equal(t, lf.Profiles[i].BinShift, lf2.Profiles[i].BinShift, "profile %d BinShift", i)
		assert.Equal(t, lf.Profiles[i].DecBinShift, lf2.Profiles[i].DecBinShift, "profile %d DecBinShift", i)
		assert.Equal(t, lf.Profiles[i].AdcBits, lf2.Profiles[i].AdcBits, "profile %d AdcBits", i)
		assert.Equal(t, lf.Profiles[i].NShots, lf2.Profiles[i].NShots, "profile %d NShots", i)
		assert.InDelta(t, lf.Profiles[i].DiscrLevel, lf2.Profiles[i].DiscrLevel, 1e-6, "profile %d DiscrLevel", i)
		assert.Equal(t, lf.Profiles[i].DeviceID, lf2.Profiles[i].DeviceID, "profile %d DeviceID", i)
		assert.Equal(t, lf.Profiles[i].NCrate, lf2.Profiles[i].NCrate, "profile %d NCrate", i)

		require.Len(t, lf2.Profiles[i].Data, len(lf.Profiles[i].Data), "profile %d Data length", i)
		for j := range lf.Profiles[i].Data {
			assert.InDelta(t, lf.Profiles[i].Data[j], lf2.Profiles[i].Data[j], 0.1, "profile %d data[%d]", i, j)
		}
	}
}

// --- SetMaxDist ---

func TestLicelFile_SetMaxDist(t *testing.T) {
	lf := LicelFile{
		Profiles: LicelProfilesList{
			{BinWidth: 7.5, NDataPoints: 100, Data: make([]float64, 100)},
			{BinWidth: 7.5, NDataPoints: 200, Data: make([]float64, 200)},
		},
	}

	err := lf.SetMaxDist(375) // idx = 50 for both
	require.NoError(t, err)
	assert.Equal(t, 50, lf.Profiles[0].NDataPoints)
	assert.Len(t, lf.Profiles[0].Data, 50)
	assert.Equal(t, 50, lf.Profiles[1].NDataPoints)
	assert.Len(t, lf.Profiles[1].Data, 50)
}

func TestLicelFile_SetMaxDist_Error(t *testing.T) {
	lf := LicelFile{
		Profiles: LicelProfilesList{
			{BinWidth: 7.5, NDataPoints: 100, Data: make([]float64, 100)},
			{BinWidth: 7.5, NDataPoints: 10, Data: make([]float64, 10)}, // will fail
		},
	}

	err := lf.SetMaxDist(375) // idx = 50 > 10 for profile 1
	assert.Error(t, err)
}

// --- SelectProfile ---

func TestLicelFile_SelectProfile(t *testing.T) {
	lf := LicelFile{
		Profiles: LicelProfilesList{
			{DeviceID: "BT", Photon: false, Wavelength: 355, Polarization: "o"},
			{DeviceID: "BC", Photon: true, Wavelength: 532, Polarization: "p"},
			{DeviceID: "BC", Photon: true, Wavelength: 1064, Polarization: "s"},
		},
	}

	pr, ok := lf.SelectProfile(true, 532, "")
	assert.True(t, ok)
	assert.Equal(t, 532.0, pr.Wavelength)

	pr, ok = lf.SelectProfile(true, 532, "p")
	assert.True(t, ok)
	assert.Equal(t, 532.0, pr.Wavelength)

	pr, ok = lf.SelectProfile(true, 532, "o")
	assert.False(t, ok)

	pr, ok = lf.SelectProfile(true, 999, "")
	assert.False(t, ok)
	assert.Equal(t, LicelProfile{}, pr)
}

// --- LoadLicelFile ---

func TestLoadLicelFile_NonExistent(t *testing.T) {
	_, err := LoadLicelFile("/nonexistent/file.licel")
	assert.Error(t, err)
}

func TestLoadLicelFile_Testdata(t *testing.T) {
	testFile := filepath.Join("..", "testdata", "b2021019.223500")
	lf, err := LoadLicelFile(testFile)
	require.NoError(t, err)

	assert.True(t, lf.FileLoaded)
	assert.Equal(t, "Vladivos", lf.MeasurementSite)
	assert.Equal(t, 20.0, lf.AltitudeAboveSeaLevel)
	assert.InDelta(t, 131.9, lf.Longitude, 0.1)
	assert.InDelta(t, 43.1, lf.Latitude, 0.1)
	assert.Equal(t, 50.0, lf.Zenith)
	assert.Equal(t, 2001, lf.Laser1NShots)
	assert.Equal(t, 20, lf.Laser1Freq)
	assert.Equal(t, 12, lf.NDatasets)
	assert.Len(t, lf.Profiles, 12)
	assert.Equal(t, 16380, lf.Profiles[0].NDataPoints)
}

// --- Save errors ---

func TestLicelFile_Save_InvalidPath(t *testing.T) {
	lf := LicelFile{Profiles: LicelProfilesList{{}}}
	err := lf.Save("/nonexistent/dir/file.dat")
	assert.Error(t, err)
}

// --- Format*Lines ---

func TestFormatFirstLine(t *testing.T) {
	lf := LicelFile{}
	s := lf.formatFirstLine("myfile.dat")
	assert.Contains(t, s, "myfile.dat")
	assert.Contains(t, s, "\r\n")
}

func TestFormatSecondLine(t *testing.T) {
	lf := LicelFile{
		MeasurementSite:       "Test",
		MeasurementStartTime:  time.Date(2020, 2, 10, 19, 22, 35, 0, time.UTC),
		MeasurementStopTime:   time.Date(2020, 2, 10, 19, 24, 15, 0, time.UTC),
		AltitudeAboveSeaLevel: 20,
		Longitude:             131.9,
		Latitude:              43.1,
		Zenith:                50,
	}
	s := lf.formatSecondLine()
	assert.Contains(t, s, "Test")
	assert.Contains(t, s, "0131.9")
	assert.Contains(t, s, "0043.1")
	assert.Contains(t, s, "50")
}

func TestFormatThirdLine(t *testing.T) {
	lf := LicelFile{
		Laser1NShots: 2001,
		Laser1Freq:   20,
		Laser2NShots: 0,
		Laser2Freq:   10,
		NDatasets:    12,
		Laser3NShots: 0,
		Laser3Freq:   10,
	}
	s := lf.formatThirdLine()
	assert.Contains(t, s, "0002001")
	assert.Contains(t, s, "0020")
	assert.Contains(t, s, "12")
}

// --- Glue ---

func TestLicelFile_Glue_Success(t *testing.T) {
	n := 100
	analogData := make([]float64, n)
	photonData := make([]float64, n)
	for i := 0; i < n; i++ {
		analogData[i] = float64(1000 + i*10)
		photonData[i] = float64(200 + i)
	}

	lf := LicelFile{
		Profiles: LicelProfilesList{
			{DeviceID: "BT", Photon: false, Wavelength: 532, Polarization: "p", BinWidth: 7.5, NDataPoints: n, Data: analogData, Active: true, LaserType: 1, NShots: 2001, DiscrLevel: 0.5},
			{DeviceID: "BC", Photon: true, Wavelength: 532, Polarization: "p", BinWidth: 7.5, NDataPoints: n, Data: photonData, Active: true, LaserType: 1, NShots: 2000, DiscrLevel: 0.005},
		},
	}

	got, err := lf.Glue(532, 150, 300, "p")
	require.NoError(t, err)

	assert.Equal(t, "BG", got.DeviceID)
	assert.False(t, got.Photon)
	assert.Equal(t, 532.0, got.Wavelength)
	assert.Equal(t, "p", got.Polarization)
	assert.Equal(t, n, len(got.Data))

	// h < h1 (i < 20): analog
	for i := 0; i < 20; i++ {
		assert.Equal(t, analogData[i], got.Data[i], "at index %d (h < h1)", i)
	}

	// k вычисляется на [idx1, idx2] = [20, 40]
	var sumK float64
	cnt := 0
	for i := 20; i <= 40; i++ {
		if photonData[i] != 0 {
			sumK += analogData[i] / photonData[i]
			cnt++
		}
	}
	k := sumK / float64(cnt)
	for i := 20; i <= 40; i++ {
		expected := 0.5 * (analogData[i] + k*photonData[i])
		assert.InDelta(t, expected, got.Data[i], 1e-9, "at index %d (glue zone)", i)
	}

	// h > h2 (i > 40): k*photon
	for i := 41; i < n; i++ {
		expected := k * photonData[i]
		assert.InDelta(t, expected, got.Data[i], 1e-9, "at index %d (h > h2)", i)
	}
}

func TestLicelFile_Glue_MissingAnalog(t *testing.T) {
	lf := LicelFile{
		Profiles: LicelProfilesList{
			{DeviceID: "BC", Photon: true, Wavelength: 532, Polarization: "p", BinWidth: 7.5, Data: make([]float64, 100)},
		},
	}
	_, err := lf.Glue(532, 10, 50, "p")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "analog channel not found")
}

func TestLicelFile_Glue_MissingPhoton(t *testing.T) {
	lf := LicelFile{
		Profiles: LicelProfilesList{
			{DeviceID: "BT", Photon: false, Wavelength: 532, Polarization: "p", BinWidth: 7.5, Data: make([]float64, 100)},
		},
	}
	_, err := lf.Glue(532, 10, 50, "p")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "photon channel not found")
}

func TestLicelFile_Glue_InvalidRange(t *testing.T) {
	lf := LicelFile{
		Profiles: LicelProfilesList{
			{DeviceID: "BT", Photon: false, Wavelength: 532, Polarization: "p", BinWidth: 7.5, Data: make([]float64, 100)},
			{DeviceID: "BC", Photon: true, Wavelength: 532, Polarization: "p", BinWidth: 7.5, Data: make([]float64, 100)},
		},
	}
	_, err := lf.Glue(532, 50, 10, "p")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "h1")
	assert.Contains(t, err.Error(), "must be less than h2")
}

func TestLicelFile_Glue_H1OutOfRange(t *testing.T) {
	lf := LicelFile{
		Profiles: LicelProfilesList{
			{DeviceID: "BT", Photon: false, Wavelength: 532, Polarization: "p", BinWidth: 7.5, Data: make([]float64, 10)},
			{DeviceID: "BC", Photon: true, Wavelength: 532, Polarization: "p", BinWidth: 7.5, Data: make([]float64, 10)},
		},
	}
	_, err := lf.Glue(532, 200, 300, "p") // h1 maps to idx 26, exceeds dataLen 10
	assert.Error(t, err)
}

func TestLicelFile_Glue_AllZeroPhoton(t *testing.T) {
	lf := LicelFile{
		Profiles: LicelProfilesList{
			{DeviceID: "BT", Photon: false, Wavelength: 532, Polarization: "p", BinWidth: 7.5, Data: []float64{100, 200, 300, 400}},
			{DeviceID: "BC", Photon: true, Wavelength: 532, Polarization: "p", BinWidth: 7.5, Data: []float64{0, 0, 0, 0}},
		},
	}
	_, err := lf.Glue(532, 0, 22, "p") // idx2=2, dataLen=4 — попадает в диапазон, но photon=0
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "all photon data values are zero")
}

func TestLicelFile_Glue_DataLengthMismatch(t *testing.T) {
	// photon shorter than analog — uses min length
	lf := LicelFile{
		Profiles: LicelProfilesList{
			{DeviceID: "BT", Photon: false, Wavelength: 532, Polarization: "p", BinWidth: 7.5, Data: []float64{100, 200, 300, 400, 500, 600}},
			{DeviceID: "BC", Photon: true, Wavelength: 532, Polarization: "p", BinWidth: 7.5, Data: []float64{10, 20, 30}},
		},
	}
	got, err := lf.Glue(532, 0, 15, "p") // idx2=2, fits in 3
	require.NoError(t, err)
	assert.Equal(t, 3, len(got.Data))
}

// --- LoadLicelFileFromReader ---

func TestLoadLicelFileFromReader_Empty(t *testing.T) {
	_, err := LoadLicelFileFromReader(os.Stdin)
	// Will fail because os.Stdin has no data
	assert.Error(t, err)
}
