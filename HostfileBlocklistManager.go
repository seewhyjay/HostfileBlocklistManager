package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

const src = "C:\\Windows\\System32\\Drivers\\etc\\hosts"
const backupName = "./backups/hosts"

func backupHostFile() {
	_ = os.Mkdir("backups", os.ModePerm)

	fin, err := os.Open(src)
	if err != nil {
		log.Fatal(err)
	}
	defer fin.Close()

	fout, err := os.Create(backupName)
	if err != nil {
		log.Fatal(err)
	}
	defer fout.Close()

	_, err = io.Copy(fout, fin)

	if err != nil {
		log.Fatal(err)
		fmt.Println("Backed up your host file!")
	}
}

// https://stackoverflow.com/a/33853856
func updateHostFile(url string, hostfile string) (updatedHostfile string) {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		//return err
		fmt.Println(err)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		//return fmt.Errorf("bad status: %s", resp.Status)
		fmt.Printf("bad status: %s", resp.Status)
	}

	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 1 {
			continue
		}
		if line[0:1] != "#" {
			hostfile += fmt.Sprintf("%s\n", line)
		}
	}
	return hostfile
}

func addBlocklist(url string) {
	// check if blocklists file already exists
	f, err := os.Open("blocklists.txt")
	if err != nil {
		log.Fatal(err)
	}
	// remember to close the file at the end of the program
	defer f.Close()

	// read the file line by line using scanner
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		if scanner.Text() == url {
			fmt.Printf("%s has already been added!\n", url)
			return
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	writeFile, err := os.OpenFile("blocklists.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer writeFile.Close()

	_, err = writeFile.WriteString(fmt.Sprintf("%s\n", url))
	if err != nil {
		fmt.Printf("There was an issue adding %s to blocklists\n", url)
	}
	writeFile.Close()
}

func updateAllBlocklists() {
	f, err := os.Open("blocklists.txt")
	if err != nil {
		log.Fatal(err)
	}
	// remember to close the file at the end of the program
	defer f.Close()

	// read backup
	backupHost, _ := os.ReadFile(backupName)
	backupHostStr := string(backupHost)

	// read the file line by line using scanner
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 5 { // http is 4 char long
			continue
		}
		if line[0:4] == "http" {
			fmt.Println(fmt.Sprintf("Updating %s", scanner.Text()))
			backupHostStr = updateHostFile(scanner.Text(), backupHostStr)
		}
	}
	f, _ = os.OpenFile(src, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	err = f.Truncate(0)
	_, err = f.Seek(0, 0)
	_, _ = f.WriteString(backupHostStr)
}

func main() {
	blocklistURL := flag.String("blocklist", "", "URL of blpcklist to add")
	updateAllFlag := flag.Bool("update", false, "Update all blocklists")
	flag.Parse()

	if _, err := os.Stat(backupName); errors.Is(err, os.ErrNotExist) {
		backupHostFile()
	}
	if *blocklistURL != "" {
		addBlocklist(*blocklistURL)
		updateAllBlocklists()
	}
	if *updateAllFlag {
		updateAllBlocklists()
	}
}
