// +build windows
package chromeprofilelist

import (
	"fmt"
	"sync"
)

var locations [3]string = [3]string{
	"\\Google\\Chrome\\User Data",
	"\\Google\\Chrome SxS\\User Data",
	"\\Chromium\\User Data",
}

func getAllProfiles() (profiles []ChromeProfile, err error) {
	var wg sync.WaitGroup
	wg.Add(len(locations))
	for _, location := range locations {
		go func(location string) {
			readProfiles, err := GetProfileFromUserdata(LocalAppDataPath + location)
			if err != nil {
				if debug {
					fmt.Printf("error reading profiles, err: %v", err)
				}
				wg.Done()
				return
			}
			profiles = append(profiles, readProfiles...)
			wg.Done()
		}(location)
	}
	wg.Wait()
	return profiles, err
}
