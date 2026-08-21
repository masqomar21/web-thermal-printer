package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type Target struct {
	OS   string
	Arch string
	Ext  string
}

func main() {
	log.Println("🚀 Starting Cross-Platform Build Process for printerRunner...")

	targets := []Target{
		{OS: "windows", Arch: "amd64", Ext: ".exe"},
		{OS: "windows", Arch: "386", Ext: ".exe"},
		{OS: "linux", Arch: "amd64", Ext: ""},
		{OS: "linux", Arch: "arm64", Ext: ""},
		{OS: "darwin", Arch: "arm64", Ext: ""},
		{OS: "darwin", Arch: "amd64", Ext: ""},
	}

	distDir := "dist"
	if err := os.MkdirAll(distDir, 0755); err != nil {
		log.Fatalf("❌ Failed to create dist directory: %v", err)
	}

	tempDir, err := os.MkdirTemp("", "printerRunner-build-*")
	if err != nil {
		log.Fatalf("❌ Failed to create temp build directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	for _, target := range targets {
		binaryName := "printerRunner" + target.Ext
		testBinaryName := "printerTest" + target.Ext
		outputPath := filepath.Join(tempDir, binaryName)
		testOutputPath := filepath.Join(tempDir, testBinaryName)

		log.Printf("🔨 Building for %s/%s -> %s & %s...", target.OS, target.Arch, binaryName, testBinaryName)

		cmd := exec.Command("go", "build", "-trimpath", "-ldflags", "-s -w", "-o", outputPath, "./cmd")
		cmd.Env = append(os.Environ(),
			"GOOS="+target.OS,
			"GOARCH="+target.Arch,
			"CGO_ENABLED=0",
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			log.Fatalf("❌ Build failed for %s/%s (%s): %v", target.OS, target.Arch, binaryName, err)
		}

		cmdTest := exec.Command("go", "build", "-trimpath", "-ldflags", "-s -w", "-o", testOutputPath, "./cmd/test-printer")
		cmdTest.Env = append(os.Environ(),
			"GOOS="+target.OS,
			"GOARCH="+target.Arch,
			"CGO_ENABLED=0",
		)
		cmdTest.Stdout = os.Stdout
		cmdTest.Stderr = os.Stderr

		if err := cmdTest.Run(); err != nil {
			log.Fatalf("❌ Build failed for %s/%s (%s): %v", target.OS, target.Arch, testBinaryName, err)
		}

		zipName := fmt.Sprintf("printerRunner-%s-%s.zip", target.OS, target.Arch)
		zipPath := filepath.Join(distDir, zipName)

		log.Printf("📦 Packaging %s...", zipName)

		filesToZip := map[string]string{
			binaryName:            outputPath,
			testBinaryName:        testOutputPath,
			".env.example":        ".env.example",
		}

		if err := createZip(zipPath, filesToZip); err != nil {
			log.Fatalf("❌ Failed to create zip for %s/%s: %v", target.OS, target.Arch, err)
		}

		log.Printf("✅ Release bundle created: %s", zipPath)
	}

	log.Println("🎉 All release bundles generated successfully in dist/ directory!")
}

func createZip(zipPath string, files map[string]string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	for archiveName, filePath := range files {
		info, err := os.Stat(filePath)
		if err != nil {
			// If optional file doesn't exist, skip gracefully
			if os.IsNotExist(err) {
				continue
			}
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = archiveName
		header.Method = zip.Deflate

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		file, err := os.Open(filePath)
		if err != nil {
			return err
		}

		_, err = io.Copy(writer, file)
		file.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
