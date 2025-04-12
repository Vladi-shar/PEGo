package main

import (
	"errors"
	"fmt"
	"image"
	"syscall"
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
	path         string
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
		path:         filePath,
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

// --- Windows GDI Structures ---
type BITMAPINFOHEADER struct {
	BiSize          uint32
	BiWidth         int32
	BiHeight        int32
	BiPlanes        uint16
	BiBitCount      uint16
	BiCompression   uint32
	BiSizeImage     uint32
	BiXPelsPerMeter int32
	BiYPelsPerMeter int32
	BiClrUsed       uint32
	BiClrImportant  uint32
}

type RGBQUAD struct {
	Blue     byte
	Green    byte
	Red      byte
	Reserved byte
}

type BITMAPINFO struct {
	BmiHeader BITMAPINFOHEADER
	BmiColors [1]RGBQUAD
}


func convertHICONToImage(hIcon uintptr) (image.Image, error) {
	const width, height = 32, 32
	modgdi32 := syscall.NewLazyDLL("gdi32.dll")
	procCreateCompatibleDC := modgdi32.NewProc("CreateCompatibleDC")
	procCreateDIBSection := modgdi32.NewProc("CreateDIBSection")
	procSelectObject := modgdi32.NewProc("SelectObject")
	procDeleteDC := modgdi32.NewProc("DeleteDC")
	procDeleteObject := modgdi32.NewProc("DeleteObject")

	moduser32 := syscall.NewLazyDLL("user32.dll")
	procDrawIconEx := moduser32.NewProc("DrawIconEx")

	// Create a memory device context.
	hdc, _, err := procCreateCompatibleDC.Call(0)
	if hdc == 0 {
		return nil, fmt.Errorf("CreateCompatibleDC failed: %v", err)
	}
	defer procDeleteDC.Call(hdc)

	var bmi BITMAPINFO
	bmi.BmiHeader.BiSize = uint32(unsafe.Sizeof(bmi.BmiHeader))
	bmi.BmiHeader.BiWidth = width
	bmi.BmiHeader.BiHeight = -height // Negative for a top-down DIB.
	bmi.BmiHeader.BiPlanes = 1
	bmi.BmiHeader.BiBitCount = 32
	bmi.BmiHeader.BiCompression = 0 // BI_RGB (no compression)

	var pvBits unsafe.Pointer
	hBitmap, _, err := procCreateDIBSection.Call(
		hdc,
		uintptr(unsafe.Pointer(&bmi)),
		0, // DIB_RGB_COLORS
		uintptr(unsafe.Pointer(&pvBits)),
		0,
		0,
	)
	if hBitmap == 0 {
		return nil, fmt.Errorf("CreateDIBSection failed: %v", err)
	}
	defer procDeleteObject.Call(hBitmap)

	// Select the DIB section into the DC.
	oldObj, _, err := procSelectObject.Call(hdc, hBitmap)
	if oldObj == 0 {
		return nil, fmt.Errorf("SelectObject failed: %v", err)
	}
	defer procSelectObject.Call(hdc, oldObj)

	// Draw the icon into the DC.
	const DI_NORMAL = 3
	ret, _, err := procDrawIconEx.Call(
		hdc,
		0, 0, // x and y coordinates
		hIcon,           // the icon handle
		uintptr(width),  // width
		uintptr(height), // height
		0,               // step (for animated icons)
		0,               // no background brush
		DI_NORMAL,       // flags
	)
	if ret == 0 {
		return nil, fmt.Errorf("DrawIconEx failed: %v", err)
	}

	// The DIB now holds the icon image in BGRA format.
	size := width * height * 4 // 4 bytes per pixel.
	// Build a slice of bytes backed by the DIB section's memory using unsafe.Slice.
	data := unsafe.Slice((*byte)(pvBits), size)

	// Copy the pixel data to a new slice (because the memory may be freed later).
	pixData := make([]byte, size)
	copy(pixData, data)

	// Convert the BGRA data to RGBA.
	for i := 0; i < size; i += 4 {
		// Swap blue and red channels.
		pixData[i], pixData[i+2] = pixData[i+2], pixData[i]
	}

	// Create the image using the RGBA pixel data.
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	copy(img.Pix, pixData)

	return img, nil
}

func extractExeIcon(filePath string) (image.Image, error) {
	shell32 := syscall.NewLazyDLL("shell32.dll")
	extractIconExW := shell32.NewProc("ExtractIconExW")

	// Load user32.dll for DestroyIcon.
	user32 := syscall.NewLazyDLL("user32.dll")
	destroyIconProc := user32.NewProc("DestroyIcon")

	// Convert exePath to a UTF-16 pointer.
	exePathPtr, err := syscall.UTF16PtrFromString(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to convert exePath: %w", err)
	}

	var hIcon uintptr

	// Extract one large icon (index 0).
	ret, _, callErr := extractIconExW.Call(
		uintptr(unsafe.Pointer(exePathPtr)),
		0, // icon index
		uintptr(unsafe.Pointer(&hIcon)),
		0, // no small icon requested
		1, // number of icons to extract
	)
	if ret == 0 || hIcon == 0 {
		return nil, fmt.Errorf("ExtractIconExW failed: %v", callErr)
	}

	// Convert the HICON to image.Image.
	img, err := convertHICONToImage(hIcon)
	// Clean up: free the HICON.
	destroyIconProc.Call(hIcon)
	if err != nil {
		return nil, fmt.Errorf("failed to convert HICON: %w", err)
	}
	return img, nil
}
