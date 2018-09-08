package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
)

var (
	directory string
	host      string
	step      int
)

func init() {
	flag.StringVar(&directory, "d", ".", "directory")
	flag.StringVar(&host, "h", "192.168.0.150", "target host address")
	flag.IntVar(&step, "s", 32, "upload steps")
}

func main() {

	flag.Parse()

	done := make(chan error)
	target := "http://" + host + "/upload.json"

	fmt.Printf("target=%s\n", target)

	d, err := os.Open(directory)
	if err != nil {
		panic(err)
	}
	defer d.Close()

	for {
		finfo, err := d.Readdir(step)
		if err != nil {
			if err != io.EOF {
				panic(err)
			}
			break
		}

		var files int

		for i := 0; i < len(finfo); i++ {
			file := finfo[i]
			if !file.IsDir() {

				form := fmt.Sprintf("files[]=@%s", file.Name())
				cmd := exec.Command("/usr/bin/curl", "-F", form, target)
				files++

				err := cmd.Start()
				if err != nil {
					panic(err)
				}

				fmt.Printf("+WIP: %5s\n", file.Name())

				go func() {
					done<-cmd.Wait()
				}()
			}
		}


		for i := 0; i < files; i++ {
			select {
			case err:= <-done:
				if err != nil {
					panic(err)
				}
			}
		}
	}

	fmt.Println("done")
}
