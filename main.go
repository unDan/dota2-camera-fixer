package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"os"
	"path/filepath"
	"strconv"
	"strings"

	cfg "github.com/undan/dota2-camera-fixer/internal/config"
	"github.com/undan/dota2-camera-fixer/internal/logger"
	"gopkg.in/ini.v1"
)

type DotaCameraAttribute struct {
	AttributeName string
	OldValue      int
	NewValue      int
}

var configFileName string = "config.ini"
var userPrefs cfg.UserPrefs = cfg.UserPrefs{}
var appSettings cfg.AppSettings = cfg.AppSettings{
	FileReaderDelimiter: '\x00',
}

var log logger.Logger

func readConfigFile() error {
	iniData, err := ini.Load(configFileName)

	if err != nil {
		return err
	}

	err = iniData.Section("UserPrefs").MapTo(&userPrefs)

	if err != nil {
		return err
	}

	log = logger.NewLogger(userPrefs.ShowLogInfo)

	err = iniData.Section("AppSettings").MapTo(&appSettings)

	return err
}

func readCameraValuesFile() ([]DotaCameraAttribute, error) {
	fbytes, err := ioutil.ReadFile(userPrefs.CameraValuesFileName)
	if err != nil {
		return nil, err
	}

	var attrs []DotaCameraAttribute

	return attrs, json.Unmarshal(fbytes, &attrs)
}

func replaceAttributeValues(attrs []DotaCameraAttribute) (map[int]string, error) {
	replacedStrs := make(map[int]string)

	log.Info("[INFO] Opening dota 2 client.dll file...")

	dllFile, err := os.Open(filepath.Join(userPrefs.SteamDirPath, appSettings.DllDirPath, appSettings.DllFileName))
	if err != nil {
		return nil, err
	}
	defer dllFile.Close()

	log.Info("[INFO] Successfully opened client.dll file.")
	log.Info("[INFO] Scanning client.dll file...")

	scanner := bufio.NewReader(dllFile)
	str, err := scanner.ReadString(appSettings.FileReaderDelimiter)
	currentStrNumber := 1

	for err == nil {
		for _, attr := range attrs {
			if !strings.Contains(str, attr.AttributeName) {
				continue
			}

			log.Info("[INFO] Found attribute %s.", attr.AttributeName)

			changedStr := ""

			for err == nil {
				str, err = scanner.ReadString(appSettings.FileReaderDelimiter)
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

			log.Info("[INFO] Successfully changed attribute '%s' value: %d -> %d.", attr.AttributeName, attr.OldValue, attr.NewValue)
		}

		str, err = scanner.ReadString(appSettings.FileReaderDelimiter)
		currentStrNumber++
	}

	if err != nil && err != io.EOF {
		return nil, err
	}

	log.Info("[INFO] Successfully scanned client.dll file and replaced camera attribute values.")

	return replacedStrs, nil
}

func backupDllFile() (isBackupNeeded bool, err error) {
	isBackupNeeded = false
	backupDirPath := filepath.Join(userPrefs.SteamDirPath, appSettings.DllDirPath, userPrefs.BackupDirName)

	_, err = os.Stat(backupDirPath)

	if os.IsNotExist(err) {
		isBackupNeeded = true

		os.Mkdir(backupDirPath, os.ModeDir)

		in, err := os.Open(filepath.Join(userPrefs.SteamDirPath, appSettings.DllDirPath, appSettings.DllFileName))
		if err != nil {
			return isBackupNeeded, err
		}
		defer in.Close()

		out, err := os.Create(filepath.Join(backupDirPath, appSettings.DllFileName))
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
	log.Info("[INFO] Opening backup file...")

	backupDllFile, err := os.Open(filepath.Join(userPrefs.SteamDirPath, appSettings.DllDirPath, userPrefs.BackupDirName, appSettings.DllFileName))
	if err != nil {
		return err
	}
	defer backupDllFile.Close()

	log.Info("[INFO] Successfully opened backup file.")
	log.Info("[INFO] Opening actual dll file for overwriting...")

	overwrittenDllFile, err := os.Create(filepath.Join(userPrefs.SteamDirPath, appSettings.DllDirPath, appSettings.DllFileName))
	if err != nil {
		return err
	}
	defer overwrittenDllFile.Close()

	log.Info("[INFO] Successfully opened actual dll file.")
	log.Info("[INFO] Rewriting data using replaced attribute values...")

	scanner := bufio.NewReader(backupDllFile)
	str, err := scanner.ReadString(appSettings.FileReaderDelimiter)
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

		str, err = scanner.ReadString(appSettings.FileReaderDelimiter)
		strNumber++
	}

	if err != nil && err != io.EOF {
		return err
	}

	return nil
}

func main() {
	log.Info("[INFO] Reading %s file...", configFileName)

	err := readConfigFile()
	if err != nil {
		log.Error("[ERROR] Could not read config file: %v", err)
	}

	log.Info("[INFO] Successfully read %s file.", configFileName)
	log.Info("[INFO] Reading %s file...", userPrefs.CameraValuesFileName)

	attrs, err := readCameraValuesFile()
	if err != nil {
		log.Error("[ERROR] Could not read camera values file: %v", err)
	}

	log.Info("[INFO] Successfully read %s file.", userPrefs.CameraValuesFileName)
	log.Info("[INFO] Backing up %s file...", appSettings.DllFileName)

	isBackupNeeded, err := backupDllFile()
	if err != nil {
		log.Error("[ERROR] Could not backup dll file: %v", err)
	}

	if isBackupNeeded {
		log.Info("[INFO] Successfully backed up %s file.", appSettings.DllFileName)
	} else {
		log.Info("[INFO] Backup file already exists, no backup needed.")
	}
	log.Info("[INFO] Changing camera attributes...")

	replacedLines, err := replaceAttributeValues(attrs)
	if err != nil {
		log.Error("[ERROR] Could not replace camera attributes: %v", err)
	}

	log.Info("[INFO] Successfully changed camera attributes.")
	log.Info("[INFO] Overwriting dll file...")

	err = rewriteDllFile(replacedLines)
	if err != nil {
		log.Error("[ERROR] Could not overwrite dll file: %v", err)
	}

	log.Print("[INFO] Done.")

	fmt.Scanf("%s")
}
