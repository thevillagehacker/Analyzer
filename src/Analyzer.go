package main

import (
    "bufio"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "sync"
    "time"
)

const (
    ColorRed    = "\033[31m"
    ColorGreen  = "\033[32m"
    ColorReset  = "\033[0m"
    ColorYellow = "\033[33m"
)

var extCountsMutex sync.Mutex

func main() {
    banner := `
        THE VILLAGEHACKER SECURITY
    https://twitter.com/thevillagehackr                                              
   `

    fmt.Println(banner)

    if len(os.Args) != 2 {
        fmt.Println("Usage: go run main.go <folder_path>")
        os.Exit(1)
    }

    root := os.Args[1]

    startTime := time.Now()

    numDirs, numFiles, numLines, extCounts, errCount := countItems(root)

    elapsedTime := time.Since(startTime)

    fmt.Printf("[+] Number of directories: %s%d%s\n", ColorGreen, numDirs, ColorReset)
    fmt.Printf("[+] Number of files: %s%d%s\n", ColorGreen, numFiles, ColorReset)
    fmt.Printf("[+] Number of lines of code: %s%d%s\n", ColorGreen, numLines, ColorReset)
    fmt.Printf("[+] Time taken: %s%s%s\n", ColorGreen, elapsedTime, ColorReset)

    for _, err := range errCount {
        fmt.Printf("%s\n", err)
    }
    fmt.Println("--------------------------")
    fmt.Printf("%s[*] File extension counts:%s\n", ColorYellow, ColorReset)
    fmt.Println("--------------------------")
    for ext, count := range extCounts {
        fmt.Printf("%s: %s%d%s\n", ext, ColorGreen, count, ColorReset)
    }
}

func countItems(root string) (int, int, int, map[string]int, []error) {
    numDirs := 0
    numFiles := 0
    numLines := 0
    extCounts := make(map[string]int)
    errCount := []error{}
    var wg sync.WaitGroup

    filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            errCount = append(errCount, err)
            fmt.Printf("%s[-] Error accessing path %q: %v%s\n", ColorRed, path, err, ColorReset)
            return nil
        }

        if info.IsDir() {
            numDirs++
        } else {
            numFiles++
            wg.Add(1)

            go func() {
                defer wg.Done()

                file, err := os.Open(path)
                if err != nil {
                    errCount = append(errCount, err)
                    fmt.Printf("%s[-] Error opening file %q: %v%s\n", ColorRed, path, err, ColorReset)
                    return
                }
                defer file.Close()

                reader := bufio.NewReader(file)
                lines := 0
                for {
                    _, err := reader.ReadString('\n')
                    if err == io.EOF {
                        break
                    } else if err != nil {
                        errCount = append(errCount, err)
                        fmt.Printf("%s[-] Error reading file %q: %v%s\n", ColorRed, path, err, ColorReset)
                        return
                    }
                    lines++
                }

                ext := filepath.Ext(path)
                if ext != "" {
                    extCountsMutex.Lock()
                    extCounts[ext]++
                    extCountsMutex.Unlock()
                }

                numLines += lines
            }()
        }

        return nil
    })

    wg.Wait()

    return numDirs, numFiles, numLines, extCounts, errCount
}
