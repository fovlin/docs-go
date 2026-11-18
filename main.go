package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
		if len(subCmdArgs) > 0 {
			err = fmt.Errorf("unknown command: %s", subCmdArgs[0])
		} else {
			err = fmt.Errorf("unknown command: %s", cmd)
		}
	}

	if err != nil {
		record.Error(err)
	}
}

func initProgram() error {
	opt.StringVar(&configFile, "c", "config", defaultConfigFile)
	opt.StringVar(&srcFlag, "s", "src", "")
	opt.StringVar(&outFlag, "o", "out", "")
	opt.StringVar(&addrFlag, "a", "addr", "localhost")
	opt.StringVar(&portFlag, "p", "port", "7280")
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

	if err := validateFlagValues(os.Args[1:]); err != nil {
		return err
	}
	if err := opt.Parse(); err != nil {
		return err
	}

	cmd, subCmdArgs = resolveCommand(opt.Args())

	return nil
}

func validateFlagValues(args []string) error {
	valueFlags := map[string]bool{
		"-c": true, "--config": true,
		"-s": true, "--src": true,
		"-o": true, "--out": true,
		"-a": true, "--addr": true,
		"-p": true, "--port": true,
	}

	for index, arg := range args {
		if valueFlags[arg] && index+1 == len(args) {
			return fmt.Errorf("missing value for %s", arg)
		}
	}

	return nil
}

func resolveCommand(args []string) (string, []string) {
	if len(args) == 0 {
		return "build", nil
	}

	switch args[0] {
	case "build", "dev", "serve", "init":
		return args[0], args[1:]
	}

	if strings.HasPrefix(args[0], ".") || filepath.IsAbs(args[0]) {
		return "build", args
	}
	if info, err := os.Stat(args[0]); err == nil && info.IsDir() {
		return "build", args
	}

	return "unknown", args
}

func commandBuild() error {
	dir, err := filepath.Abs(projectArg())
	if err != nil {
		return err
	}
	if err := os.Chdir(dir); err != nil {
		return err
	}

	app, err := LoadApp(dir, configFile, srcFlag, outFlag)
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

	dir, err := filepath.Abs(projectArg())
	if err != nil {
		return err
	}
	if err := os.Chdir(dir); err != nil {
		return err
	}

	app, err := LoadApp(dir, configFile, srcFlag, outFlag)
	if err != nil {
		return err
	}

	return runDev(app, addrFlag, port)
}

func commandInit() error {
	return runInit(projectArg(), forceFlag, configFile)
}

func projectArg() string {
	if len(subCmdArgs) > 0 {
		return subCmdArgs[0]
	}

	return "."
}
