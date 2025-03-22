package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/GoLink/GoLink/DB"
	"github.com/GoLink/GoLink/DependencyAnalysis"
	"github.com/GoLink/GoLink/GoMod"
	"github.com/GoLink/GoLink/smt"
	"github.com/GoLink/GoLink/utils"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var expressionFile string

func SolverDeps(directDeps map[int]string, optDeps map[int]int) []string {

	var requireDeps []string
	smtStmts, _ := smt.GenerateSMTStmt(directDeps)
	//fmt.Println(pkgId)

	err := utils.WriteToFile(smtStmts, expressionFile)
	if err != nil {
		return nil
	}

	args := []string{expressionFile}
	if len(optDeps) > 0 {
		for packageId, versionId := range optDeps {
			args = append(args, fmt.Sprintf("dep_%d:%d", packageId, versionId))
		}
	}

	output := ""
	output, err = utils.ExecuteCommand("./GoLink_z3", args)
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "dep") {
			packageNameId, _ := strconv.Atoi(strings.Split(strings.Split(line, ":")[0], "_")[1])
			packageVersionId, _ := strconv.Atoi(strings.Split(line, ":")[1])

			results, _ := DB.QueryPackageNameAndVersionByIds(packageNameId, packageVersionId)
			pkgName := results["package_name"]
			pkgVer := results["version"]
			pkgVer = VNIncompatible(packageNameId, packageVersionId, fmt.Sprintf("%s", pkgVer))
			if packageVersionId > 0 {
				modulePath := DB.ReadModulePath(packageNameId, packageVersionId)

				if modulePath == "" || pkgName == modulePath {
					requireDeps = append(requireDeps, fmt.Sprintf("require %s %s", pkgName, pkgVer))
				} else {
					requireDeps = append(requireDeps, fmt.Sprintf("replace %s => %s %s", pkgName, modulePath, pkgVer))

				}
			}
		}
	}
	return requireDeps
}

var (
	baseName   = flag.String("baseName", "Go_Example", "full project name")
	rootPkgDir = flag.String("projectDir", "./", "Your project Directory")
	DatabaseIp = flag.String("database_ip", "127.0.0.1", "database ip")
	err        error
)

func init() {
	flag.Parse()
	DB.DbConn, err = sql.Open("mysql", "golink:123456@tcp("+*DatabaseIp+":3306)/golink?charset=utf8")
	if err != nil {
		fmt.Println(err)
	}

	DB.DbConn.SetMaxOpenConns(200)

	DB.DbConn.SetMaxIdleConns(20)

	DB.DbConn.SetConnMaxLifetime(5 * time.Minute)

	if err := DB.DbConn.Ping(); err != nil {
		fmt.Println("open DB fail")
		fmt.Println(err)
	} else {
		fmt.Println("open DB SUCCESS!")
	}
}

func main() {
	// ./GoLink -baseName=Example/Go_Example -projectDir=../Example/Go_Example
	expressionFile = "./expression.txt"
	fmt.Printf("ProjectDir: %s\n", *rootPkgDir)
	fmt.Printf("BaseName: %s\n", *baseName)

	startTime := time.Now()
	handleSingleRepo(*rootPkgDir, *baseName)
	endTime := time.Now()
	durations := endTime.Sub(startTime)
	fmt.Println(durations)

	DB.DbConn.Close()
}

func handleSingleRepo(rootPkgDir, baseName string) {

	directDeps := make(map[int]string)
	directDeps, _ = DependencyAnalysis.AnalysisImpVerList(baseName, rootPkgDir)

	var requireDep []string

	optDeps := make(map[int]int)

	requireDep = SolverDeps(directDeps, optDeps)

	GoMod.GenerateGoMod(rootPkgDir, baseName, requireDep)

}

func VNIncompatible(packageId int, versionId int, version string) string {
	VN, _ := DB.QueryVN(packageId, versionId)

	re := regexp.MustCompile(`^v([0-9]+)`)

	if VN == 0 {
		matches := re.FindStringSubmatch(version)
		if matches != nil && len(matches) > 1 {

			extractedNumberStr := matches[1]

			extractedNumber, _ := strconv.Atoi(extractedNumberStr)
			if extractedNumber > 1 {
				return version + "+incompatible"
			}
		}
	}
	return version
}
