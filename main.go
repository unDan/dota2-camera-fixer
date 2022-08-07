package main

import (
	"bufio"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type dotaCameraAttribute struct {
	AttributeName string
	OldValue      int
	NewValue      int
}

var dllDirPath string = filepath.Join("steamapps", "common", "dota 2 beta", "game", "dota", "bin", "win64")
var dllFileName string = "client.dll"

var backupDirName string = "client_dll_backup"

var configFileName string = "config.json"
var steampathFileName string = "steampath.txt"

var delimChar byte = '\x00'

func readConfigFile() ([]dotaCameraAttribute, error) {
	fbytes, err := ioutil.ReadFile(configFileName)
	if err != nil {
		return nil, err
	}

	var attrs []dotaCameraAttribute

	return attrs, json.Unmarshal(fbytes, &attrs)
}

func readSteampathFile() error {
	fbytes, err := ioutil.ReadFile(steampathFileName)

	if err != nil {
		return err
	}

	steamDir := string(fbytes)

	dllDirPath = filepath.Join(steamDir, dllDirPath)

	return nil
}

func replaceAttributeValues(attrs []dotaCameraAttribute) (map[int]string, error) {
	replacedStrs := make(map[int]string)

	log.Println("[INFO] Opening dota 2 client.dll file...")

	dllFile, err := os.Open(filepath.Join(dllDirPath, dllFileName))
	if err != nil {
		return nil, err
	}
	defer dllFile.Close()

	log.Println("[INFO] Successfully opened client.dll file.")
	log.Println("[INFO] Scanning client.dll file...")

	scanner := bufio.NewReader(dllFile)
	str, err := scanner.ReadString(delimChar)
	currentStrNumber := 1

	for err == nil {
		for _, attr := range attrs {
			if !strings.Contains(str, attr.AttributeName) {
				continue
			}

			log.Printf("[INFO] Found attribute %s.\n", attr.AttributeName)

			changedStr := ""

			for err == nil {
				str, err = scanner.ReadString(delimChar)
				currentStrNumber++

				if strings.Contains(str, strconv.Itoa(attr.OldValue)) {
					changedStr = strings.Replace(
						str,
						strconv.Itoa(attr.OldValue),
						strconv.Itoa(attr.NewValue),
						1,
					)
					break
				}

			}

			if err != nil && err != io.EOF {
				return nil, err
			}

			replacedStrs[currentStrNumber] = changedStr

			log.Printf("[INFO] Successfully changed attribute '%s' value: %d -> %d.\n", attr.AttributeName, attr.OldValue, attr.NewValue)
		}

		str, err = scanner.ReadString(delimChar)
		currentStrNumber++
	}

	if err != nil && err != io.EOF {
		return nil, err
	}

	log.Println("[INFO] Successfully scanned client.dll file and replaced camera attribute values.")

	return replacedStrs, nil
}

func backupDllFile() (isBackupNeeded bool, err error) {
	isBackupNeeded = false
	backupDirPath := filepath.Join(dllDirPath, backupDirName)

	_, err = os.Stat(backupDirPath)

	if os.IsNotExist(err) {
		isBackupNeeded = true

		os.Mkdir(backupDirPath, os.ModeDir)

		in, err := os.Open(filepath.Join(dllDirPath, dllFileName))
		if err != nil {
			return isBackupNeeded, err
		}
		defer in.Close()

		out, err := os.Create(filepath.Join(backupDirPath, dllFileName))
		if err != nil {
			return isBackupNeeded, err
		}
		defer out.Close()

		_, err = io.Copy(out, in)
		if err != nil {
			return isBackupNeeded, err
		}

		return isBackupNeeded, out.Close()
	}

	return isBackupNeeded, nil
}

func rewriteDllFile(replacedStrs map[int]string) error {
	log.Println("[INFO] Opening backup file...")

	backupDllFile, err := os.Open(filepath.Join(dllDirPath, backupDirName, dllFileName))
	if err != nil {
		return err
	}
	defer backupDllFile.Close()

	log.Println("[INFO] Successfully opened backup file.")
	log.Println("[INFO] Opening actual dll file for overwriting...")

	overwrittenDllFile, err := os.Create(filepath.Join(dllDirPath, dllFileName))
	if err != nil {
		return err
	}
	defer overwrittenDllFile.Close()

	log.Println("[INFO] Successfully opened actual dll file.")
	log.Println("[INFO] Rewriting data using replaced attribute values...")

	scanner := bufio.NewReader(backupDllFile)
	str, err := scanner.ReadString(delimChar)
	strNumber := 1

	for err == nil {
		useLineFromOriginalDll := true

		for snum, sstr := range replacedStrs {
			if strNumber == snum {
				_, err := overwrittenDllFile.WriteString(sstr)
				if err != nil {
					return err
				}

				useLineFromOriginalDll = false
				break
			}
		}

		if useLineFromOriginalDll {
			_, err := overwrittenDllFile.WriteString(str)
			if err != nil {
				return err
			}
		}

		str, err = scanner.ReadString(delimChar)
		strNumber++
	}

	if err != nil && err != io.EOF {
		return err
	}

	return nil
}

func main() {
	log.Printf("[INFO] Reading %s file...\n", configFileName)

	attrs, err := readConfigFile()
	if err != nil {
		log.Fatalf("[ERROR] Could not read config file: %v", err)
	}

	log.Printf("[INFO] Successfully read %s file.\n", configFileName)
	log.Printf("[INFO] Reading %s file...\n", steampathFileName)

	err = readSteampathFile()
	if err != nil {
		log.Fatalf("[ERROR] Could not read steampath file: %v", err)
	}

	log.Printf("[INFO] Successfully read %s file.\n", steampathFileName)
	log.Printf("[INFO] Backing up %s file...\n", dllFileName)

	isBackupNeeded, err := backupDllFile()
	if err != nil {
		log.Fatalf("[ERROR] Could not backup dll file: %v", err)
	}

	if isBackupNeeded {
		log.Printf("[INFO] Successfully backed up %s file.\n", dllFileName)
	} else {
		log.Println("[INFO] Backup file already exists, no backup needed.")
	}
	log.Println("[INFO] Changing camera attributes...")

	replacedLines, err := replaceAttributeValues(attrs)
	if err != nil {
		log.Fatalf("[ERROR] Could not replace camera attributes: %v", err)
	}

	log.Println("[INFO] Successfully changed camera attributes.")
	log.Println("[INFO] Overwriting dll file...")

	err = rewriteDllFile(replacedLines)
	if err != nil {
		log.Fatalf("[ERROR] Could not overwrite dll file: %v", err)
	}

	log.Println("[INFO] Done.")
}
