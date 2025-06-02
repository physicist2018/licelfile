# LicelFormat

[![Go Reference](https://pkg.go.dev/badge/github.com/physicist2018/licelfile/tree/v1.1.5.svg)](https://pkg.go.dev/github.com/physicist2018/licelfile/tree/v1.1.5)


The `licelformat` package provides utilities for parsing and processing Licel format data files. It supports reading, extracting metadata, and converting binary data into usable formats. This package is intended for working with Licel files, which contain measurement profiles and other associated data.

## Features

- **Parsing**: Parse Licel data files to extract metadata and measurement profiles.
- **Data conversion**: Convert raw binary data into readable floating-point values.
- **File handling**: Load multiple Licel files matching a file mask.
- **Profile selection**: Filter profiles by photon type and wavelength.

## Installation

To install the `licelformat` package, use the following Go command:

```bash
go get github.com/physicist2018/licelfile
```

## Usage
### Load a Licel File
To load and parse a Licel file:

```go
package main

import (
	"log"
	"github.com/physicist2018/licelfile/licelformat"
)

func main() {
	licelFile := licelformat.LoadLicelFile("path/to/file")
}
```

### Selecting Profiles by Wavelength
To select profiles by wavelength and photon type:

```go
package main

import (
	"log"
	"github.com/physicist2018/licelfile/licelformat"
)

func main() {
	licelPack := licelformat.NewLicelPack("path/to/files/*.licel")
	profiles := licelformat.SelectCertainWavelength2(&licelPack, true, 532.0)

	for _, profile := range profiles {
		log.Printf("Profile: %+v", profile)
	}
}
```

### Functions
`NewLicelProfile(line string) LicelProfile`
Parses a single line of text from a Licel file and returns a LicelProfile struct.

`LoadLicelFile(fname string) LicelFile`
Loads a Licel file, parses its headers and binary data, and returns a LicelFile struct with the parsed data.

`LoadLicelFileFromReader(f io.Reader, size int64) LicelFile`
Loads Licel file from reader, parses its headers and binary data, and returns a LicelFile struct with the parsed data.

`NewLicelPackFromZip(zipPath string) *LicelPack`
Loads licel measurements from zip archive

`SelectCertainWavelength1(lf *LicelFile, isPhoton bool, wavelength float64) LicelProfile`
Selects a profile from a single Licel file by photon type and wavelength.

`SelectCertainWavelength2(lp *LicelPack, isPhoton bool, wavelength float64) LicelProfilesList`
Selects profiles from multiple Licel files in a LicelPack by photon type and wavelength.

`NewLicelPack(mask string) LicelPack`
Loads multiple Licel files matching the specified file mask and returns a LicelPack.

### Structs
`LicelProfile`
A struct representing a measurement profile. It includes various fields such as:

- Active: Indicates if the channel is active.
- Photon: Indicates if the measurement is based on photon data.
- LaserType: The type of laser used for the measurement.
- NDataPoints: The number of data points in the profile.
- Wavelength: The wavelength of the laser used in the measurement.
- Data: A slice of floating-point values representing the data points.

`LicelFile`
A struct representing a Licel file, including metadata about the measurement and the list of profiles.

- MeasurementSite: The location of the measurement.
- MeasurementStartTime: The start time of the measurement.
- Profiles: A list of LicelProfile structs representing the measurement profiles.

`LicelPack`
A map of LicelFile structs, representing multiple Licel files loaded by a file mask.

## Logging
This package uses the zerolog package for logging errors and warnings. Ensure that you import and configure zerolog in your main program to capture detailed logging information.

## License
The code is released under the LGPL V3 License.



## TODO
- переделать читатель, оптимизировать для более строгого чтения файлов
