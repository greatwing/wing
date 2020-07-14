package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

var protoDir string

func init() {
	flag.StringVar(&protoDir, "i", "./proto", "protocol files dir")
}

func walkFunc(path string, info os.FileInfo, err error) error {
	if info.IsDir() {
		return nil
	}

	fileSuffix := filepath.Ext(info.Name())
	if fileSuffix == ".proto" {
		fmt.Printf("file: %s\n", path)
		cmd := exec.Command("protoc",
			"--gogofaster_out", protoDir,
			"--proto_path", protoDir,
			"--proto_path", filepath.Join(protoDir, "/thirdparty"),
			info.Name())
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println(string(out), err)
			return err
		}
		fmt.Println(string(out))
	}

	return nil
}

func main() {
	flag.Parse()

	infos, err := ioutil.ReadDir(protoDir)
	for _, info := range infos {
		walkFunc(filepath.Join(protoDir, info.Name()), info, nil)
	}
	//err := filepath.Walk(protoDir, walkFunc)
	if err != nil {
		fmt.Print(err)
	}
}
