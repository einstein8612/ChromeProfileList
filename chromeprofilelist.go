package chromeprofilelist

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"sync"
)

var LocalAppDataPath string
var HomeDirectory string

var debug bool = false
var allLocations map[string][]string

type ChromeProfile struct {
	DisplayName          string
	ProfileDirectoryName string
	ProfileDirectory     string
	ProfilePictureURL    string
}

type chromeProfilePreferences struct {
	Profile struct {
		GaiaInfoPictureURL string `json:"gaia_info_picture_url"`
		Name               string `json:"name"`
	} `json:"profile"`
}

func init() {
	LocalAppDataPath = os.Getenv("localappdata")
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("There was an error getting the home directory, err: %v\n", err)
		fmt.Println("If used on Linux or Darwin this package will be a stub")
		return
	}
	HomeDirectory = home

	allLocations = make(map[string][]string)
	allLocations["windows"] = []string{
		LocalAppDataPath + "\\Google\\Chrome\\User Data",
		LocalAppDataPath + "\\Google\\Chrome SxS\\User Data",
		LocalAppDataPath + "\\Chromium\\User Data",
	}
	allLocations["linux"] = []string{
		HomeDirectory + "/.config/google-chrome",
		HomeDirectory + "/.config/google-chrome-beta",
		HomeDirectory + "/.config/chromium",
	}
	allLocations["darwin"] = []string{
		HomeDirectory + "/Library/Application Support/Google/Chrome",
		HomeDirectory + "/Library/Application Support/Google/Chrome Canary",
		HomeDirectory + "/Library/Application Support/Chromium",
	}
}

func EnableDebug() {
	debug = true
}

func DisableDebug() {
	debug = false
}

func GetAllProfiles() (profiles []ChromeProfile, err error) {
	locations, supported := allLocations[runtime.GOOS]
	if !supported {
		return profiles, errors.New("your os is not supported")
	}
	var wg sync.WaitGroup
	wg.Add(len(locations))
	for _, location := range locations {
		go func(location string) {
			readProfiles, err := GetProfileFromUserdata(location)
			if err != nil {
				if debug {
					fmt.Printf("error reading profiles, err: %v", err)
				}
				return
			}
			profiles = append(profiles, readProfiles...)
			wg.Done()
		}(location)
	}
	wg.Wait()
	return profiles, err
}

func GetProfileFromUserdata(location string) (profiles []ChromeProfile, err error) {
	files, err := ioutil.ReadDir(location)
	if err != nil {
		if debug {
			fmt.Printf("error reading directory, err: %v", err)
		}
		return profiles, err
	}
	for _, file := range files {
		if file.Name() == "System Profile" {
			continue
		}
		if !file.IsDir() {
			continue
		}

		if _, err := os.Stat(location + "/" + file.Name() + "/Preferences"); err != nil && os.IsNotExist(err) {
			continue
		}

		fileBytes, err := ioutil.ReadFile(location + "/" + file.Name() + "/Preferences")
		if err != nil {
			if debug {
				fmt.Printf("error reading file, error: %v", err)
			}
			continue
		}
		var profilePreferences chromeProfilePreferences
		err = json.Unmarshal(fileBytes, &profilePreferences)
		if err != nil {
			if debug {
				fmt.Printf("error unmarshalling file, error: %v", err)
			}
			continue
		}

		profiles = append(profiles, ChromeProfile{
			DisplayName:          profilePreferences.Profile.Name,
			ProfileDirectoryName: file.Name(),
			ProfileDirectory:     location + "/" + file.Name(),
			ProfilePictureURL:    profilePreferences.Profile.GaiaInfoPictureURL,
		})
	}
	return profiles, err
}
