package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/schollz/progressbar/v3"
)

// InstallVersion downloads and installs the specified Go version
func InstallVersion(version Version) error {
	fmt.Printf("\nStarting installation of Go %s...\n", FormatVersion(version.Version))

	file, err := GetFile(version)
	if err != nil {
		return err
	}

	downloadURL := fmt.Sprintf("https://go.dev/dl/%s", file.Filename)

	fmt.Printf("Downloading: %s\n", downloadURL)
	fmt.Printf("Size: %.2f MB\n", float64(file.Size)/(1024*1024))

	tmpDir := os.TempDir()
	downloadPath := filepath.Join(tmpDir, file.Filename)

	if err := downloadFile(downloadURL, downloadPath, file.Size); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	fmt.Println("Verifying checksum...")
	if err := verifyChecksum(downloadPath, file.SHA256); err != nil {
		_ = os.Remove(downloadPath)
		return fmt.Errorf("checksum verification failed: %w", err)
	}
	fmt.Println("✓ Checksum verified")

	fmt.Println("Installing...")
	if err := extractAndInstall(downloadPath); err != nil {
		_ = os.Remove(downloadPath)
		return fmt.Errorf("installation failed: %w", err)
	}

	_ = os.Remove(downloadPath)

	fmt.Printf("\n✓ Go %s installed successfully!\n", FormatVersion(version.Version))
	fmt.Println("\nTo use the new version, restart your terminal or run:")
	fmt.Println("  source ~/.zshrc  # or ~/.bashrc")

	return nil
}

func downloadFile(url, destPath string, size int64) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	bar := progressbar.NewOptions64(
		size,
		progressbar.OptionSetDescription("Downloading"),
		progressbar.OptionSetWidth(40),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
	fmt.Println()
	return err
}

func verifyChecksum(filePath, expectedHash string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return err
	}

	actualHash := hex.EncodeToString(hasher.Sum(nil))
	if actualHash != expectedHash {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedHash, actualHash)
	}

	return nil
}

func extractAndInstall(archivePath string) error {
	goRoot := "/usr/local/go"

	needSudo := false
	if _, err := os.Stat("/usr/local"); err == nil {
		testFile := "/usr/local/.update-go-test"
		if f, err := os.Create(testFile); err != nil {
			needSudo = true
		} else {
			_ = f.Close()
			_ = os.Remove(testFile)
		}
	}

	if _, err := os.Stat(goRoot); err == nil {
		fmt.Println("Backing up existing Go installation...")
		backupPath := goRoot + ".backup"

		if needSudo {
			_ = exec.Command("sudo", "rm", "-rf", backupPath).Run()
			if err := exec.Command("sudo", "mv", goRoot, backupPath).Run(); err != nil {
				return fmt.Errorf("backup failed: %w", err)
			}
		} else {
			_ = os.RemoveAll(backupPath)
			if err := os.Rename(goRoot, backupPath); err != nil {
				return fmt.Errorf("backup failed: %w", err)
			}
		}
	}

	fmt.Println("Extracting archive...")

	var cmd *exec.Cmd
	ext := filepath.Ext(archivePath)

	switch runtime.GOOS {
	case "darwin", "linux":
		if ext == ".gz" || filepath.Ext(filepath.Base(archivePath[:len(archivePath)-3])) == ".tar" {
			if needSudo {
				cmd = exec.Command("sudo", "tar", "-C", "/usr/local", "-xzf", archivePath)
			} else {
				cmd = exec.Command("tar", "-C", "/usr/local", "-xzf", archivePath)
			}
		}
	case "windows":
		return fmt.Errorf("automatic installation on Windows is not currently supported")
	}

	if cmd == nil {
		return fmt.Errorf("unsupported archive format: %s", ext)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		backupPath := goRoot + ".backup"
		if _, statErr := os.Stat(backupPath); statErr == nil {
			if needSudo {
				_ = exec.Command("sudo", "mv", backupPath, goRoot).Run()
			} else {
				_ = os.Rename(backupPath, goRoot)
			}
		}
		return fmt.Errorf("extraction failed: %w", err)
	}

	backupPath := goRoot + ".backup"
	if needSudo {
		_ = exec.Command("sudo", "rm", "-rf", backupPath).Run()
	} else {
		_ = os.RemoveAll(backupPath)
	}

	newGoPath := filepath.Join(goRoot, "bin", "go")
	if _, err := os.Stat(newGoPath); os.IsNotExist(err) {
		return fmt.Errorf("installation verification failed: go binary not found")
	}

	out, err := exec.Command(newGoPath, "version").Output()
	if err != nil {
		return fmt.Errorf("failed to verify version: %w", err)
	}
	fmt.Printf("Installation complete: %s", string(out))

	return nil
}
