package cli

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/tombenke/go-12f-common/log"
)

type FlagSetFunc func(fs *flag.FlagSet)

func InitConfigs(args []string, flagsetFuncs []FlagSetFunc) {
	fs := flag.NewFlagSet("fs-name", flag.ContinueOnError)
	for _, flagsetFunc := range flagsetFuncs {
		flagsetFunc(fs)
	}
	parseFlagSet(fs, args)
}

func parseFlagSet(fs *flag.FlagSet, args []string) {
	appName := ""
	if len(args) > 0 {
		appName = args[0]
	}
	// Add usage printer function
	fs.Usage = usage(fs, appName)

	err := fs.Parse(args[1:])
	if err != nil {
		log.Logger.Warningf(err.Error())
	}

	// Handle the -h flag
	helpFlag := fs.Lookup("help")
	if helpFlag != nil {
		showHelp, _ := strconv.ParseBool(helpFlag.Value.String())
		if showHelp {
			showUsageAndExit(fs, appName, 0)
		}
	}
}

// Show usage info then exit
func showUsageAndExit(fs *flag.FlagSet, appName string, exitcode int) {
	usage(fs, appName)()
	os.Exit(exitcode)
}

// Print usage information
func usage(fs *flag.FlagSet, appName string) func() {
	return func() {
		fmt.Println("Usage: " + appName + " -h\n")
		fs.PrintDefaults()
	}
}
