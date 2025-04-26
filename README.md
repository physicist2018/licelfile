# LicelFormat
A package for working with Licel files, used for reading data related to laser measurements and profiles.

## Installation
To install the package, run the command:

```bash
go get github.com/physicist2018/licelfile
```

## Function Description

`NewLicelProfile(line string) LicelProfile`

Parses a profile string and returns a `LicelProfile` structure containing channel and measurement data.

```go
profile := NewLicelProfile("1 0 1 100 0 1000 0.5 400.0.POL 0 0 0 10 1000 0.2 DeviceID 100")
```

`LoadLicelFile(fname string) LicelFile`

Loads a Licel file from the specified path fname and returns a LicelFile structure containing information about the file, profiles, and data. Example usage:

```go
licelFile := LoadLicelFile("path/to/file.txt")
```

`NewLicelPack(mask string) LicelPack`
Loads multiple Licel files matching the specified mask and returns a map of files in the LicelPack type. Example usage:

```go
pack := NewLicelPack("path/to/files/*.txt")
```

`SelectCertainWavelength1(lf *LicelFile, isPhoton bool, wavelength float64) LicelProfile`

Selects a profile by wavelength from a single file. Example usage:

```go
profile := SelectCertainWavelength1(&licelFile, true, 400.0)
```

`SelectCertainWavelength2(lp *LicelPack, isPhoton bool, wavelength float64) LicelProfilesList`
Selects all profiles by wavelength from a set of files. Example usage:

```go
profiles := SelectCertainWavelength2(&pack, true, 400.0)
```

Utility Functions(non exportable)

`str2Bool(str string) bool`: Converts a string to a boolean value.

`str2Int(str string) int64`: Converts a string to an integer.

`str2Float(str string) float64`: Converts a string to a floating-point number.

`bytesToFloat64Array(b []byte) []float64`: Converts a byte array to a float64 array.

`readAndTrimLine(r *bufio.Reader) string`: Reads a line from the buffer and trims whitespace characters on the right.

`skipCRLF(r *bufio.Reader)`: Skips CR and LF characters.

`parseTime(s string) time.Time`: Converts a string to a time format.

## Logging
The package uses zerolog for logging errors and important events. Logging examples:

Logging an error when reading a file:

```go
log.Fatal().Err(err).Str("file", fname).Msg("Ошибка при открытии файла")
```

Logging successful data loading:

```go
log.Info().Str("file", fname).Msg("Файл успешно загружен")
```

File Format

Licel files contain the following data:

- Header lines, containing information about the measurement site, time, lasers, and other parameters.
- Measurement data, presented in binary format (32-bit numbers, little-endian).


Example Usage

```go
package main

import (
	"fmt"
	"log"
	"github.com/physicist/licelfile/licelformat"
)

func main() {
	// Загрузка файла
	licelFile := licelfile.LoadLicelFile("path/to/file.txt")

	// Получение профиля по длине волны
	profile := licelfile.SelectCertainWavelength1(&licelFile, true, 400.0)

	// Вывод данных профиля
	fmt.Printf("Profile: %+v\n", profile)
}
```

## License
This package is distributed under the LGPL V3 license. See the LICENSE file for details.
