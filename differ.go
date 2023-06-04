package main

import (
	"archive/tar"
	"fmt"
	"io"
	"log"
	"os"
)

type FileDiff struct {
	Name        string
	SizeDiff    bool
	ContentDiff bool
}

func compareTarFiles(file1, file2 string) ([]FileDiff, error) {
	file1Reader, err := os.Open(file1)
	if err != nil {
		return nil, fmt.Errorf("failed to open file1: %w", err)
	}
	defer file1Reader.Close()

	file2Reader, err := os.Open(file2)
	if err != nil {
		return nil, fmt.Errorf("failed to open file2: %w", err)
	}
	defer file2Reader.Close()

	tarReader1 := tar.NewReader(file1Reader)
	tarReader2 := tar.NewReader(file2Reader)

	var diffs []FileDiff

	for {
		header1, err1 := tarReader1.Next()
		if err1 != nil {
			if err1 == io.EOF {
				break
			}
			return nil, fmt.Errorf("error reading file1 tar: %w", err1)
		}

		header2, err2 := tarReader2.Next()
		if err2 != nil {
			if err2 == io.EOF {
				break
			}
			return nil, fmt.Errorf("error reading file2 tar: %w", err2)
		}

		// Compare the file headers
		if header1.Name != header2.Name || header1.Size != header2.Size {
			// Differences found in file headers
			diffs = append(diffs, FileDiff{Name: header1.Name, SizeDiff: true})
			continue
		}

		// Compare file contents
		if err := compareFileContents(tarReader1, tarReader2); err != nil {
			diffs = append(diffs, FileDiff{Name: header1.Name, ContentDiff: true})
		}
	}

	return diffs, nil
}

func compareFileContents(reader1, reader2 io.Reader) error {
	buf1 := make([]byte, 8192)
	buf2 := make([]byte, 8192)

	for {
		n1, err1 := reader1.Read(buf1)
		n2, err2 := reader2.Read(buf2)

		if err1 == io.EOF && err2 == io.EOF {
			// Reached the end of both file contents
			break
		}

		if err1 != nil || err2 != nil || n1 != n2 {
			// Differences found in file contents
			return fmt.Errorf("differences found in file contents")
		}

		if !byteSlicesEqual(buf1[:n1], buf2[:n2]) {
			// Differences found in file contents
			return fmt.Errorf("differences found in file contents")
		}
	}

	return nil
}

func byteSlicesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func main() {
	// file1 := "/home/albi/forensicImages/smart-home/szenen/00_Default/00_Default_HA.tar"
	file1 := "/home/albi/forensicImages/smart-home/szenen/02_Scene/02_Scene_HA.tar"
	file2 := "/home/albi/forensicImages/smart-home/szenen/01_Configured/01_Configured_HA.tar"

	diffs, err := compareTarFiles(file1, file2)
	if err != nil {
		log.Fatalf("Error comparing tar files: %v", err)
	}

	if len(diffs) == 0 {
		fmt.Println("The contents of the tar files are equal.")
	} else {
		fmt.Println("The following files are different:")
		for _, diff := range diffs {
			if diff.SizeDiff {
				fmt.Printf(" %s - Size differs\n", diff.Name)
			}
			if diff.ContentDiff {
				fmt.Printf(" %s  - Content differs\n", diff.Name)
			}
		}
	}
}
