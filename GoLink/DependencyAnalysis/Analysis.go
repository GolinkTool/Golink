package DependencyAnalysis

import (
	"github.com/GoLink/GoLink/DB"
	"github.com/GoLink/GoLink/extractCall"
)

func AnalysisImpVerList(baseName string, rootPkgDir string) (map[int]string, map[string]int) {

	Imps, funCalls := extractCall.AstCall(rootPkgDir, baseName)

	var allImps []string
	for _, imp := range Imps {
		allImps = append(allImps, imp.ImpName)
	}

	var pkgs []string
	for _, imp := range allImps {
		pkgs = append(pkgs, extractPkgName(imp))
	}

	pkgIdMap := make(map[string]int)
	for _, pkg := range pkgs {
		if _, ok := pkgIdMap[pkg]; ok {
			continue
		}
		pkgIdMap[pkg] = DB.QueryPackageId(pkg)
	}

	pkgVerMap := make(map[int]string)
	for _, fc := range funCalls {
		pkgName := extractPkgName(fc.ImpName)
		pkgId := pkgIdMap[pkgName]
		pkgVerMap[pkgId] = DB.SelectV0pkg(pkgId)
	}

	for _, fc := range funCalls {
		pkgName := extractPkgName(fc.ImpName)
		pkgId := pkgIdMap[pkgName]
		if pkgId == 0 {
			continue
		}

		VerList := SelectVerList(fc.XdotF, pkgId, fc.ImpName)
		if _, ok := pkgVerMap[pkgId]; ok {
			pkgVerMap[pkgId] = extractSameVer(pkgVerMap[pkgId], VerList)
		} else {
			pkgVerMap[pkgId] = VerList
		}
	}

	pkgVerMapNew := make(map[int]string)

	for id, ver := range pkgVerMap {
		if ver == "" {
			continue
		}
		pkgVerMapNew[id] = ver
	}

	return pkgVerMapNew, pkgIdMap
}
