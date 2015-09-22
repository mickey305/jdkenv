package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

const VERSION = `0.0.2`

var (
	version       = flag.Bool("version", false, "display version information")
	v             = flag.Bool("v", false, "display version information")
	help          = flag.Bool("help", false, "display help information")
	h             = flag.Bool("h", false, "display help information")
	jdkdir        = homeDir() + "/.jdkenv/java"
	macSystemJdk  = "/System/Library/Java/JavaVirtualMachines/"
	macLibraryJdk = "/Library/Java/JavaVirtualMachines/"
)

func main() {
	//var jdkdir = homeDir() + "/.jdkenv/java"
	flag.Parse()

	if *v || *version {
		fmt.Printf("jdkenv %s\n", VERSION)
		os.Exit(0)
	}

	if *h || *help {
		fmt.Println("print help")
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("print help")
	} else {
		switch args[0] {
		case "init":
			initialize()
		case "list", "versions":
			printList()
		case "use", "set":
			if len(args) >= 2 {
				use(args[1])
				return
			} else {
				fmt.Println("please select jdk directory")
			}
		case "current", "version":
			fmt.Println(getCurrent())
		default:
			fmt.Printf("not found: %s\n", args[0])
		}
	}

	os.Exit(0)
}

func use(ver string) {
	ver = getSearchedJdkName(ver)

	if runtime.GOOS == "darwin" {
		macUse(ver)
		return
	}
	if !exist(jdkdir + "/" + ver) {
		fmt.Println(ver + "is not exist")
		return
	}

	jdkpath := jdkdir + "/" + ver
	javahomesymlink := jdkdir + "/current"

	removeCurrnetSymlink(javahomesymlink)
	makeJavahomeSymlink(jdkpath, javahomesymlink)
}

func getSearchedJdkName(ver string) string {
	jdkList := getList()
	if haveAJdk(jdkList, ver) {
		for _, value := range jdkList {
			if strings.Contains(value, ver) { ver = value }
		}
	}

	return ver
}

func haveAJdk(jdkList []string, ver string) bool {
	count := 0
	for _, value := range jdkList {
		if strings.Contains(value, ver) { count++ }
	}

	return count == 1
}

func macUse(ver string) {
	var jdkpath string
	if exist(macSystemJdk + ver) {
		jdkpath = macSystemJdk + ver + "/Contents/Home"
	} else if exist(macLibraryJdk + ver) {
		jdkpath = macLibraryJdk + ver + "/Contents/Home"
	} else {
		fmt.Println(ver + " isn't exists at this System")
		return
	}

	javahomesymlink := jdkdir + "/current"

	removeCurrnetSymlink(javahomesymlink)
	makeJavahomeSymlink(jdkpath, javahomesymlink)
}

func removeCurrnetSymlink(javahomesymlink string) {
	if exist(javahomesymlink) {
		if err := os.Remove(javahomesymlink); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func makeJavahomeSymlink(jdkpath, javahomesymlink string) {
	if err := os.Symlink(jdkpath, javahomesymlink); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func printList()  {
	for _, value := range getList() {
		if getCurrent() == value {
			fmt.Println("* " + value)
		} else {
			fmt.Println("  " + value)
		}
	}
}

func getList() []string {
	if runtime.GOOS == "darwin" {
		return getMacJdkList()
	}
	dirs, err := ioutil.ReadDir(jdkdir)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	if len(dirs) == 0 {
		fmt.Println("jdk isn't exists at " + jdkdir)
		return nil
	}

	slice := make([]string, 0)
	for _, value := range dirs {
		if strings.HasPrefix(value.Name(), "jdk") {
			slice = append(slice, value.Name())
		}
	}

	return slice
}

func getMacJdkList() []string {
	slice := make([]string, 0)
	slice = append(slice, getJdkList(macSystemJdk)...)
	slice = append(slice, getJdkList(macLibraryJdk)...)
	return slice
}

func getJdkList(dirPath string) []string {
	dirs, err := ioutil.ReadDir(dirPath)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	slice := make([]string, 0)
	for _, value := range dirs { slice = append(slice, value.Name()) }

	return slice
}

func getCurrent() string {
	javahomesymlink := jdkdir + "/current"
	if !exist(javahomesymlink) { return "jdkenv not used" }

	dest, err := os.Readlink(javahomesymlink)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	currentJdkVersion := ""
	switch runtime.GOOS {
	case "darwin":
		{
			splitedpath := strings.Split(dest, string(os.PathSeparator))
			currentJdkVersion = splitedpath[len(splitedpath)-3]
		}
	default:
		currentJdkVersion = filepath.Base(dest)
	}

	return currentJdkVersion
}

func initialize() {
	if !exist(jdkdir) {
		err := os.MkdirAll(jdkdir, 0777)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	switch runtime.GOOS {
	case "windows":
		windowsInit()
	default:
		unixTypeInit()
	}
}

func windowsInit() {
	_, err := exec.Command("setx", "JAVA_HOME", jdkdir+"/current").Output()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("please reboot command prompt to recognize JAVA_HOME")

	if hasGitBash() {
		fmt.Println("if you use git bash, write in your .bashrc below")
		printSetJavaHomeMsg()
	}
}

func unixTypeInit() {
	fmt.Println("write in your .bashrc below")
	printSetJavaHomeMsg()
}

func printSetJavaHomeMsg() {
	fmt.Println("export JAVA_HOME=" + jdkdir + "/current")
	fmt.Println("and execute below")
	fmt.Println(". ~/.bashrc")
}

func homeDir() string {
	usr, err := user.Current()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return usr.HomeDir
}

func exist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func hasGitBash() bool {
	return runtime.GOOS == "windows" && os.Getenv("HOME") == homeDir()
}
