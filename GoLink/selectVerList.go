package DependencyAnalysis

import (
	"fmt"
	"github.com/GoLink/GoLink/DB"
	"os"
	"sort"
	"strings"
)

func SelectVerList(apiName string, pkgId int, impName string) string {

	apis, err := DB.ReadAPI(pkgId, apiName, impName)
	if err != nil {
		fmt.Printf("Fail to select api %s :%v\n", apiName, err)
		os.Exit(1)
	}

	var verList string
	var verListArr []string
	if len(apis) != 1 {
		for _, api := range apis {
			verListArr = append(verListArr, api.VerList)
		}
		sort.Strings(verListArr)
		verList = strings.Join(verListArr, ", ")
	} else {
		verList = apis[0].VerList
	}

	return verList
}
