// Copyright 2015 Jeff Martinez. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE.txt file
// or at http://opensource.org/licenses/MIT

/*
See README.md for full description and usage info.
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"

	"github.com/jeffbmartinez/cleanexit"
	"github.com/jeffbmartinez/delay"
	"github.com/jeffbmartinez/devserver/handler"
)

const EXIT_SUCCESS = 0
const EXIT_FAILURE = 1
const EXIT_USAGE_FAILURE = 2 // Same as golang's flag module uses, hardcoded at https://github.com/golang/go/blob/release-branch.go1.4/src/flag/flag.go#L812

const PROJECT_NAME = "devserver"

func main() {
	cleanexit.SetUpExitOnCtrlC(showNiceExitMessage)

	randomSeed := time.Now().UnixNano()
	rand.Seed(randomSeed)

	allowAnyHostToConnect, listenPort, directoryToServe, noDirectory := getCommandLineArgs()

	verifyDirectoryOrDie(directoryToServe)

	router := mux.NewRouter()

	if !noDirectory {
		router.Handle("/dir/{pathname:.*}", handler.NewFileServer("/dir/", directoryToServe))
	}
	router.HandleFunc("/echo/{echoString:.*}", handler.Echo)
	router.HandleFunc("/random", handler.Random)
	router.Handle("/counter", handler.NewCounter())

	n := negroni.New()
	n.Use(delay.Middleware{})
	n.UseHandler(router)

	listenHost := "localhost"
	if allowAnyHostToConnect {
		listenHost = ""
	}

	displayServerInfo(directoryToServe, listenHost, listenPort, noDirectory)

	listenAddress := fmt.Sprintf("%v:%v", listenHost, listenPort)
	n.Run(listenAddress)
}

func verifyDirectoryOrDie(dir string) {
	fileInfo, err := os.Stat(dir)
	if err != nil {
		log.Fatal(fmt.Sprintf("Cannot read directory '%v'", dir))
		os.Exit(EXIT_FAILURE)
	}

	if !fileInfo.IsDir() {
		log.Fatal(fmt.Sprintf("This is not a directory: '%v'", dir))
		os.Exit(EXIT_FAILURE)
	}
}

func getCanonicalDirName(dir string) string {
	canonicalDirName, err := filepath.Abs(dir)

	if err != nil {
		/* After following golang's source code, this should only happen
		in fairly odd conditions such as being unable to resolve the working
		directory. */
		log.Fatal(fmt.Sprintf("Cannot serve from directory '%v'", dir))
		os.Exit(EXIT_FAILURE)
	}

	return canonicalDirName
}

func showNiceExitMessage() {
	/* \b is the equivalent of hitting the back arrow. With the two following
	   space characters they serve to hide the "^C" that is printed when
	   ctrl-c is typed.
	*/
	fmt.Printf("\b\b  \n[ctrl-c] %v is shutting down\n", PROJECT_NAME)
}

func getCommandLineArgs() (allowAnyHostToConnect bool, port int, directoryToServe string, noDirectory bool) {
	const DEFAULT_PORT = 8000
	const DEFAULT_DIR = "."

	flag.BoolVar(&allowAnyHostToConnect, "a", false, "Use to allow any ip address (any host) to connect. Default allows ony localhost.")
	flag.IntVar(&port, "port", DEFAULT_PORT, "Port on which to listen for connections.")
	flag.StringVar(&directoryToServe, "dir", DEFAULT_DIR, "Directory to serve. Default is current directory.")
	flag.BoolVar(&noDirectory, "nodir", false, "Disable the file server.")

	flag.Parse()

	/* Don't accept any positional command line arguments. flag.NArgs()
	counts only non-flag arguments. */
	if flag.NArg() != 0 {
		/* flag.Usage() isn't in the golang.org documentation,
		but it's right there in the code. It's the same one used when an
		error occurs parsing the flags so it makes sense to use it here as
		well. Hopefully the lack of documentation doesn't mean it's gonna be
		changed it soon. Worst case can always copy that code into a local
		function if it goes away :p
		Currently using go 1.4.1
		https://github.com/golang/go/blob/release-branch.go1.4/src/flag/flag.go#L411
		*/
		flag.Usage()
		os.Exit(EXIT_USAGE_FAILURE)
	}

	return
}

func displayServerInfo(directoryToServe string, listenHost string, listenPort int, disableFileServer bool) {
	visibleTo := listenHost
	if visibleTo == "" {
		visibleTo = "All ip addresses"
	}

	directoryNameText := "[File Server is disabled]"
	if !disableFileServer {
		directoryNameText = getCanonicalDirName(directoryToServe)
	}

	fmt.Printf("%v is running.\n\n", PROJECT_NAME)
	fmt.Printf("Directory: '%v'\n           ^---> Available at http://[domain]:%v/dir/*\n", directoryNameText, listenPort)
	fmt.Printf("Visible to: %v\n", visibleTo)
	fmt.Printf("Port: %v\n\n", listenPort)
	fmt.Printf("Hit [ctrl-c] to quit\n")
}
