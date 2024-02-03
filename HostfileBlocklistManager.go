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
	"slices"
	"strings"
)

const src = "C:\\Windows\\System32\\Drivers\\etc\\hosts"
const backupName = "./backups/hosts"
const whitelistsFile = "./whitelist.txt"
const blocklistsFile = "./blocklists.txt"
const blacklistsFile = "./blacklists.txt"

func backupHostFile() {
	_ = os.Mkdir("backups", os.ModePerm)

	fin, err := os.Open(src)
	if err != nil {
		fmt.Printf("Unable to backup your hosts file due to: %s\n This might be due to you not running this "+
			"in an elevated shell.", err)
		log.Fatal(err)
	}
	defer fin.Close()

	fout, err := os.Create(backupName)
	if err != nil {
		fmt.Printf("Unable to backup your hosts file due to: %s\n This might be due to you not running this "+
			"in an elevated shell.", err)
		log.Fatal(err)
	}
	defer fout.Close()

	_, err = io.Copy(fout, fin)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Backed up your host file!")
}

// https://stackoverflow.com/a/33853856
func updateHostFile(url string, hostfile string, whitelist []string) (updatedHostfile string) {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		//return err
		fmt.Printf("Unable to fetch %s due to %s\nSkipping.", url, err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			// Doesn't really matter
		}
	}(resp.Body)

	// Check server response
	if resp.StatusCode != http.StatusOK {
		//return fmt.Errorf("bad status: %s", resp.Status)
		fmt.Printf("%s gave bad status: %s\nSkippping.", url, resp.Status)
	}

	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 1 {
			continue
		}
		if line[0:1] != "#" {
			domain := strings.Split(line, " ")[1]
			if slices.Contains(whitelist, domain) {
				continue
			}

			if strings.HasPrefix(line, "127.0.0.1") || strings.HasPrefix(line, "0.0.0.0") {
				hostfile += fmt.Sprintf("%s\n", line)
			} else {
				hostfile += fmt.Sprintf("127.0.0.1 %s\n", line)
			}
		}
	}
	return hostfile
}

func addBlocklist(url string) {
	// check if blocklists file already exists
	f, err := os.Open(blocklistsFile)
	if err != nil {
		log.Fatal(err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(f)

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

	writeFile, err := os.OpenFile(blocklistsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer func(writeFile *os.File) {
		err := writeFile.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(writeFile)

	_, err = writeFile.WriteString(fmt.Sprintf("%s\n", url))
	if err != nil {
		fmt.Printf("There was an issue adding %s to blocklists\n", url)
	}
}

func updateAllBlocklists() {
	f, err := os.Open(blocklistsFile)
	if err != nil {
		log.Fatal(err)
	}
	// remember to close the file at the end of the program
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(f)

	// read and load whitelists
	var whitelistedDomains []string
	whitelist, _ := os.Open(whitelistsFile)
	whitelistScanner := bufio.NewScanner(whitelist)
	for whitelistScanner.Scan() {
		whitelistedDomains = append(whitelistedDomains, whitelistScanner.Text())
	}

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
			backupHostStr = updateHostFile(scanner.Text(), backupHostStr, whitelistedDomains)
		}
	}

	// Add blacklisted domains
	blacklist, _ := os.Open(blacklistsFile)
	blacklistScanner := bufio.NewScanner(blacklist)
	for blacklistScanner.Scan() {
		backupHostStr += fmt.Sprintf("127.0.0.1 %s\n", blacklistScanner.Text())
	}

	f, _ = os.OpenFile(src, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(f)
	err = f.Truncate(0)
	_, err = f.Seek(0, 0)
	_, _ = f.WriteString(backupHostStr)
}

func whitelistDomain(domain string) {
	readOnly, _ := os.Open(whitelistsFile)
	scanner := bufio.NewScanner(readOnly)
	for scanner.Scan() {
		if scanner.Text() == domain {
			fmt.Printf("%s has already been whitelisted!\n", domain)
			return
		}
	}

	f, err := os.OpenFile(whitelistsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(f)

	_, err = f.WriteString(fmt.Sprintf("%s\n", domain))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(fmt.Sprintf("%s has been whitelisted!", domain))
}

func blacklistDomain(domain string) {
	readOnly, _ := os.Open(blacklistsFile)
	scanner := bufio.NewScanner(readOnly)
	for scanner.Scan() {
		if scanner.Text() == domain {
			fmt.Printf("%s has already been blacklisted!\n", domain)
			return
		}
	}
	f, err := os.OpenFile(blacklistsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(f)

	_, err = f.WriteString(fmt.Sprintf("%s\n", domain))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(fmt.Sprintf("%s has been blacklisted!", domain))
}

func main() {
	blocklistURL := flag.String("blocklist", "", "URL of blocklist to add")
	updateAllFlag := flag.Bool("update", false, "Update all blocklists")
	whitelistDomainFlag := flag.String("whitelist", "", "Domain to whitelist")
	blacklistDomainFlag := flag.String("blacklist", "", "Domain to blacklist")
	flag.Parse()

	// Create required files if they don't exist on first run
	if _, err := os.Stat(backupName); errors.Is(err, os.ErrNotExist) {
		backupHostFile()
	}
	if _, err := os.Stat(whitelistsFile); errors.Is(err, os.ErrNotExist) {
		wl, _ := os.Create(whitelistsFile)
		err := wl.Close()
		if err != nil {
			return
		}
	}
	if _, err := os.Stat(blocklistsFile); errors.Is(err, os.ErrNotExist) {
		bl, _ := os.Create(blocklistsFile)
		err := bl.Close()
		if err != nil {
			return
		}
	}
	if _, err := os.Stat(blacklistsFile); errors.Is(err, os.ErrNotExist) {
		bl, _ := os.Create(blacklistsFile)
		err := bl.Close()
		if err != nil {
			return
		}
	}
	if *blocklistURL != "" {
		addBlocklist(*blocklistURL)
		updateAllBlocklists()
	}
	if *whitelistDomainFlag != "" {
		whitelistDomain(*whitelistDomainFlag)
	}
	if *blacklistDomainFlag != "" {
		blacklistDomain(*blacklistDomainFlag)
	}
	if *updateAllFlag {
		updateAllBlocklists()
	}
}
