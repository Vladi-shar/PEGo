//go:build windows

package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"debug/pe"
	"encoding/hex"
	"fmt"
	"os"
	"reflect"
	"time"

	"golang.org/x/sys/windows"
)

type FileResources struct {
	CompanyName      string
	ProductName      string
	FileVersion      string
	ProductVersion   string
	OriginalFilename string
	Copyright        string
}

type FileProperties struct {
	FileName      string
	FileType      string
	FileSize      int64
	CreationDate  windows.Filetime
	ModifiedDate  windows.Filetime
	AccessDate    windows.Filetime
	Md5Hash       [16]byte
	Sha1Hash      [20]byte
	Sha256Hash    [32]byte
	FileResources FileResources
}

func getFileType(peFull *PeFull) (string, error) {
	optionalHdr, err := getOptionalHeader(peFull.peFile)
	if err != nil {
		return "", err
	}

	// Retrieve the Magic field from the optional header.
	var magic uint16
	switch hdr := optionalHdr.(type) {
	case *pe.OptionalHeader32:
		magic = hdr.Magic
	case *pe.OptionalHeader64:
		magic = hdr.Magic
	default:
		return "", fmt.Errorf("unexpected optional header type: %T", hdr)
	}

	// Determine the PE type based on the Magic value.
	var fileType string
	switch magic {
	case 0x10b:
		fileType = "PE32"
	case 0x20b:
		fileType = "PE32+ (x64)"
	default:
		fileType = "Unknown PE type"
	}

	return fileType, nil
}

func getFileTimes(filePath string) (windows.Filetime, windows.Filetime, windows.Filetime, error) {
	var creation windows.Filetime
	var access windows.Filetime
	var modified windows.Filetime

	file, err := os.Open(filePath)
	if err != nil {
		return creation, access, modified, err
	}
	defer file.Close()

	h := windows.Handle(file.Fd())
	err = windows.GetFileTime(h, &creation, &access, &modified)
	return creation, access, modified, err
}

func getFileProperties(peFull *PeFull, filePath string) (FileProperties, error) {
	var fileProperties FileProperties
	var err error
	fileProperties.FileName = filePath
	fileProperties.FileType, err = getFileType(peFull)
	if err != nil {
		return fileProperties, err
	}
	fileProperties.FileSize = int64(len(peFull.fileData))
	fileProperties.CreationDate, fileProperties.AccessDate, fileProperties.ModifiedDate, err = getFileTimes(filePath)
	if err != nil {
		return fileProperties, err
	}

	fileProperties.Md5Hash = md5.Sum(peFull.fileData)
	fileProperties.Sha1Hash = sha1.Sum(peFull.fileData)
	fileProperties.Sha256Hash = sha256.Sum256(peFull.fileData)

	frc, err := NewFileResourceCollector(filePath)
	if err != nil {
		return fileProperties, err
	}

	fileProperties.FileResources.CompanyName = frc.GetResource("CompanyName")
	fileProperties.FileResources.Copyright = frc.GetResource("LegalCopyright")
	fileProperties.FileResources.ProductName = frc.GetResource("ProductName")
	fileProperties.FileResources.OriginalFilename = frc.GetResource("OriginalFilename")
	fileProperties.FileResources.FileVersion = frc.GetResource("FileVersion")
	fileProperties.FileResources.ProductVersion = frc.GetResource("ProductVersion")

	return fileProperties, nil
}

func filetimeToTime(ft windows.Filetime) time.Time {
	// Combine high and low parts (in 100-ns units)
	ft100 := (int64(ft.HighDateTime) << 32) | int64(ft.LowDateTime)
	const offset int64 = 116444736000000000
	diff := ft100 - offset
	sec := diff / 10000000
	nsec := (diff % 10000000) * 100
	return time.Unix(sec, nsec)
}

func createTableForResources(fileResources FileResources) (*sortableTable, error) {
	t := reflect.TypeOf(fileResources)
	v := reflect.ValueOf(fileResources)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected a struct or a pointer to a struct, got %s", t.Kind())
	}

	// Prepare the data slice
	data := [][]string{
		{"Property", "Value"},
	}

	var longestFieldName = 0
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		valueStr := value.String()

		data = append(data, []string{
			field.Name,
			valueStr,
		})
		if len(field.Name) > longestFieldName {
			longestFieldName = len(field.Name)
		}
	}

	colWidths := []float32{200, float32(longestFieldName) * 30}
	colTypes := []ColumnType{unsortableCol, unsortableCol}
	colProps := []ColumnProps{{false, false}, {false, true}}

	return createNewSortableTable(colWidths, data, colTypes, colProps)
}

func createTableForProperties(fileProperties FileProperties) (*sortableTable, error) {
	// Use reflection to iterate over the struct fields
	t := reflect.TypeOf(fileProperties)
	v := reflect.ValueOf(fileProperties)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected a struct or a pointer to a struct, got %s", t.Kind())
	}

	// Prepare the data slice
	data := [][]string{
		{"Property", "Value"},
	}

	var longestFieldName = 0
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		var valueStr string
		switch value.Kind() {
		case reflect.String:
			valueStr = value.String()
		case reflect.Int, reflect.Int64:
			valueStr = fmt.Sprintf("%d", value.Int())
		case reflect.Struct:
			switch field.Type {
			case reflect.TypeOf(windows.Filetime{}):
				tm := filetimeToTime(value.Interface().(windows.Filetime))
				valueStr = tm.Format("Monday 02 January 2006, 15:04:05")
			default:
				continue
			}
		case reflect.Array:
			switch field.Type.Elem().Kind() {
			case reflect.Uint8:
				b := make([]byte, value.Len())
				for i := 0; i < value.Len(); i++ {
					b[i] = byte(value.Index(i).Uint())
				}
				valueStr = hex.EncodeToString(b)
			default:
				continue
			}
		default:
			continue
		}

		data = append(data, []string{
			field.Name,
			valueStr,
		})
		if len(field.Name) > longestFieldName {
			longestFieldName = len(field.Name)
		}
	}

	colWidths := []float32{200, float32(longestFieldName) * 30}
	colTypes := []ColumnType{unsortableCol, unsortableCol}
	colProps := []ColumnProps{{false, false}, {false, true}}

	return createNewSortableTable(colWidths, data, colTypes, colProps)
}
