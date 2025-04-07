package main

import (
	"errors"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	ENGLISH    = 0x0409
	CP_UNICODE = 0x04b0
	CP_USASCII = 0x04e4
)

type LANGANDCODEPAGE struct {
	wLanguage uint16
	wCodePage uint16
}

type FileResourceCollector struct {
	data         []byte
	translations []LANGANDCODEPAGE
}

var (
	versionDll              = windows.NewLazySystemDLL("version.dll")
	getFileVersionInfoSizeW = versionDll.NewProc("GetFileVersionInfoSizeW")
	getFileVersionInfoW     = versionDll.NewProc("GetFileVersionInfoW")
	verQueryValueW          = versionDll.NewProc("VerQueryValueW")
)

func BuildTranslations(data []byte) ([]LANGANDCODEPAGE, error) {
	var translations []LANGANDCODEPAGE
	if len(data) == 0 {
		return translations, errors.New("empty data")
	}

	subBlock, err := windows.UTF16PtrFromString(`\VarFileInfo\Translation`)
	if err != nil {
		return translations, err
	}
	var translate *LANGANDCODEPAGE
	var translateSize uint32
	ret, _, err := verQueryValueW.Call(
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(unsafe.Pointer(subBlock)),
		uintptr(unsafe.Pointer(&translate)),
		uintptr(unsafe.Pointer(&translateSize)),
	)
	if ret == 0 {
		return translations, err
	}

	numOfTranslations := translateSize / uint32(unsafe.Sizeof(LANGANDCODEPAGE{}))

	entries := (*[1 << 20]LANGANDCODEPAGE)(unsafe.Pointer(translate))[:numOfTranslations:numOfTranslations]
	for _, entry := range entries {
		translations = append(translations, entry)
		translations = append(translations, LANGANDCODEPAGE{wLanguage: entry.wLanguage, wCodePage: CP_UNICODE})
		translations = append(translations, LANGANDCODEPAGE{wLanguage: entry.wLanguage, wCodePage: CP_USASCII})
		translations = append(translations, LANGANDCODEPAGE{wLanguage: ENGLISH, wCodePage: entry.wCodePage})
	}

	translations = append(translations, LANGANDCODEPAGE{wLanguage: ENGLISH, wCodePage: CP_USASCII})
	translations = append(translations, LANGANDCODEPAGE{wLanguage: ENGLISH, wCodePage: CP_USASCII})

	return translations, nil
}

func NewFileResourceCollector(filePath string) (*FileResourceCollector, error) {
	path, err := windows.UTF16PtrFromString(filePath)
	if err != nil {
		return nil, err
	}

	pathPtr := uintptr(unsafe.Pointer(path))

	size, _, err := getFileVersionInfoSizeW.Call(pathPtr, 0)
	if size == 0 {
		return nil, err
	}

	data := make([]byte, size)

	ret, _, err := getFileVersionInfoW.Call(pathPtr, 0, size, uintptr(unsafe.Pointer(&data[0])))
	if ret == 0 {
		return nil, err
	}

	translations, err := BuildTranslations(data)
	if err != nil {
		return nil, err
	}

	return &FileResourceCollector{
		data:         data,
		translations: translations,
	}, nil

}

func QueryValue(translate LANGANDCODEPAGE, data []byte, resourceName string) (string, error) {
	if len(data) == 0 || len(resourceName) == 0 {
		return "", errors.New("empty data")
	}

	resourceString := fmt.Sprintf(`\StringFileInfo\%04X%04X\%s`, translate.wLanguage, translate.wCodePage, resourceName)

	resourceStringPtr, err := windows.UTF16PtrFromString(resourceString)
	if err != nil {
		return "", err
	}

	var output uintptr
	var outputSize uint32
	ret, _, err := verQueryValueW.Call(
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(unsafe.Pointer(resourceStringPtr)),
		uintptr(unsafe.Pointer(&output)),
		uintptr(unsafe.Pointer(&outputSize)),
	)
	if ret == 0 {
		return "", err
	}

	return windows.UTF16ToString((*[1 << 20]uint16)(unsafe.Pointer(output))[:outputSize]), nil
}

func (frc *FileResourceCollector) GetResource(resourceName string) string {
	for _, translation := range frc.translations {
		res, err := QueryValue(translation, frc.data, resourceName)
		if err == nil && len(res) != 0 {
			return res
		}
	}
	return ""
}
