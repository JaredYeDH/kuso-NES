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

const Tmpdir = "tmp"
const NESMagicNumber = 0x1a53454e // "NES^Z"
const ZIPMagicNumber = 0x04034B50 // "PK.."

func ReadFile(path string) (string, bool) {
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
		return path, false
	case ZIPMagicNumber:
		return Zip(path), true
	}
	return "", false
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

func findNES(dirname string) string {
	dir, err := ioutil.ReadDir(dirname)
	if err != nil {
		log.Printf("Find nes in %v : %v", dirname, err)
	}
	for _, file := range dir {
		if file.IsDir() {
			res := findNES(dirname + "/" + file.Name())
			if res != "" {
				return res
			}
		}
		if isNES(dirname + "/" + file.Name()) {
			return dirname + "/" + file.Name()
		}
	}
	log.Fatal("Can't find any nes rom file in the zip.")
	return ""
}

func RemoveDir(path string) error {
	dir, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}
	for _, file := range dir {
		if file.IsDir() {
			RemoveDir(path + "/" + file.Name())
		} else {
			err := os.Remove(path + "/" + file.Name())
			if err != nil {
				return err
			}
		}
	}
	err = os.Remove(path)
	return err
}

func Zip(path string) string {

	err := unzip(path, Tmpdir)
	if err != nil {
		log.Fatal(err)
		return ""
	}
	return findNES(Tmpdir)
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
