/*
 * signing
 *
 * singing files with a given sign key
 *
 * Copyright (c) 2018, 2019 aintahydra <aintahydra@gmail.com>
 *
 * Licensed under the GPLv2.
 */
 /*
 * NOTE: the "evmctl" utility is needed (e.g., "apt-get install ima-evm-utils")
 * NOTE: running "evmctl" may require the administrative privilege
*/
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

/**
 * doSign() - sign on files using the evmctl utility
 *
 * @kpath: a path to the sign key
 * @fpath: files to get signed
 * Returns: none
*/
func doSign(kpath, fpath string) {
	signcmdstr := fmt.Sprintf("%s %s %s %s %s", "/usr/bin/evmctl", "ima_sign", "--key", kpath, fpath)
	signcmd := exec.Command("sh", "-c", signcmdstr)
	_, err := signcmd.Output()
	if err != nil {
		log.Fatal(err)
	}
}

/**
 * getFiles() - (recursively) collect a list of files under a given file path
 *
 * @dir: a given file path
 * Returns: a list files under the given file path
*/
func getFiles(dir string) []string {

	var flist []string

	dirstr := handlePathPrefix(dir)
	
	err := filepath.Walk(dirstr, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		/* fmt.Printf("File name: %s   Permission: %s\n", f.Name(), f.Mode()) */
		switch fmode := f.Mode(); {
		case fmode.IsRegular():
			flist = append(flist, path)
		default:
		/* mode.IsDir(),
    mode&os.ModeSymlink, mode&os.ModeNamedPipe, mode&os.ModeSocket,
    mode&os.ModeDevice, mode&os.ModeIrregular */
			/* fmt.Println("Ignoring ", f.Name()) */
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return flist
}

/**
 * handlePathPrefix() - get rid of the path shorthands(~ or .)
 *
 * @src: a file path
 * Returns: a full path string after replacing the ~ or . prefix
 */
func handlePathPrefix(src string) string {
	usr, err := user.Current()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	homedir := usr.HomeDir
	if strings.HasPrefix(src, "~/") {
		src = filepath.Join(homedir, src[2:])
	}

	workingdir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	if strings.HasPrefix(src, "./") {
		src = filepath.Join(workingdir, src[2:])
	}

	return src
}

/**
 * parseFlags() - parse the command-line arguments
 *
 * @key: a .pem file
 * @in:  a text file that contains target paths
 * @num: the number of go routines that sign in parallel
 * Returns: key/in/pdgree strings
 */
func parseFlags() (string, string, int) {
	key := flag.String("key", "", "signing key")
	in := flag.String("in", "", "target")
	nog := flag.Int("pdgree", 1, "parallization degree")

	flag.Parse()

	if flag.NFlag() < 2 {
		flag.Usage()
		os.Exit(1)
	}

	return *key, *in, *nog
}

func main() {

	start := time.Now()

	keyPath, inFilePath, nSigners := parseFlags()

	keypathstr := handlePathPrefix(keyPath)
	inputdirpathstr := handlePathPrefix(inFilePath)

	var dirs []string

	var filepaths []string

	/* Step 1: read the input file */
	infile, err := os.Open(inputdirpathstr)
	if err != nil {
		log.Fatal(err)
	}
	defer infile.Close()

	scanner := bufio.NewScanner(infile)
	for scanner.Scan() {
		dirs = append(dirs, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	
	/* Step 2: find files */
	var findwg sync.WaitGroup
	var signwg sync.WaitGroup
	
	for _, dirname := range dirs {
		findwg.Add(1)
		go func(dir string) {
			defer findwg.Done()
			files := getFiles(dir)
			for _, t := range files {
				if len(t) != 0 {
					filepaths = append(filepaths, t)
				}
			}
		}(dirname)
	}
	findwg.Wait()

	/* Step 3: Sign files */	
	fmt.Printf("Signing with %d routine(s)\n", nSigners)
	guard := make(chan struct{}, nSigners)

	for _, fpath := range filepaths {
		signwg.Add(1)
		guard <- struct{}{}
		go func(path string) {
			defer signwg.Done()
			fmt.Println("Sign on: ", path)
			doSign(keypathstr, path)
			<-guard
		}(fpath)
	}
	signwg.Wait()

	fmt.Println("Done! It took ", time.Since(start))
}

