// +build linux
package chromeprofilelist

import (
	"fmt"
	"sync"
)

var locations [3]string = [3]string{
	"/.config/google-chrome",
	"/.config/google-chrome-beta",
	"/.config/chromium",
}

func getAllProfiles() (profiles []ChromeProfile, err error) {
	if HomeDirectory == "" {
		return profiles, err
	}

	var wg sync.WaitGroup
	wg.Add(len(locations))
	for _, location := range locations {
		go func(location string) {
			readProfiles, err := GetProfileFromUserdata(HomeDirectory + location)
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
