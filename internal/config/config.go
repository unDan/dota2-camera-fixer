package config

type UserPrefs struct {
	SteamDirPath         string
	ShowLogInfo          bool
	BackupDirName        string
	CameraValuesFileName string
}

type AppSettings struct {
	DllDirPath          string
	DllFileName         string
	SaveFileName        string
	FileReaderDelimiter byte
	TimeFormat          string
}
