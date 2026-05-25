package licelformat

import (
	"fmt"
	"path/filepath"
	"testing"

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
					{Wavelength: 355, Photon: false},
					{Wavelength: 532, Photon: true},
				},
			},
			"file2": {
				Profiles: LicelProfilesList{
					{Wavelength: 355, Photon: true},
				},
			},
		},
	}

	result := lp.SelectProfiles(false, 355)
	assert.Len(t, result, 1)
	assert.Equal(t, 355.0, result[0].Wavelength)
	assert.False(t, result[0].Photon)

	result = lp.SelectProfiles(true, 355)
	assert.Len(t, result, 1)

	result = lp.SelectProfiles(true, 999)
	assert.Len(t, result, 0)
}

// --- SaveToZip ---

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
