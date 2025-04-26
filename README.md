# LicelFormat
Пакет для работы с файлами Licel, используемый для чтения данных, связанных с измерениями лазеров и профилями

## Установка
Чтобы установить пакет, выполните команду:

```bash
go get github.com/physicist2018/licelfile
```

## Описание функций

`NewLicelProfile(line string) LicelProfile`

Парсит строку профиля и возвращает структуру LicelProfile, содержащую данные о канале и измерениях. Пример использования:

```go
profile := NewLicelProfile("1 0 1 100 0 1000 0.5 400.0.POL 0 0 0 10 1000 0.2 DeviceID 100")
```

`LoadLicelFile(fname string) LicelFile`

Загружает файл Licel по заданному пути fname и возвращает структуру LicelFile, которая содержит информацию о файле, профилях и данных. Пример использования:

```go
licelFile := LoadLicelFile("path/to/file.txt")
```

`NewLicelPack(mask string) LicelPack`
Загружает несколько файлов Licel, соответствующих маске, и возвращает карту файлов в типе LicelPack. Пример использования:

```go
pack := NewLicelPack("path/to/files/*.txt")
```

`SelectCertainWavelength1(lf *LicelFile, isPhoton bool, wavelength float64) LicelProfile`
Выбирает профиль по длине волны из одного файла. Пример использования:

```go
profile := SelectCertainWavelength1(&licelFile, true, 400.0)
```

`SelectCertainWavelength2(lp *LicelPack, isPhoton bool, wavelength float64) LicelProfilesList`
Выбирает все профили по заданной длине волны из набора файлов. Пример использования:

```go
profiles := SelectCertainWavelength2(&pack, true, 400.0)
```

Утилитные функции(не экспортируемые)

`str2Bool(str string) bool`: Преобразует строку в булево значение.

`str2Int(str string) int64`: Преобразует строку в целое число.

`str2Float(str string) float64`: Преобразует строку в число с плавающей запятой.

`bytesToFloat64Array(b []byte) []float64`: Преобразует массив байт в массив чисел типа float64.

`readAndTrimLine(r *bufio.Reader) string`: Читает строку из буфера и удаляет символы пробела справа.

`skipCRLF(r *bufio.Reader)`: Пропускает символы CR и LF.

`parseTime(s string) time.Time`: Преобразует строку в формат времени.

## Логирование
Пакет использует zerolog для логирования ошибок и важных событий. Примеры логирования:

Логирование ошибки при чтении файла:

```go
log.Fatal().Err(err).Str("file", fname).Msg("Ошибка при открытии файла")
```

Логирование успешной загрузки данных:

```go
log.Info().Str("file", fname).Msg("Файл успешно загружен")
```

Формат файлов
Файлы Licel содержат следующие данные:

Строки заголовков, содержащие информацию о месте измерений, времени, лазерах и других параметрах.

Данные измерений, представленные в бинарном формате (32-битные числа, little-endian).

Пример использования

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

## Лицензия
Этот пакет распространяется под лицензией GNU GPL v3. См. файл LICENSE для подробностей.

