package util

import (
	"archive/zip"
	"github.com/eyebluecn/tank/code/tool/result"
	"io"
	"os"
	"path/filepath"
	"strings"
)

//zip srcPath to destPath。
func Zip(srcPath string, destPath string) error {

	srcPath = UniformPath(srcPath)

	if PathExists(destPath) {
		panic(result.BadRequest("%s exits", destPath))
	}

	fileWriter, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer func() {
		err := fileWriter.Close()
		if err != nil {
			panic(err)
		}
	}()

	zipWriter := zip.NewWriter(fileWriter)
	defer func() {

		if err := zipWriter.Close(); err != nil {
			panic(err)
		}
	}()

	baseDirPath := GetDirOfPath(srcPath) + "/"
	err = filepath.Walk(srcPath, func(path string, fileInfo os.FileInfo, errBack error) (err error) {
		if errBack != nil {
			return errBack
		}

		path = UniformPath(path)

		fileHeader, err := zip.FileInfoHeader(fileInfo)
		if err != nil {
			return
		}

		fileHeader.Name = strings.TrimPrefix(path, baseDirPath)

		// directory need /
		if fileInfo.IsDir() {
			fileHeader.Name += "/"
		}

		writer, err := zipWriter.CreateHeader(fileHeader)
		if err != nil {
			return
		}

		//only regular has things to write.
		if !fileHeader.Mode().IsRegular() {
			return nil
		}

		fileToBeZip, err := os.Open(path)
		defer func() {
			err = fileToBeZip.Close()
			if err != nil {
				panic(err)
			}
		}()
		if err != nil {
			return
		}

		_, err = io.Copy(writer, fileToBeZip)
		if err != nil {
			return
		}

		return nil
	})
	return err
}
