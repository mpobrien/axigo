package main

import (
	"encoding/base64"
	"io"
	"strings"
)

// func main() {
// 	if len(os.Args) < 2 {
// 		//TODO: print usage
// 		fmt.Println("Please provide at least one argument")
// 		os.Exit(1)
// 	}

// 	pipeReader, pipeWriter := io.Pipe()

// 	go func() {
// 		defer pipeWriter.Close()
// 		err := writeImagesFromPaths(os.Args[1:], pipeWriter)
// 		if err != nil {
// 			panic(err)
// 		}
// 	}()

// 	_, err := io.Copy(os.Stdout, pipeReader)
// 	if err != nil {
// 		panic(err)
// 	}
// }

// func writeImagesFromPaths(args []string, writer io.Writer) error {
// 	for _, filepath := range args {
// 		err := writeInlineImage(filepath, writer)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// func writeInlineImage(filepath string, writer io.Writer) error {
// 	file, err := getFile(filepath)
// 	if err != nil {
// 		return err
// 	}
// 	err = imgcat(file, writer)
// 	if err != nil {
// 		return err
// 	}
// 	defer file.Close()

// 	return nil
// }

// func getFile(filepath string) (file *os.File, err error) {
// 	file, _err := os.Open(filepath)
// 	if _err != nil {
// 		return nil, _err
// 	}
// 	return file, nil
// }

func imgcat(imgdata io.Reader, out io.Writer) error {
	_, err := io.Copy(out, strings.NewReader("\033]1337;File=inline=1:"))
	if err != nil {
		return err
	}
	wc := base64.NewEncoder(base64.StdEncoding, out)
	if _, err := io.Copy(wc, imgdata); err != nil {
		return err
	}
	_, err = io.WriteString(out, "\a\n")
	return err
}

// func writeAsBase64(reader io.Reader, writer io.Writer) error {
// 	wc := base64.NewEncoder(base64.StdEncoding, writer)
// 	_, err := io.Copy(wc, reader)
// 	if err != nil {
// 		return err
// 	}
// 	return wc.Close()
// }
