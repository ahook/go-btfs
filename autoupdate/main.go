package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"

	shell "github.com/ipfs/go-ipfs-api"
)

const url = "localhost:50051"

// Rollback function of auto update.
func rollback(wg sync.WaitGroup, defaultProjectPath, defaultDownloadPath string) {
	defer func() {
		wg.Done()
	}()

	// Check if the BTFS daemon server is up every 5 seconds, checked a total of five times.
	for i := 0; i < 5; i++ {
		time.Sleep(time.Second * 5)
		// Where your local node is running on localhost:5001
		sh := shell.NewShell(url)
		if sh.IsUp() {
			fmt.Println("BTFS auto update SUCCESS!")
			return
		}
	}

	fmt.Println("BTFS rollback begin!")

	var currentConfigPath string
	var btfsBackupPath string
	var btfsBinaryPath string
	var backupConfigPath string

	// Select binary files based on operating system.
	if (runtime.GOOS == "darwin" || runtime.GOOS == "linux" || runtime.GOOS == "windows") && (runtime.GOARCH == "amd64" || runtime.GOARCH == "386") {
		ext := ""
		if runtime.GOOS == "windows" {
			ext = ".exe"
		}

		currentConfigPath = fmt.Sprint(defaultProjectPath, "config.yaml")
		backupConfigPath = fmt.Sprint(defaultDownloadPath, "config.yaml.bk")
		btfsBackupPath = fmt.Sprint(defaultDownloadPath, fmt.Sprintf("btfs%s.bk", ext))
		btfsBinaryPath = fmt.Sprint(defaultProjectPath, fmt.Sprintf("btfs%s", ext))
	} else {
		fmt.Printf("Operating system [%s], arch [%s] does not support rollback\n", runtime.GOOS, runtime.GOARCH)
		return
	}

	var err error

	// Check if the backup binary file exists.
	if !pathExists(btfsBackupPath) {
		fmt.Printf("BTFS backup binary is not exists.")
		return
	}

	// Check if the current configure file exists.
	if pathExists(currentConfigPath) {
		// Delete current configure file.
		err = os.Remove(currentConfigPath)
		if err != nil {
			fmt.Printf("Delete backup configure file error, reasons: [%v]\n", err)
			return
		}
	}

	// Check if the backup configure file exists.
	if pathExists(backupConfigPath) {
		// Move backup configure file to current configure file.
		err = os.Rename(backupConfigPath, currentConfigPath)
		if err != nil {
			fmt.Printf("Move backup configure file error, reasons: [%v]\n", err)
			return
		}
	}

	// Check if the btfs binary file exists.
	if pathExists(btfsBinaryPath) {
		// Delete the btfs binary file.
		err = os.Remove(btfsBinaryPath)
		if err != nil {
			fmt.Printf("Delete btfs binary file error, reasons: [%v]\n", err)
			return
		}
	}

	// Move backup btfs binary file to current btfs binary file.
	err = os.Rename(btfsBackupPath, btfsBinaryPath)
	if err != nil {
		fmt.Printf("Move backup btfs binary file error, reasons: [%v]\n", err)
		return
	}

	// Add executable permissions to btfs binary.
	err = os.Chmod(btfsBinaryPath, 0775)
	if err != nil {
		fmt.Printf("Chmod file error, reasons: [%v]\n", err)
		return
	}

	// Start the btfs daemon according to different operating systems.
	if runtime.GOOS == "windows" {
		cmd := exec.Command(btfsBinaryPath, "daemon")
		err = cmd.Start()
	} else {
		fmt.Println(btfsBinaryPath)
		cmd := exec.Command("sudo", btfsBinaryPath, "daemon")
		err = cmd.Start()
	}

	// Check if the btfs daemon start success.
	if err != nil {
		fmt.Printf("BTFS rollback failed, reasons: [%v]", err)
		return
	}

	fmt.Println("BTFS rollback SUCCESS!")
}

func main() {
	time.Sleep(time.Second * 5)
	defaultProjectPath := flag.String("project", "", "default project path")
	defaultDownloadPath := flag.String("download", "", "default download path")

	flag.Parse()

	if *defaultProjectPath == "" || *defaultDownloadPath == "" {
		fmt.Println("Request param is nil.")
		return
	}

	var currentConfigPath string
	var latestConfigPath string
	var btfsBackupPath string
	var btfsBinaryPath string
	var latestBtfsBinaryPath string
	var backupConfigPath string

	// Select binary files based on operating system.
	if (runtime.GOOS == "darwin" || runtime.GOOS == "linux" || runtime.GOOS == "windows") && (runtime.GOARCH == "amd64" || runtime.GOARCH == "386") {
		ext := ""
		if runtime.GOOS == "windows" {
			ext = ".exe"
		}

		currentConfigPath = fmt.Sprint(*defaultProjectPath, "config.yaml")
		backupConfigPath = fmt.Sprint(*defaultDownloadPath, "config.yaml.bk")
		latestConfigPath = fmt.Sprint(*defaultDownloadPath, fmt.Sprintf("config_%s_%s.yaml", runtime.GOOS, runtime.GOARCH))
		btfsBackupPath = fmt.Sprint(*defaultDownloadPath, fmt.Sprintf("btfs%s.bk", ext))
		btfsBinaryPath = fmt.Sprint(*defaultProjectPath, fmt.Sprintf("btfs%s", ext))
		latestBtfsBinaryPath = fmt.Sprint(*defaultDownloadPath, fmt.Sprintf("btfs-%s-%s%s", runtime.GOOS, runtime.GOARCH, ext))
	} else {
		fmt.Printf("Operating system [%s], arch [%s] does not support automatic updates\n", runtime.GOOS, runtime.GOARCH)
		return
	}

	var err error

	// Delete backup configure file.
	if pathExists(backupConfigPath) {
		err = os.Remove(backupConfigPath)
		if err != nil {
			fmt.Printf("Delete backup config file error, reasons: [%v]\n", err)
			return
		}
	}

	// Move current config file if existed.
	if pathExists(currentConfigPath) {
		err = os.Rename(currentConfigPath, backupConfigPath)
		if err != nil {
			fmt.Printf("Move current config file error, reasons: [%v]\n", err)
			return
		}
	}

	// Move latest configure file to current configure file.
	err = os.Rename(latestConfigPath, currentConfigPath)
	if err != nil {
		fmt.Printf("Move file error, reasons: [%v]\n", err)
		return
	}

	// Delete btfs backup file.
	if pathExists(btfsBackupPath) {
		err = os.Remove(btfsBackupPath)
		if err != nil {
			fmt.Printf("Move file error, reasons: [%v]\n", err)
			return
		}
	}

	// Backup btfs binary file.
	err = os.Rename(btfsBinaryPath, btfsBackupPath)
	if err != nil {
		fmt.Printf("Move file error, reasons: [%v]\n", err)
		return
	}

	// Move latest btfs binary file to current btfs binary file.
	err = os.Rename(latestBtfsBinaryPath, btfsBinaryPath)
	if err != nil {
		fmt.Printf("Move file error, reasons: [%v]\n", err)
		return
	}

	// Add executable permissions to btfs binary.
	err = os.Chmod(btfsBinaryPath, 0775)
	if err != nil {
		fmt.Printf("Chmod file error, reasons: [%v]\n", err)
		return
	}

	wg := sync.WaitGroup{}

	wg.Add(1)

	go rollback(wg, *defaultProjectPath, *defaultDownloadPath)

	if runtime.GOOS == "windows" {
		cmd := exec.Command(btfsBinaryPath, "daemon")
		err = cmd.Start()
	} else {
		fmt.Println(btfsBinaryPath)
		cmd := exec.Command("sudo", btfsBinaryPath, "daemon")
		err = cmd.Start()
	}

	// Wait for the rollback program to complete.
	wg.Wait()
}

// Determine if the path file exists.
func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}
