package main

import (
	"flag"
	"fmt"
	"github.com/katie31/wal-g"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var nop bool
var s3 bool
var outDir string

func init() {
	flag.BoolVar(&nop, "n", false, "Use a NOP writer (false on default).")
	flag.BoolVar(&s3, "s", false, "Upload compressed tar files to s3 (write to disk on default)")
	flag.StringVar(&outDir, "out", "", "Directory compressed tar files will be written to (unset on default)")
}

func main() {
	flag.Parse()
	all := flag.Args()
	part, err := strconv.Atoi(all[0])
	if err != nil {
		panic(err)
	}
	in := all[1]

	bundle := &walg.Bundle{
		MinSize: int64(part),
	}

	if nop {
		bundle.Tbm = &walg.NOPTarBallMaker{
			BaseDir: filepath.Base(in),
			Trim:    in,
			Nop:     true,
		}
	} else if !s3 && outDir == "" {
		fmt.Printf("Please provide a directory to write to.\n")
		os.Exit(1)
	} else if !s3 {
		c, err := walg.Connect()
		if err != nil {
			panic(err)
		}
		lab, _ := walg.QueryFile(c, "hello")
		bkout := filepath.Join(outDir, walg.FormatName(lab))

		bundle.Tbm = &walg.FileTarBallMaker{
			BaseDir: filepath.Base(in),
			Trim:    in,
			Out:     bkout,
		}
		os.MkdirAll(bkout, 0766)

	} else if s3 {
		c, err := walg.Connect()
		if err != nil {
			panic(err)
		}
		lbl, sc := walg.QueryFile(c, "hello")
		n := walg.FormatName(lbl)

		bundle.Tbm = &walg.S3TarBallMaker{
			BaseDir:  filepath.Base(in),
			Trim:     in,
			BkupName: n,
			Tu:       walg.Configure(),
		}

		bundle.NewTarBall()
		bundle.UploadLabelFiles(lbl, sc)

	}

	bundle.NewTarBall()
	defer walg.TimeTrack(time.Now(), "MAIN")
	fmt.Println("Walking ...")
	err = filepath.Walk(in, bundle.TarWalker)
	if err != nil {
		panic(err)
	}
	bundle.Tb.CloseTar()
	bundle.Tb.Finish()

}
