package main

import (
	"fmt"
	"os"
	"strconv"

	"acovia.net/docs-go/internal/record"
	"github.com/fovlin/opt"
)

const programName = "docs-go"

var (
	configFile string
	srcFlag    string
	outFlag    string
	addrFlag   string
	portFlag   string
	forceFlag  bool

	cmd        string
	subCmdArgs []string
)

func main() {
	if err := initProgram(); err != nil {
		record.Error(err)
	}

	var err error
	switch cmd {
	case "build":
		err = commandBuild()
	case "dev", "serve":
		err = commandDev()
	case "init":
		err = commandInit()
	default:
		err = fmt.Errorf("unknown command: %s", cmd)
	}

	if err != nil {
		record.Error(err)
	}
}

func initProgram() error {
	opt.StringVar(&configFile, "c", "config", defaultConfigFile)
	opt.StringVar(&srcFlag, "s", "src", "")
	opt.StringVar(&outFlag, "o", "out", "")
	opt.StringVar(&addrFlag, "a", "addr", "0.0.0.0")
	opt.StringVar(&portFlag, "p", "port", "8080")
	opt.BoolSwitch(&logQuiet, "q", "quiet")
	opt.BoolSwitch(&record.EnableColor, "", "color")
	opt.BoolSwitch(&record.EnableTimestamp, "t", "time")
	opt.BoolSwitch(&forceFlag, "f", "force")

	opt.BoolFunc("h", "help", func() {
		record.Print(helpText)
		os.Exit(0)
	})
	opt.BoolFunc("v", "version", func() {
		record.Println(programName, version)
		os.Exit(0)
	})

	if err := opt.Parse(); err != nil {
		return err
	}

	cmd, subCmdArgs = resolveCommand(opt.Args())

	return nil
}

func resolveCommand(args []string) (string, []string) {
	if len(args) > 0 {
		switch args[0] {
		case "build", "dev", "serve", "init":
			return args[0], args[1:]
		}
	}

	return "build", args
}

func commandBuild() error {
	app, err := LoadApp(projectArg(), configFile, srcFlag, outFlag)
	if err != nil {
		return err
	}

	return runBuild(app)
}

func commandDev() error {
	port, err := strconv.Atoi(portFlag)
	if err != nil || port <= 0 || port > 65535 {
		return fmt.Errorf("invalid port: %v", portFlag)
	}

	app, err := LoadApp(projectArg(), configFile, srcFlag, outFlag)
	if err != nil {
		return err
	}

	return runDev(app, addrFlag, port)
}

func commandInit() error {
	return runInit(projectArg(), forceFlag)
}

func projectArg() string {
	if len(subCmdArgs) > 0 {
		return subCmdArgs[0]
	}

	return "."
}
