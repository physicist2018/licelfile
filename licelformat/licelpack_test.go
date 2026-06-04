package licelformat

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsValidFilename(t *testing.T) {
	assert.True(t, isValidFilename("b2021019.223500"))
	assert.True(t, isValidFilename("b0000000.000000"))
	assert.True(t, isValidFilename("a12345.678901"))
	assert.False(t, isValidFilename("b12345"))
	assert.False(t, isValidFilename(""))
}

func TestSelectProfiles(t *testing.T) {
	lp := &LicelPack{
		Data: map[string]LicelFile{
			"file1": {
				Profiles: LicelProfilesList{
					{Wavelength: 355, Photon: false, Polarization: "o"},
					{Wavelength: 532, Photon: true, Polarization: "p"},
				},
			},
			"file2": {
				Profiles: LicelProfilesList{
					{Wavelength: 355, Photon: true, Polarization: "s"},
					{Wavelength: 355, Photon: false, Polarization: "p"},
				},
			},
		},
	}

	result := lp.SelectProfiles(false, 355, "")
	assert.Len(t, result, 2)
	assert.Equal(t, 355.0, result[0].Wavelength)
	assert.False(t, result[0].Photon)

	result = lp.SelectProfiles(true, 355, "")
	assert.Len(t, result, 1)

	result = lp.SelectProfiles(false, 355, "p")
	assert.Len(t, result, 1)
	assert.Equal(t, "p", result[0].Polarization)

	result = lp.SelectProfiles(true, 999, "")
	assert.Len(t, result, 0)
}

// --- SetMaxDist ---

func TestLicelPack_SetMaxDist(t *testing.T) {
	lp := &LicelPack{
		Data: map[string]LicelFile{
			"file1": {Profiles: LicelProfilesList{{BinWidth: 7.5, NDataPoints: 100, Data: make([]float64, 100)}}},
			"file2": {Profiles: LicelProfilesList{{BinWidth: 7.5, NDataPoints: 200, Data: make([]float64, 200)}}},
		},
	}

	err := lp.SetMaxDist(375) // idx = 50
	require.NoError(t, err)

	pf1 := lp.Data["file1"].Profiles[0]
	assert.Equal(t, 50, pf1.NDataPoints)
	assert.Len(t, pf1.Data, 50)

	pf2 := lp.Data["file2"].Profiles[0]
	assert.Equal(t, 50, pf2.NDataPoints)
	assert.Len(t, pf2.Data, 50)
}

func TestLicelPack_SetMaxDist_Error(t *testing.T) {
	lp := &LicelPack{
		Data: map[string]LicelFile{
			"file1": {Profiles: LicelProfilesList{{BinWidth: 7.5, NDataPoints: 10, Data: make([]float64, 10)}}},
		},
	}

	err := lp.SetMaxDist(375) // idx = 50 > 10
	assert.Error(t, err)
}

// --- Filter ---

func TestLicelPack_Filter_AllMatch(t *testing.T) {
	lp := &LicelPack{
		Data: map[string]LicelFile{
			"f1": {
				MeasurementSite:      "A",
				MeasurementStartTime: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
				MeasurementStopTime:  time.Date(2024, 1, 1, 10, 5, 0, 0, time.UTC),
			},
			"f2": {
				MeasurementStartTime: time.Date(2024, 1, 1, 10, 10, 0, 0, time.UTC),
				MeasurementStopTime:  time.Date(2024, 1, 1, 10, 15, 0, 0, time.UTC),
			},
		},
	}

	result := lp.Filter(func(lf *LicelFile) bool { return true })

	assert.Len(t, result.Data, 2)
	assert.Contains(t, result.Data, "f1")
	assert.Contains(t, result.Data, "f2")
	assert.Equal(t, time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), result.StartTime)
	assert.Equal(t, time.Date(2024, 1, 1, 10, 15, 0, 0, time.UTC), result.StopTime)

	// исходный пак не изменён
	assert.Len(t, lp.Data, 2)
}

func TestLicelPack_Filter_NoneMatch(t *testing.T) {
	lp := &LicelPack{
		Data: map[string]LicelFile{
			"f1": {},
			"f2": {},
		},
	}

	result := lp.Filter(func(lf *LicelFile) bool { return false })

	assert.Len(t, result.Data, 0)
	assert.True(t, result.StartTime.IsZero())
	assert.True(t, result.StopTime.IsZero())
}

func TestLicelPack_Filter_Partial(t *testing.T) {
	lp := &LicelPack{
		Data: map[string]LicelFile{
			"f1": {
				MeasurementSite:      "A",
				MeasurementStartTime: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
				MeasurementStopTime:  time.Date(2024, 1, 1, 10, 5, 0, 0, time.UTC),
			},
			"f2": {
				MeasurementSite:      "B",
				MeasurementStartTime: time.Date(2024, 1, 1, 10, 10, 0, 0, time.UTC),
				MeasurementStopTime:  time.Date(2024, 1, 1, 10, 15, 0, 0, time.UTC),
			},
			"f3": {
				MeasurementSite:      "A",
				MeasurementStartTime: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
				MeasurementStopTime:  time.Date(2024, 1, 1, 9, 30, 0, 0, time.UTC),
			},
		},
	}

	result := lp.Filter(func(lf *LicelFile) bool { return lf.MeasurementSite == "A" })

	assert.Len(t, result.Data, 2)
	assert.Contains(t, result.Data, "f1")
	assert.NotContains(t, result.Data, "f2")
	assert.Contains(t, result.Data, "f3")
	// StartTime = min(f1=10:00, f3=9:00) = 9:00
	assert.Equal(t, time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC), result.StartTime)
	// StopTime = max(f1=10:05, f3=9:30) = 10:05
	assert.Equal(t, time.Date(2024, 1, 1, 10, 5, 0, 0, time.UTC), result.StopTime)
}

func TestLicelPack_Filter_PreservesCompression(t *testing.T) {
	lp := &LicelPack{
		ZipCompressionLevel: 7,
		Data: map[string]LicelFile{
			"f1": {},
		},
	}

	result := lp.Filter(func(lf *LicelFile) bool { return true })

	assert.Equal(t, 7, result.ZipCompressionLevel)
}

// --- SaveToZip ---

// --- FilterProfilesList ---

func TestLicelPack_FilterProfilesList_AllMatch(t *testing.T) {
	lp := &LicelPack{
		Data: map[string]LicelFile{
			"f1": {
				Profiles: LicelProfilesList{
					{Wavelength: 355, Photon: false, Polarization: "o"},
					{Wavelength: 532, Photon: true, Polarization: "p"},
				},
			},
			"f2": {
				Profiles: LicelProfilesList{
					{Wavelength: 355, Photon: true, Polarization: "s"},
					{Wavelength: 355, Photon: false, Polarization: "p"},
				},
			},
		},
	}

	result := lp.FilterProfilesList(func(pr *LicelProfile) bool { return true })

	assert.Len(t, result, 4)
	assert.Equal(t, 355.0, result[0].Wavelength)
	assert.False(t, result[0].Photon)
	assert.Equal(t, "o", result[0].Polarization)

	// исходный пак не изменён
	assert.Len(t, lp.Data, 2)
}

func TestLicelPack_FilterProfilesList_NoneMatch(t *testing.T) {
	lp := &LicelPack{
		Data: map[string]LicelFile{
			"f1": {
				Profiles: LicelProfilesList{{Wavelength: 355}},
			},
			"f2": {
				Profiles: LicelProfilesList{{Wavelength: 532}},
			},
		},
	}

	result := lp.FilterProfilesList(func(pr *LicelProfile) bool { return false })

	assert.Len(t, result, 0)
}

func TestLicelPack_FilterProfilesList_Partial(t *testing.T) {
	lp := &LicelPack{
		Data: map[string]LicelFile{
			"f1": {
				Profiles: LicelProfilesList{
					{Wavelength: 355, Photon: false},
					{Wavelength: 532, Photon: true},
				},
			},
			"f2": {
				Profiles: LicelProfilesList{
					{Wavelength: 355, Photon: true},
					{Wavelength: 1064, Photon: false},
				},
			},
			"f3": {
				Profiles: LicelProfilesList{
					{Wavelength: 999, Photon: true},
					{Wavelength: 888, Photon: true},
				},
			},
		},
	}

	// только аналоговые профили (!Photon)
	result := lp.FilterProfilesList(func(pr *LicelProfile) bool { return !pr.Photon })

	assert.Len(t, result, 2) // 355 (f1) и 1064 (f2)
	assert.Equal(t, 355.0, result[0].Wavelength)
	assert.Equal(t, 1064.0, result[1].Wavelength)
	assert.False(t, result[0].Photon)
	assert.False(t, result[1].Photon)

	// исходный пак не изменён
	assert.Contains(t, lp.Data, "f1")
	assert.Contains(t, lp.Data, "f2")
	assert.Contains(t, lp.Data, "f3")
}

func TestLicelPack_FilterProfilesList_SelectsAnalog(t *testing.T) {
	lp := &LicelPack{
		Data: map[string]LicelFile{
			"f1": {
				Profiles: LicelProfilesList{
					{Wavelength: 355, Photon: false, Polarization: "o", NDataPoints: 100, Data: make([]float64, 100)},
					{Wavelength: 355, Photon: false, Polarization: "p", NDataPoints: 200, Data: make([]float64, 200)},
					{Wavelength: 532, Photon: true, Polarization: "o", NDataPoints: 300, Data: make([]float64, 300)},
				},
			},
		},
	}

	// выбираем только аналоговые 355нм с поляризацией "o"
	result := lp.FilterProfilesList(func(pr *LicelProfile) bool {
		return !pr.Photon && pr.Wavelength == 355.0 && pr.Polarization == "o"
	})

	assert.Len(t, result, 1)
	assert.Equal(t, 355.0, result[0].Wavelength)
	assert.Equal(t, "o", result[0].Polarization)
	assert.Len(t, result[0].Data, 100)
}

// --- FilterProfiles ---

func TestLicelPack_FilterProfiles_AllMatch(t *testing.T) {
	lp := &LicelPack{
		ZipCompressionLevel: 5,
		Data: map[string]LicelFile{
			"f1": {
				NDatasets: 2,
				Profiles: LicelProfilesList{
					{Wavelength: 355, Photon: false, Polarization: "o"},
					{Wavelength: 532, Photon: true, Polarization: "p"},
				},
			},
		},
	}

	result := lp.FilterProfiles(func(pr *LicelProfile) bool { return true })

	require.Len(t, result.Data, 1)
	f1 := result.Data["f1"]
	assert.Len(t, f1.Profiles, 2)
	assert.Equal(t, 2, f1.NDatasets)
	assert.Equal(t, 5, result.ZipCompressionLevel)

	// исходный пак не изменён
	assert.Len(t, lp.Data, 1)
	assert.Equal(t, 2, lp.Data["f1"].NDatasets)
}

func TestLicelPack_FilterProfiles_NoneMatch(t *testing.T) {
	lp := &LicelPack{
		Data: map[string]LicelFile{
			"f1": {
				NDatasets: 1,
				Profiles:  LicelProfilesList{{Wavelength: 355}},
			},
		},
	}

	result := lp.FilterProfiles(func(pr *LicelProfile) bool { return false })

	assert.Len(t, result.Data, 0)
	assert.True(t, result.StartTime.IsZero())
	assert.True(t, result.StopTime.IsZero())
}

func TestLicelPack_FilterProfiles_Partial(t *testing.T) {
	lp := &LicelPack{
		Data: map[string]LicelFile{
			"f1": {
				MeasurementStartTime: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
				MeasurementStopTime:  time.Date(2024, 1, 1, 10, 5, 0, 0, time.UTC),
				NDatasets:            2,
				Profiles: LicelProfilesList{
					{Wavelength: 355, Photon: false},
					{Wavelength: 532, Photon: true},
				},
			},
			"f2": {
				MeasurementStartTime: time.Date(2024, 1, 1, 10, 10, 0, 0, time.UTC),
				MeasurementStopTime:  time.Date(2024, 1, 1, 10, 15, 0, 0, time.UTC),
				NDatasets:            2,
				Profiles: LicelProfilesList{
					{Wavelength: 355, Photon: true},
					{Wavelength: 1064, Photon: false},
				},
			},
			"f3": {
				NDatasets: 2,
				Profiles: LicelProfilesList{
					{Wavelength: 999, Photon: true},
					{Wavelength: 888, Photon: true},
				},
			},
		},
	}

	// только аналоговые профили (!Photon)
	result := lp.FilterProfiles(func(pr *LicelProfile) bool { return !pr.Photon })

	require.Len(t, result.Data, 2) // f1 и f2 остались, f3 исключён

	f1 := result.Data["f1"]
	assert.Len(t, f1.Profiles, 1)
	assert.Equal(t, 1, f1.NDatasets)
	assert.Equal(t, 355.0, f1.Profiles[0].Wavelength)

	f2 := result.Data["f2"]
	assert.Len(t, f2.Profiles, 1)
	assert.Equal(t, 1, f2.NDatasets)
	assert.Equal(t, 1064.0, f2.Profiles[0].Wavelength)

	assert.NotContains(t, result.Data, "f3") // нет аналоговых профилей

	// время пересчитано: min = f1(10:00), max = f2(10:15)
	assert.Equal(t, time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), result.StartTime)
	assert.Equal(t, time.Date(2024, 1, 1, 10, 15, 0, 0, time.UTC), result.StopTime)
}

func TestLicelPack_FilterProfiles_SelectsAnalog(t *testing.T) {
	lp := &LicelPack{
		ZipCompressionLevel: 3,
		Data: map[string]LicelFile{
			"f1": {
				MeasurementStartTime: time.Date(2024, 6, 1, 8, 0, 0, 0, time.UTC),
				MeasurementStopTime:  time.Date(2024, 6, 1, 8, 5, 0, 0, time.UTC),
				NDatasets:            3,
				Profiles: LicelProfilesList{
					{Wavelength: 355, Photon: false, Polarization: "o", NDataPoints: 100, Data: make([]float64, 100)},
					{Wavelength: 355, Photon: false, Polarization: "p", NDataPoints: 200, Data: make([]float64, 200)},
					{Wavelength: 532, Photon: true, Polarization: "o", NDataPoints: 300, Data: make([]float64, 300)},
				},
			},
		},
	}

	// выбираем только аналоговые 355нм с поляризацией "o"
	result := lp.FilterProfiles(func(pr *LicelProfile) bool {
		return !pr.Photon && pr.Wavelength == 355.0 && pr.Polarization == "o"
	})

	require.Len(t, result.Data, 1)
	f1 := result.Data["f1"]
	assert.Len(t, f1.Profiles, 1)
	assert.Equal(t, 1, f1.NDatasets)
	assert.Equal(t, 355.0, f1.Profiles[0].Wavelength)
	assert.Equal(t, "o", f1.Profiles[0].Polarization)
	assert.Len(t, f1.Profiles[0].Data, 100)
	assert.Equal(t, 3, result.ZipCompressionLevel)
}

func TestLicelPack_SaveToZip_Roundtrip(t *testing.T) {
	testFile := filepath.Join("..", "testdata", "b2021019.223500")
	lf, err := LoadLicelFile(testFile)
	require.NoError(t, err)

	pack := &LicelPack{
		StartTime: lf.MeasurementStartTime,
		Data: map[string]LicelFile{
			"b2021019.223500": lf,
		},
	}

	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "test.zip")

	err = pack.SaveToZip(zipPath)
	require.NoError(t, err)

	// Read back
	pack2, err := NewLicelPackFromZip(zipPath)
	require.NoError(t, err)

	assert.Equal(t, pack.StartTime, pack2.StartTime)
	require.Contains(t, pack2.Data, "/b2021019.223500")

	lf2 := pack2.Data["/b2021019.223500"]
	assert.Equal(t, lf.MeasurementSite, lf2.MeasurementSite)
	assert.Equal(t, lf.NDatasets, lf2.NDatasets)
	require.Len(t, lf2.Profiles, len(lf.Profiles))
	for i := range lf.Profiles {
		require.Len(t, lf2.Profiles[i].Data, len(lf.Profiles[i].Data))
		for j := range lf.Profiles[i].Data {
			assert.InDelta(t, lf.Profiles[i].Data[j], lf2.Profiles[i].Data[j], 0.1, "profile %d data[%d]", i, j)
		}
	}
}

// --- SaveToZip with compression levels ---

func TestLicelPack_SaveToZip_CompressionLevels(t *testing.T) {
	testFile := filepath.Join("..", "testdata", "b2021019.223500")
	lf, err := LoadLicelFile(testFile)
	require.NoError(t, err)

	levels := []int{0, 1, 5, 9}

	for _, level := range levels {
		t.Run(fmt.Sprintf("level_%d", level), func(t *testing.T) {
			pack := &LicelPack{
				StartTime:           lf.MeasurementStartTime,
				ZipCompressionLevel: level,
				Data: map[string]LicelFile{
					"b2021019.223500": lf,
				},
			}

			tmpDir := t.TempDir()
			zipPath := filepath.Join(tmpDir, "test.zip")

			err := pack.SaveToZip(zipPath)
			require.NoError(t, err, "level %d", level)

			pack2, err := NewLicelPackFromZip(zipPath)
			require.NoError(t, err, "level %d", level)

			lf2 := pack2.Data["/b2021019.223500"]
			assert.Equal(t, lf.NDatasets, lf2.NDatasets)
			require.Len(t, lf2.Profiles, len(lf.Profiles))
			for i := range lf.Profiles {
				require.Len(t, lf2.Profiles[i].Data, len(lf.Profiles[i].Data))
				for j := range lf.Profiles[i].Data {
					assert.InDelta(t, lf.Profiles[i].Data[j], lf2.Profiles[i].Data[j], 0.1, "level %d profile %d data[%d]", level, i, j)
				}
			}
		})
	}
}
