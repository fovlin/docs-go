package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"acovia.net/docs-go/internal/record"
	"github.com/fovlin/opt"
)

const programName = "docs-go"

var (
	config      Config
	sidebar     *SidebarConfig
	root        string
	srcDir      string
	outDir      string
	publicDir   string
	sidebarPath string
	configPath  string

	layout       []layoutPart
	favicon      []byte
	assetVersion string
	bootstrap    string

	configFile string
	srcFlag    string
	outFlag    string
	addrFlag   string
	portFlag   string
	forceFlag  bool

	cmd        string
	subCmdArgs []string
	buildStart time.Time
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
	buildStart = time.Now()

	if err := loadProject(projectArg()); err != nil {
		return err
	}

	return runBuild()
}

func commandDev() error {
	port, err := strconv.Atoi(portFlag)
	if err != nil || port <= 0 || port > 65535 {
		return fmt.Errorf("invalid port: %v", portFlag)
	}

	if err := loadProject(projectArg()); err != nil {
		return err
	}

	return runDev(addrFlag, port)
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

func loadProject(projectRoot string) error {
	absoluteRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return fmt.Errorf("project root: %w", err)
	}

	loaded, path, err := loadConfig(absoluteRoot, configFile)
	if err != nil {
		return err
	}
	applyConfigOverrides(&loaded, srcFlag, outFlag)

	root = absoluteRoot
	config = loaded
	configPath = path
	srcDir = resolveConfigPath(root, config.SrcDir)
	outDir = resolveConfigPath(root, config.OutDir)
	publicDir = resolveConfigPath(root, config.PublicDir)
	sidebarPath = resolveConfigPath(root, config.SidebarFile)

	info, err := os.Stat(srcDir)
	if err != nil {
		return fmt.Errorf("source directory %s: %w", srcDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("source directory is not a directory: %s", srcDir)
	}

	sidebar, err = loadSidebar(sidebarPath, config.Base)
	if err != nil {
		return err
	}

	layoutData, err := themeFS.ReadFile("theme/layout.html")
	if err != nil {
		return fmt.Errorf("read theme layout: %w", err)
	}
	layout = compileLayout(string(layoutData))

	favicon, err = themeFS.ReadFile("theme/favicon.ico")
	if err != nil {
		return fmt.Errorf("read theme favicon: %w", err)
	}
	assetVersion = themeVersion()

	payload, err := json.Marshal(struct {
		Base string   `json:"base"`
		UI   UIConfig `json:"ui"`
	}{
		Base: config.Base,
		UI:   config.UI,
	})
	if err != nil {
		return fmt.Errorf("encode bootstrap: %w", err)
	}
	bootstrap = string(payload)

	titlesMu.Lock()
	titles = make(map[string]titleEntry)
	titlesMu.Unlock()

	return buildURLTab()
}
