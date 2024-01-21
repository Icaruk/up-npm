package packagejson

import (
	"fmt"
	"os"
	"time"
)

func CreatePackageJsonBackup() (bool, error) {
	date := time.Now().Format("2006-01-02-15-04-05")

	backupFileName := fmt.Sprintf("backup.%s.package.json", date)

	file, err := os.ReadFile("package.json")
	if err != nil {
		fmt.Println(err)
		return false, err
	}

	err = os.WriteFile(backupFileName, file, 0644)
	if err != nil {
		fmt.Println(err)
		return false, err
	}

	return true, nil
}
