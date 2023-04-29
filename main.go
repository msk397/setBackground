package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"
)

const (
	SPI_SETDESKWALLPAPER     = 20
	SPIF_UPDATEINIFILE       = 0x01
	SPIF_SENDWININICHANGE    = 0x02
	SPI_SETDESKWALLPAPERHORZ = 0x0013
)

func main() {
	// 每天早上8点请求服务器最新的一张图片
	t := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-t.C:
			go downloadAndSetWallpaper()
		}
	}
}

func downloadAndSetWallpaper() {
	// 发送请求获取最新的一张图片
	resp, err := http.Get("http://localhost:8080/download")
	if err != nil || resp.StatusCode != http.StatusOK {
		fmt.Println("请求失败：", err)
		return
	}
	defer resp.Body.Close()

	// 保存图片
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取响应失败：", err)
		return
	}
	filename := fmt.Sprintf("%d.png", time.Now().UnixNano())
	f, err := os.Create(filename)
	if err != nil {
		fmt.Println("保存文件失败：", err)
		return
	}
	f.Write(body)
	f.Close()
	// 设置桌面背景
	err = setWallpaper(filename)
	if err != nil {
		fmt.Println("设置桌面背景失败：", err)
		return
	}
	err = os.Remove(filename)
}

func setWallpaper(filename string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	// 构建壁纸路径
	imagePath := filepath.Join(dir, filename)

	primaryImagePath := imagePath

	// 设置显示器的桌面背景
	ret, _, err := syscall.NewLazyDLL("user32.dll").NewProc("SystemParametersInfoW").Call(
		SPI_SETDESKWALLPAPER,
		uintptr(0),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(primaryImagePath))),
		uintptr(SPIF_UPDATEINIFILE|SPIF_SENDWININICHANGE),
	)
	if ret == 0 {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Desktop background for primary monitor changed successfully.")
	}

	return nil
}
