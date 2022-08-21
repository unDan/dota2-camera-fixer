package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"time"

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

var dllFilePath string
var backupDirPath string
var saveFilePath string

var log logger.Logger = logger.NewLogger(true)

func readConfigFile() error {
	iniData, err := ini.Load(configFileName)

	if err != nil {
		return err
	}

	err = iniData.Section("UserPrefs").MapTo(&userPrefs)

	if err != nil {
		return err
	}

	err = iniData.Section("AppSettings").MapTo(&appSettings)

	if err != nil {
		return err
	}

	log = logger.NewLogger(userPrefs.ShowLogInfo)

	dllFilePath = filepath.Join(userPrefs.SteamDirPath, appSettings.DllDirPath, appSettings.DllFileName)
	backupDirPath = filepath.Join(userPrefs.SteamDirPath, appSettings.DllDirPath, userPrefs.BackupDirName)
	saveFilePath = filepath.Join(userPrefs.SteamDirPath, appSettings.DllDirPath, userPrefs.BackupDirName, appSettings.SaveFileName)

	return nil
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

	log.Info("Opening dota 2 client.dll file...")

	dllFile, err := os.Open(dllFilePath)
	if err != nil {
		return nil, err
	}
	defer dllFile.Close()

	log.Info("Successfully opened client.dll file.")
	log.Info("Scanning client.dll file...")

	reader := bufio.NewReader(dllFile)
	str, err := reader.ReadString(appSettings.FileReaderDelimiter)
	currentStrNumber := 1

	for err == nil {
		for _, attr := range attrs {
			if !strings.Contains(str, attr.AttributeName) {
				continue
			}

			changedStr := ""

			for err == nil {
				str, err = reader.ReadString(appSettings.FileReaderDelimiter)
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

			log.Info("Successfully changed attribute '%s' value: %d -> %d.", attr.AttributeName, attr.OldValue, attr.NewValue)
		}

		str, err = reader.ReadString(appSettings.FileReaderDelimiter)
		currentStrNumber++
	}

	if err != nil && err != io.EOF {
		return nil, err
	}

	return replacedStrs, nil
}

func isBackupDirExists() bool {
	_, err := os.Stat(backupDirPath)

	return !os.IsNotExist(err)
}

func hasGameBeenUpdated() (bool, error) {
	_, err := os.Stat(saveFilePath)

	if os.IsNotExist(err) {
		return true, nil
	}

	fbytes, err := ioutil.ReadFile(saveFilePath)
	if err != nil {
		return false, err
	}

	lastSavedDllModifyDate, err := time.Parse(appSettings.TimeFormat, string(fbytes))

	if err != nil {
		return false, err
	}

	dllStats, err := os.Stat(dllFilePath)

	if err != nil {
		return false, err
	}

	actualDllModifyDate := dllStats.ModTime().UTC()

	if actualDllModifyDate.Unix() <= lastSavedDllModifyDate.Unix() {
		return false, nil
	}

	return true, nil
}

func saveDllLastModifyDate() error {
	saveFile, err := os.Create(saveFilePath)
	if err != nil {
		return err
	}
	defer saveFile.Close()

	dllStats, err := os.Stat(dllFilePath)
	if err != nil {
		return err
	}

	saveFile.WriteString(dllStats.ModTime().UTC().Format(appSettings.TimeFormat))

	return nil
}

func backupDllFile() (err error) {
	from, err := os.Open(dllFilePath)
	if err != nil {
		return err
	}
	defer from.Close()

	to, err := os.Create(filepath.Join(backupDirPath, appSettings.DllFileName))
	if err != nil {
		return err
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	if err != nil {
		return err
	}

	return to.Close()
}

func cleanDllFileChanges() error {
	from, err := os.Open(filepath.Join(backupDirPath, appSettings.DllFileName))
	if err != nil {
		return err
	}
	defer from.Close()

	to, err := os.Create(dllFilePath)
	if err != nil {
		return err
	}
	defer to.Close()

	_, err = io.Copy(from, to)
	if err != nil {
		return err
	}

	return to.Close()
}

func rewriteDllFile(replacedStrs map[int]string) error {
	backupDllFile, err := os.Open(filepath.Join(backupDirPath, appSettings.DllFileName))
	if err != nil {
		return err
	}
	defer backupDllFile.Close()

	overwrittenDllFile, err := os.Create(dllFilePath)
	if err != nil {
		return err
	}
	defer overwrittenDllFile.Close()

	reader := bufio.NewReader(backupDllFile)
	str, err := reader.ReadString(appSettings.FileReaderDelimiter)
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

		str, err = reader.ReadString(appSettings.FileReaderDelimiter)
		strNumber++
	}

	if err != nil && err != io.EOF {
		return err
	}

	return nil
}

func main() {
	log.Info("Reading %s file...", configFileName)

	err := readConfigFile()
	if err != nil {
		log.Error("Could not read config file: %v", err)
	}

	log.Info("Successfully read %s file.", configFileName)
	log.Info("Reading %s file...", userPrefs.CameraValuesFileName)

	attrs, err := readCameraValuesFile()
	if err != nil {
		log.Error("Could not read camera values file: %v", err)
	}

	log.Info("Successfully read %s file.", userPrefs.CameraValuesFileName)

	isBackupNeeded := false

	// if backup dir does not exist (the first app launch after installation) then create it
	// otherwise decide whether to backup dll or not depending on the whether the game has been updated
	if !isBackupDirExists() {
		os.Mkdir(filepath.Join(userPrefs.SteamDirPath, appSettings.DllDirPath, userPrefs.BackupDirName), os.ModeDir)
		isBackupNeeded = true

		log.Info("Backup file does not exist. Backing up %s file...", appSettings.DllFileName)
	} else {
		log.Info("Checking whether the game has been updated...")

		isBackupNeeded, err = hasGameBeenUpdated()

		if err != nil {
			log.Error("Could not check whether the game has been updated: %v", err)
		}

		if isBackupNeeded {
			log.Info("Game has been updated. Backing up %s file...", appSettings.DllFileName)
		} else {
			log.Info("Game has not been updated, no backup needed.")
		}
	}

	if isBackupNeeded {
		err = backupDllFile()
		if err != nil {
			log.Error("Could not backup dll file: %v", err)
		}

		log.Info("Successfully backed up %s file", appSettings.DllFileName)
	} else {
		err = cleanDllFileChanges() // rewrite the whole dll file with contents from backup file (return dll to its original state)
		if err != nil {
			log.Error("Could not rewrite dll file to its original state: %v", err)
		}
	}

	log.Info("Changing camera attributes...")

	replacedLines, err := replaceAttributeValues(attrs)
	if err != nil {
		log.Error("Could not replace camera attributes: %v", err)
	}

	log.Info("Successfully changed camera attributes.")
	log.Info("Overwriting dll file...")

	err = rewriteDllFile(replacedLines)
	if err != nil {
		log.Error("Could not overwrite dll file: %v", err)
	}

	log.Info("Saving application state...")

	err = saveDllLastModifyDate()
	if err != nil {
		log.Error("Could not save application state: %v", err)
	}

	log.Info("Successfully saved application state.")

	log.Print("Done.")

	fmt.Scanf("%s")
}
