package GoMod

import (
	"fmt"
	"os"
)

func GenerateGoMod(rootDir string, baseName string, requireDep []string) {

	content := "module " + baseName + "\n\ngo 1.23\n\n"

	for _, line := range requireDep {
		content += line + "\n"
	}

	filename := rootDir + "/go.mod"

	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Fail to create go.mod file：%v\n", err)
		return
	}

	_, err = file.WriteString(content)
	if err != nil {
		fmt.Printf("Error when write to go.mod file：%v\n", err)
		return
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	fmt.Println("Successfully create go.mod file!")

}
