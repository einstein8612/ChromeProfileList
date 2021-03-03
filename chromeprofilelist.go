package chromeprofilelist

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
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
}

func EnableDebug() {
	debug = true
}

func DisableDebug() {
	debug = false
}

func GetAllProfiles() (profiles []ChromeProfile, err error) {
	return getAllProfiles()
}

func GetProfileFromUserdata(location string) (profiles []ChromeProfile, err error) {
	files, err := ioutil.ReadDir(location)
	if err != nil {
		if debug {
			fmt.Printf("error reading directory, err: %v", err)
		}
		return profiles, err
	}
	var wg sync.WaitGroup
	wg.Add(len(files))
	for _, file := range files {
		go func(file os.FileInfo) {
			func() {
				if file.Name() == "System Profile" {
					return
				}
				if !file.IsDir() {
					return
				}

				if _, err := os.Stat(location + "/" + file.Name() + "/Preferences"); err != nil && os.IsNotExist(err) {
					return
				}

				fileBytes, err := ioutil.ReadFile(location + "/" + file.Name() + "/Preferences")
				if err != nil {
					if debug {
						fmt.Printf("error reading file, error: %v", err)
					}
					return
				}
				var profilePreferences chromeProfilePreferences
				err = json.Unmarshal(fileBytes, &profilePreferences)
				if err != nil {
					if debug {
						fmt.Printf("error unmarshalling file, error: %v", err)
					}
					return
				}

				profiles = append(profiles, ChromeProfile{
					DisplayName:          profilePreferences.Profile.Name,
					ProfileDirectoryName: file.Name(),
					ProfileDirectory:     location + "/" + file.Name(),
					ProfilePictureURL:    profilePreferences.Profile.GaiaInfoPictureURL,
				})
			}()
			wg.Done()
		}(file)
	}
	wg.Wait()
	return profiles, err
}
