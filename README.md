# LicelFormat

[![Go Reference](https://pkg.go.dev/badge/github.com/physicist2018/licelfile/v2.svg)](https://pkg.go.dev/github.com/physicist2018/licelfile/v2)

The `licelformat` package provides utilities for parsing and processing Licel format data files. It supports reading, extracting metadata, converting binary data, and round-tripping files through save/load.

## Features

- **Parsing**: Parse Licel binary files to extract metadata and measurement profiles.
- **Data conversion**: Convert raw little-endian int32 binary data into float64 values with proper per-channel scaling.
- **Safe round-trip**: Save → load produces identical data; scaling is handled transparently.
- **Zip support**: Load packs from and save packs to zip archives.
- **Profile selection**: Filter profiles by photon type and wavelength across single files or entire packs.

## Installation

```bash
go get github.com/physicist2018/licelfile/v2
```

## Usage

### Load a Licel file

```go
lf, err := licelformat.LoadLicelFile("path/to/file")
if err != nil {
    log.Fatal(err)
}
fmt.Println(lf.MeasurementSite, lf.NDatasets)
```

### Load from an io.Reader

```go
lf, err := licelformat.LoadLicelFileFromReader(myReader)
```

### Save a file

```go
if err := lf.Save("output.dat"); err != nil {
    log.Fatal(err)
}
```

### Load a pack by glob mask

```go
pack, err := licelformat.NewLicelPack("data/*.licel")
```

### Load a pack from zip

```go
pack, err := licelformat.NewLicelPackFromZip("archive.zip")
```

### Save a pack to zip

```go
if err := pack.SaveToZip("output.zip"); err != nil {
    log.Fatal(err)
}
```

### Select profiles

```go
// From a single file — match any polarization
profile, ok := lf.SelectProfile(true, 532.0, "")

// From a single file — match specific polarization
profile, ok := lf.SelectProfile(false, 355.0, "o")

// Across all files in a pack — match any polarization
profiles := pack.SelectProfiles(false, 355.0, "")

// Across all files in a pack — match specific polarization
profiles := pack.SelectProfiles(true, 1064.0, "s")
```

## API

### Types

**`LicelFile`** — a single measurement with metadata and profiles.

| Field                  | Type             | Description                  |
|-----------------------|------------------|------------------------------|
| `MeasurementSite`     | `string`         | Measurement location         |
| `MeasurementStartTime`| `time.Time`      | Start time                   |
| `MeasurementStopTime` | `time.Time`      | Stop time                    |
| `AltitudeAboveSeaLevel`| `float64`       | Lidar altitude               |
| `Longitude`           | `float64`        | Longitude                    |
| `Latitude`            | `float64`        | Latitude                     |
| `Zenith`              | `float64`        | Zenith angle                 |
| `Laser1NShots`        | `int`            | Laser 1 shot count           |
| `Laser1Freq`          | `int`            | Laser 1 frequency            |
| `Laser2NShots`        | `int`            | Laser 2 shot count           |
| `Laser2Freq`          | `int`            | Laser 2 frequency            |
| `NDatasets`           | `int`            | Number of profiles           |
| `Laser3NShots`        | `int`            | Laser 3 shot count           |
| `Laser3Freq`          | `int`            | Laser 3 frequency            |
| `Profiles`            | `LicelProfilesList` | Measurement profiles       |

**`LicelProfile`** — a single measurement channel.

| Field         | Type       | Description        |
|---------------|-----------|---------------------|
| `Active`      | `bool`    | Channel active      |
| `Photon`      | `bool`    | Photon counting mode|
| `LaserType`   | `int`     | Laser type          |
| `NDataPoints` | `int`     | Number of data points|
| `Wavelength`  | `float64` | Wavelength (nm)     |
| `Polarization`| `string`  | Polarization        |
| `Data`        | `[]float64`| Scaled data points |

**`LicelPack`** — collection of `LicelFile` instances.

| Field        | Type              | Description                  |
|--------------|-------------------|------------------------------|
| `StartTime`  | `time.Time`       | Earliest measurement start   |
| `StopTime`   | `time.Time`       | Latest measurement stop      |
| `Data`       | `map[string]LicelFile` | Files keyed by filename  |
| `ZipCompressionLevel` | `int`     | Deflate level for zip (0–9)  |

### Functions

| Function | Signature |
|----------|-----------|
| `LoadLicelFile` | `(fname string) (LicelFile, error)` |
| `LoadLicelFileFromReader` | `(r io.Reader) (LicelFile, error)` |
| `NewLicelProfile` | `(line string) (LicelProfile, error)` |
| `NewLicelPack` | `(mask string) (*LicelPack, error)` |
| `NewLicelPackFromZip` | `(zipPath string) (*LicelPack, error)` |

### Methods

| Method | Receiver | Signature |
|--------|----------|-----------|
| `Save` | `*LicelFile` | `(fname string) error` |
| `WriteTo` | `*LicelFile` | `(w io.Writer, fname string) error` |
| `SelectProfile` | `*LicelFile` | `(isPhoton bool, wavelength float64, polarization string) (LicelProfile, bool)` |
| `SetMaxDist` | `*LicelFile` | `(alt float64) error` |
| `FormatFirstLine` | `*LicelFile` | `(fname string) string` |
| `FormatSecondLine` | `*LicelFile` | `() string` |
| `FormatThirdLine` | `*LicelFile` | `() string` |
| `Metadata` | `*LicelProfile` | `() string` |
| `ProfileRaw` | `*LicelProfile` | `() ([]byte, error)` |
| `SetMaxDist` | `*LicelProfile` | `(alt float64) error` |
| `Save` | `*LicelPack` | `() error` |
| `SaveToZip` | `*LicelPack` | `(zipPath string) error` |
| `SelectProfiles` | `*LicelPack` | `(isPhoton bool, wavelength float64, polarization string) LicelProfilesList` |
| `Filter` | `*LicelPack` | `(cond func(lf *LicelFile) bool) LicelPack` |
| `FilterProfiles` | `*LicelPack` | `(cond func(pr *LicelProfile) bool) LicelPack` |
| `FilterProfilesList` | `*LicelPack` | `(cond func(pr *LicelProfile) bool) LicelProfilesList` |
| `SetMaxDist` | `*LicelPack` | `(alt float64) error` |

## License

LGPL V3.
