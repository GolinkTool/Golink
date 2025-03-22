package smt

import (
	"fmt"
	"github.com/GoLink/GoLink/DB"
	"strconv"
	"strings"
)

func GenerateSMTStmt(packageWithVersionList map[int]string) (string, []int) {
	var SMTStmt []string
	commonPrefix := "dep_"
	allPackageNameId := make(map[int]bool)
	var queue []int

	for packageId, versionList := range packageWithVersionList {
		SMTStmt = append(SMTStmt, addConstraintStmt(commonPrefix+strconv.Itoa(packageId), versionList))
		queue = append(queue, packageId)
	}

	var pkgId []int
	for len(queue) > 0 {
		current := queue[0] // package name Id
		queue = queue[1:]
		pkgId = append(pkgId, current)

		if allPackageNameId[current] {
			continue
		}
		allPackageNameId[current] = true

		results, err := DB.QueryDepsByPId(current)
		if err != nil {
			fmt.Println("Error querying deps:", err)
			continue
		}

		allDepId := make(map[int]bool)
		versionMap := ParseVersionList(packageWithVersionList[current])

		for _, line := range results {
			if depNameId, ok := line["dep_name_id"].(int64); !ok || depNameId == 0 {
				continue
			}

			packageVersionId := int(line["package_version_id"].(int64))
			if versionMap[packageVersionId] {

				SMTStmt = append(SMTStmt, addDepStmt(
					current,
					int(line["package_version_id"].(int64)),
					int(line["dep_name_id"].(int64)),
					line["dep_version_list"].(string),
				))
				allDepId[int(line["dep_name_id"].(int64))] = true
			}
		}

		for depId := range allDepId {
			queue = append(queue, depId)
		}

	}

	var declareSMTStmt []string
	for packageNameId, _ := range allPackageNameId {
		declareSMTStmt = append(declareSMTStmt, addDeclareStmt(packageNameId))
	}
	SMTStmt = append(declareSMTStmt, SMTStmt...)

	return finishCheckStmt(SMTStmt), pkgId
}

func addDepStmt(packageNameId int, version int, depNameId int, versionList string) string {
	condition := fmt.Sprintf("(= %s %d)", getShortName(packageNameId), version)

	versions := strings.Split(versionList, ",")
	var orClauses []string
	for _, v := range versions {
		if intValue, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			orClauses = append(orClauses, fmt.Sprintf("(= %s %d)", getShortName(depNameId), intValue))
		}
	}
	orExpression := fmt.Sprintf("(or %s)", strings.Join(orClauses, " "))

	return fmt.Sprintf("(assert (=> %s %s))", condition, orExpression)
}

func addConstraintStmt(shortName string, versionList string) string {

	versions := strings.Split(versionList, ",")
	var orClauses []string
	for _, v := range versions {
		if intValue, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			orClauses = append(orClauses, fmt.Sprintf("(= %s %d)", shortName, intValue))
		}
	}
	Stmt := fmt.Sprintf("(assert (or %s))", strings.Join(orClauses, " "))
	return Stmt
}

func addDeclareStmt(packageNameId int) string {
	return fmt.Sprintf("(declare-const %s Int)", getShortName(packageNameId))
}

func finishCheckStmt(Stmt []string) string {

	combined := strings.Join(Stmt, "\n")

	return fmt.Sprintf("%s\n(check-sat)", combined)
}

func getShortName(nameId int) string {
	return fmt.Sprintf("dep_%d", nameId)
}

func ParseVersionList(versionList string) map[int]bool {

	versionMap := make(map[int]bool)

	versionStrings := strings.Split(versionList, ",")

	for _, versionStr := range versionStrings {

		versionStr = strings.TrimSpace(versionStr)
		if versionStr == "" {
			continue
		}

		versionID, err := strconv.Atoi(versionStr)
		if err != nil {

			fmt.Printf("ParseVersionList: 无法解析版本号 '%s': %s\n", versionStr, err)
			continue
		}

		versionMap[versionID] = true
	}

	return versionMap
}
