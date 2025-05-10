package test

import "os"

func GetConFigFile(seeker string) string {
	// get current directory
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return dir + seeker + "config.yaml"
}
