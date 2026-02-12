package archive

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

type FileEntry struct {
	Name string
	Data []byte
}

func CreateDirArchive(out io.Writer, dirPath string) error {
	files, err := readDir(dirPath)
	if err != nil {
		return err
	}

	err = CreateArchive(out, files)
	if err != nil {
		return err
	}

	return nil
}

func readDir(dir string) ([]string, error) {
	contents, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	files := make([]string, 0)

	for _, entry := range contents {
		if entry.IsDir() {
			sub, err := readDir(filepath.Join(dir, entry.Name()))
			if err != nil {
				return nil, errors.New(fmt.Sprintf("Failed to read dir: %s, %v", entry.Name(), err))
			}
			files = append(files, sub...)
		} else {
			if entry.Name() != "" {
				files = append(files, filepath.Join(dir, entry.Name()))
			}
		}
	}

	return files, nil
}

func CreateArchive(out io.Writer, files []string) error {
	gw := gzip.NewWriter(out)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	for _, file := range files {
		err := addToArchive(tw, file)
		if err != nil {
			return errors.New(fmt.Sprintf("Failed to add %s to archive: %v", file, err))
		}
	}

	return nil
}

func CreateArchiveEnt(out io.Writer, files []FileEntry) error {
	gw := gzip.NewWriter(out)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	for _, entry := range files {
		err := addToArchiveEnt(tw, entry)
		if err != nil {
			return err
		}
	}

	return nil
}

func addToArchive(tw *tar.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to open %s: %v", filename, err))
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}

	header.Name = filename

	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(tw, file)
	if err != nil {
		return err
	}

	return nil
}

func addToArchiveEnt(tw *tar.Writer, file FileEntry) error {
	reader := bytes.NewReader(file.Data)

	header := &tar.Header{
		Name:     file.Name,
		Mode:     0644,
		Size:     int64(len(file.Data)),
		ModTime:  time.Now(),
		Typeflag: tar.TypeReg,
	}

	err := tw.WriteHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(tw, reader)
	if err != nil {
		return err
	}

	return nil
}

func ReadArchiveEnt(stream io.Reader) []FileEntry {
	uncompressedStream, err := gzip.NewReader(stream)
	if err != nil {
		log.Fatalf("Read Archive: NewReader failed")
	}

	tarReader := tar.NewReader(uncompressedStream)

	files := make([]FileEntry, 0)

	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("ReadArchiveEnt: Next() failed: %s", err.Error())
		}

		switch header.Typeflag {
		case tar.TypeReg:
			var out = &bytes.Buffer{}
			if _, err := io.Copy(out, tarReader); err != nil {
				log.Fatalf("ReadArchiveEnt: Copy() failed: %s", err.Error())
			}
			files = append(files, FileEntry{header.Name, out.Bytes()})
		default:
			log.Fatalf("ReadArchiveEnt: Unsupported File Structure")
		}
	}

	return files
}
