package nes

import "io"
import (
	"archive/zip"
	"encoding/binary"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

const NESMagicNumber = 0x1a53454e // "NES^Z"
const ZIPMagicNumber = 0x04034B50 // "PK.."

func ReadFile(path string) string {
	file, err := os.Open(path)
	if err != nil {
		log.Printf("Readfile : %v", err)
	}
	header, err := ReadMagicNumber(file)
	if err != nil {
		log.Printf("Readfile : %v", err)
	}
	switch header {
	case NESMagicNumber:
		return path
	case ZIPMagicNumber:
		return Zip(path)
	}
	return ""
}

func ReadMagicNumber(w io.Reader) (uint32, error) {
	var header uint32
	if err := binary.Read(w, binary.LittleEndian, &header); err != nil {
		return 0, err
	}
	return header, nil
}

func isNES(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	header, err := ReadMagicNumber(file)
	return header == NESMagicNumber
}

func Zip(path string) string {
	tmpdir := "tmp/"

	err := unzip(path, tmpdir)
	if err != nil {
		log.Fatal(err)
		return ""
	}
	dir, err := ioutil.ReadDir(tmpdir)
	if err != nil {
		log.Fatal(err)
		return ""
	}
	for _, file := range dir {
		log.Print(file.Name())
		if isNES(tmpdir + file.Name()) {
			return tmpdir + file.Name()
		}
	}
	return ""
}

func unzip(archive, target string) error {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(target, 0755); err != nil {
		return err
	}

	for _, file := range reader.File {
		path := filepath.Join(target, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			return err
		}
	}

	return nil
}
