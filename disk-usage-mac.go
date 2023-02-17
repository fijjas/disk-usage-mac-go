package main

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type FsItem struct {
	isDir      bool
	name       string
	fullPath   string
	parentPath string
	sizeBytes  int64
	n          int
}

type FsItemSize struct {
	path string
	size int64
}

func main() {
	fmt.Println("Initializing...")

	startPath, err := getStartDir()
	if err != nil {
		fmt.Println("Cannot get start path", err)
		os.Exit(1)
	}

	inputChan := make(chan string)

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			inputChan <- scanner.Text()
		}
	}()

	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, syscall.SIGINT, syscall.SIGTERM)

	mainLoop(*startPath, inputChan, exitChan)
}

func mainLoop(startPath string, inputChan chan string, exitChan chan os.Signal) {
	var rootPath = "/"
	var currentPath = startPath
	var command = "."

	byeBye := func() {
		fmt.Println()
		fmt.Println("Bye! See ya!")
		time.Sleep(100 * time.Millisecond)
	}

mainLoop:
	for {
		dirItems, err := listDirContents(currentPath)
		if err != nil {
			logError("Fatal error: Cannot read dir contents!", err,
				"Switching to a fallback path in 5s...", 5)
			currentPath = rootPath
			continue mainLoop
		}

		draw(currentPath, &dirItems)

		select {
		case <-exitChan:
			byeBye()
			return
		case command = <-inputChan:
			switch command {
			case "q", "\\q", "quit", "exit", "bye", "bye!":
				byeBye()
				return
			case "open", "finder":
				openInFinder(currentPath, false)
				continue mainLoop
			case "reveal":
				openInFinder(currentPath, true)
				continue mainLoop
			case ".":
				continue mainLoop
			case "..":
				if currentPath == rootPath {
					continue mainLoop
				}
				currentPath = filepath.Dir(currentPath)
				break
			default:
				selectedNum, err := strconv.Atoi(command)
				if err != nil {
					continue mainLoop
				}
				for _, item := range dirItems {
					if item.n == selectedNum {
						currentPath = path.Join(currentPath, item.name)
						continue mainLoop
					}
				}
			}
		}
	}
}

func listDirContents(path string) ([]FsItem, error) {
	dirItems, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	results := make([]FsItem, 0, len(dirItems)-1)

	for _, dirItem := range dirItems {
		fullPath := filepath.Join(path, dirItem.Name())

		fi := FsItem{
			parentPath: path,
			fullPath:   fullPath,
			name:       dirItem.Name(),
			sizeBytes:  int64(0),
			isDir:      dirItem.IsDir(),
			n:          0,
		}
		if !dirItem.IsDir() {
			fileInfo, err := dirItem.Info()
			if err != nil {
				fi.sizeBytes = -1
			} else {
				fi.sizeBytes = fileInfo.Size()
			}
		}

		results = append(results, fi)
	}

	var getSizesWg sync.WaitGroup
	sizesChan := make(chan FsItemSize)

	for _, fsItem := range results {
		if fsItem.isDir {
			getSizesWg.Add(1)
			go calcDirSize(fsItem.fullPath, &getSizesWg, sizesChan)
		}
	}

	go func() {
		getSizesWg.Wait()
		close(sizesChan)
	}()

	for sizeRes := range sizesChan {
		for j := range results {
			if results[j].fullPath == sizeRes.path && results[j].isDir {
				results[j].sizeBytes = sizeRes.size
			}
		}
	}

	sorter := func(i, j int) bool {
		return results[i].sizeBytes > results[j].sizeBytes
	}

	sort.Slice(results, sorter)

	return results, nil
}

func calcDirSize(path string, wg *sync.WaitGroup, resChan chan<- FsItemSize) {
	defer wg.Done()
	var sum = int64(0)
	err := filepath.Walk(path, func(subPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			sum += info.Size()
		}
		return nil
	})

	if err != nil {
		logError("Walk error!", err,
			fmt.Sprintf("Path: %s", path), 3)
	}

	resChan <- FsItemSize{path: path, size: sum}
}

func getStartDir() (*string, error) {
	var defaultPath = "/"

	if len(os.Args) == 1 {
		return &defaultPath, nil
	}

	if len(os.Args) > 2 {
		return nil, errors.New("invalid number of arguments, expected 0-1")
	}

	var pathFromArg = os.Args[1]
	st, err := os.Stat(pathFromArg)
	if err != nil {
		return nil, err
	} else if !st.IsDir() {
		return nil, errors.New("the path is not a directory")
	}
	return &pathFromArg, nil
}

func clr() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		fmt.Println("Cannot clear screen", err)
	}
}

func draw(currentPath string, items *[]FsItem) {
	const ignoreZeroSizeItems = true
	clr()
	fmt.Println()
	fmt.Println("============================== Mac Disk Usage Utility ==============================")
	fmt.Println("Enter folder # to explore, \"..\" to go back, \".\" to refresh, Ctrl+C to quit ")
	fmt.Println("------------------------------------------------------------------------------------")
	fmt.Println(currentPath)
	fmt.Println("------------------------------------------------------------------------------------")

	var numSec, nameSec, sizeSec, delimiter string
	var num = 1

	for j := range *items {
		if ignoreZeroSizeItems && (*items)[j].sizeBytes < 1 {
			continue
		}
		if (*items)[j].isDir {
			(*items)[j].n = num
			num++
			numSec = formatItemNumber5((*items)[j].n)
			nameSec = formatFsItemName64(fmt.Sprintf("[%s]", (*items)[j].name))
			delimiter = "+"
		} else {
			numSec = strings.Repeat(" ", 5)
			nameSec = formatFsItemName64((*items)[j].name)
			delimiter = "-"
		}
		sizeSec = formatBytes10((*items)[j].sizeBytes)
		fmt.Printf("%s %s %s%s\n", numSec, delimiter, nameSec, sizeSec)
	}
	fmt.Println("------------------------------------------------------------------------------------")
	fmt.Print("> ")
}

func formatBytes10(bytes int64) string {
	base := float64(1024)
	if bytes < int64(base) {
		return fmt.Sprintf("%10d B", bytes)
	}
	exp := int64(math.Log(float64(bytes)) / math.Log(base))
	units := []string{"KB", "MB", "GB", "TB", "PB"}
	rounded := math.Round(float64(bytes)/math.Pow(base, float64(exp))*100) / 100
	return fmt.Sprintf("%10.2f %s", rounded, units[exp-1])
}

func formatFsItemName64(name string) string {
	const width = 64
	if len(name) > width {
		name = name[:(width-3)] + "..."
	}
	if len(name) == width {
		return name
	}
	padding := width - len(name)
	return fmt.Sprintf("%s%s", name, strings.Repeat(" ", padding))
}

func formatItemNumber5(num int) string {
	const width = 5
	var ns = fmt.Sprintf("%d", num)
	padding := width - len(ns)
	if padding <= 0 {
		return ns
	}
	return fmt.Sprintf("%s%s", strings.Repeat(" ", padding), ns)
}

func logError(msg string, err error, preTimerMsg string, timeoutSec time.Duration) {
	fmt.Println(msg, err)
	fmt.Println(preTimerMsg)
	time.Sleep(timeoutSec * time.Second)
}

func openInFinder(dir string, reveal bool) {
	var cmd *exec.Cmd
	if reveal {
		cmd = exec.Command("open", "-R", dir)
	} else {
		cmd = exec.Command("open", dir)
	}
	err := cmd.Run()
	if err != nil {
		logError("Cannot open the directory in Finder", err, "3s...", 3)
	}
}
