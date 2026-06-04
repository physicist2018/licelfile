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

### Filter files in a pack

```go
// Keep only files from a specific site
filtered := pack.Filter(func(lf *licelformat.LicelFile) bool {
    return lf.MeasurementSite == "Observatory"
})

// Keep only files within a time range
filtered := pack.Filter(func(lf *licelformat.LicelFile) bool {
    return lf.MeasurementStartTime.After(startOfDay) &&
           lf.MeasurementStartTime.Before(endOfDay)
})
```

### Filter profiles within a pack

```go
// Keep only analog profiles, drop files with none
analogPack := pack.FilterProfiles(func(pr *licelformat.LicelProfile) bool {
    return !pr.Photon
})

// Keep only 355nm profiles
pack355 := pack.FilterProfiles(func(pr *licelformat.LicelProfile) bool {
    return pr.Wavelength == 355.0
})
```

### Collect matching profiles into a flat list

```go
// Get all analog profiles as a flat list
analogProfiles := pack.FilterProfilesList(func(pr *licelformat.LicelProfile) bool {
    return !pr.Photon
})
for _, pr := range analogProfiles {
    fmt.Println(pr.Wavelength, pr.Polarization)
}

// Get a flat list of all 355nm profiles
profiles355 := pack.FilterProfilesList(func(pr *licelformat.LicelProfile) bool {
    return pr.Wavelength == 355.0
})
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
| `NewLicelPack` | `(mask string) (*LicelPack, error)` |
| `NewLicelPackFromZip` | `(zipPath string) (*LicelPack, error)` |

### Methods

| Method | Receiver | Signature |
|--------|----------|-----------|
| `Save` | `*LicelFile` | `(fname string) error` |
| `WriteTo` | `*LicelFile` | `(w io.Writer, fname string) error` |
| `SelectProfile` | `*LicelFile` | `(isPhoton bool, wavelength float64, polarization string) (LicelProfile, bool)` |
| `Glue` | `*LicelFile` | `(wvl float64, h1, h2 float64) (LicelProfile, error)` |
| `SetMaxDist` | `*LicelFile` | `(alt float64) error` |
| `IsPhoton` | `*LicelProfile` | `() bool` |
| `IsAnalog` | `*LicelProfile` | `() bool` |
| `IsGlued` | `*LicelProfile` | `() bool` |
| `SetMaxDist` | `*LicelProfile` | `(alt float64) error` |
| `Save` | `*LicelPack` | `() error` |
| `SaveToZip` | `*LicelPack` | `(zipPath string) error` |
| `SelectProfiles` | `*LicelPack` | `(isPhoton bool, wavelength float64, polarization string) LicelProfilesList` |
| `Filter` | `*LicelPack` | `(cond func(lf *LicelFile) bool) LicelPack` |
| `FilterProfiles` | `*LicelPack` | `(cond func(pr *LicelProfile) bool) LicelPack` |
| `FilterProfilesList` | `*LicelPack` | `(cond func(pr *LicelProfile) bool) LicelProfilesList` |
| `ToProfilesList` | `*LicelPack` | `() LicelProfilesList` |
| `SetMaxDist` | `*LicelPack` | `(alt float64) error` |
| `Glue` | `*LicelPack` | `(wvl float64, h1, h2 float64) error` |

### Glue analog and photon channels

```go
// In a single file: glue 355nm analog+photon, compute ratio in [500; 2000]m
glued, err := lf.Glue(355.0, 500.0, 2000.0)
if err != nil {
    log.Fatal(err)
}
// glued.DeviceID == "BG"
// glued.IsGlued() == true

// In a whole pack: glue every file
if err := pack.Glue(355.0, 500.0, 2000.0); err != nil {
    log.Fatal(err)
}
```

## License

LGPL V3.
