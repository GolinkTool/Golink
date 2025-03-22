package extractCall

import (
	"bufio"
	"fmt"
	"github.com/GoLink/GoLink/DB"
	"go/ast"
	"go/build"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

func AstCall(rootPkgDir string, baseName string) ([]impt, FunCallArray) {

	var funCalls FunCallArray
	var allImps imptArr

	fs := token.NewFileSet()
	err := filepath.Walk(rootPkgDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			dirName := filepath.Base(path)
			if dirName == "vendor" || strings.HasPrefix(dirName, "_") ||
				strings.HasPrefix(dirName, ".") {
				return filepath.SkipDir // 跳过当前目录及其子目录
			}
		} else if filepath.Ext(info.Name()) == ".go" {

			if hasIgnoreBuildConstraint(path) {
				fmt.Printf("Skipping +build ignore %s\n", path)
				return nil
			}

			file, err := parser.ParseFile(fs, path, nil, 0)
			if err != nil {
				log.Printf("Error parsing file %s: %v", path, err)
				return nil
			}

			var imports []*ast.ImportSpec
			ast.Inspect(file, func(node ast.Node) bool {
				if imp, ok := node.(*ast.ImportSpec); ok {
					imports = append(imports, imp)
				}
				return true
			})

			var imps []impt

			for _, imp := range imports {
				impName := imp.Path.Value
				impName = strings.Trim(impName, "\"")

				if !strings.Contains(impName, "github.com") {
					continue
				}

				if isStdLib(impName) || strings.Contains(impName, baseName) {
					continue
				}

				var pkgName string
				if imp.Name != nil {
					pkgName = imp.Name.Name
				} else {

					pkgName = SelectPkgName(impName)
				}

				imp1 := impt{ImpName: impName, PkgName: pkgName}
				imps = append(imps, imp1)
			}

			if imps == nil {
				//跳过没有第三方库调用的文件
				return nil
			}
			allImps = append(allImps, imps...)

			var selectorExprs []*ast.SelectorExpr
			ast.Inspect(file, func(node ast.Node) bool {
				if selectorExpr, ok := node.(*ast.SelectorExpr); ok {
					selectorExprs = append(selectorExprs, selectorExpr)
				}
				return true
			})

			for _, selectorExpr := range selectorExprs {

				x := getTypeStringF(selectorExpr.X, fs)

				impName := belongsTo(x, imps)
				if impName != "" {

					f := selectorExpr.Sel.String()

					funcall := funCall{
						XdotF:   f,
						ImpName: impName,
					}
					funCalls = append(funCalls, funcall)
				}
			}

		}

		return nil
	})
	if err != nil {
		log.Fatalf("Error walking directory: %v", err)
	}

	allImps = removeDupImp(allImps)
	sort.Sort(allImps)
	funCalls = removeDupFunCall(funCalls)

	sort.Sort(funCalls)
	return allImps, funCalls
}

func getTypeStringF(expr ast.Expr, fileSet *token.FileSet) string {

	var typeStr string
	if expr != nil {
		var buf strings.Builder
		err := printer.Fprint(&buf, fileSet, expr)
		if err != nil {
			return "Unknown TypeString"
		}
		typeStr = buf.String()
	}
	return typeStr

}

func belongsTo(s string, imps []impt) string {
	for _, imp := range imps {
		if strings.EqualFold(s, imp.PkgName) {
			return imp.ImpName
		}
	}
	return ""
}

var (
	PathValid = regexp.MustCompile(`^([A-Za-z0-9-]+)(\.[A-Za-z0-9-]+)+(/[A-Za-z0-9-_.~]+)*$`)
)

func isStdLib(pkg string) bool {

	pkgInfo, err := build.Default.Import(pkg, "", build.IgnoreVendor)
	if err != nil {
		return false // 如果有错误，说明不是标准库
	}
	return pkgInfo.ImportPath == pkg && pkgInfo.Goroot
}

func hasIgnoreBuildConstraint(filename string) bool {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", filename, err)
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "// +build") && strings.Contains(line, "ignore") {
			return true
		}
		if len(strings.TrimSpace(line)) > 0 && !strings.HasPrefix(line, "//") {

			break
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file %s: %v\n", filename, err)
		return false
	}

	return false
}

func SelectPkgName(impUrl string) string {
	impToPkgs, err := DB.ReadPkgName(impUrl)
	if err != nil {
		fmt.Printf("Fail to select pkgName for %s :%v\n", impUrl, err)
		os.Exit(1)
	}
	if len(impToPkgs) < 1 {
		sp := strings.Split(impUrl, "/")
		pkgName := sp[len(sp)-1]
		return pkgName
	}
	if len(impToPkgs) > 1 {
		for _, impToPkg := range impToPkgs {
			if strings.HasSuffix(impToPkg.FileName, "_test.go") && strings.HasSuffix(impToPkg.FilePkgName, "_test") {
				continue
			}
			return impToPkg.FilePkgName
		}
	}

	return impToPkgs[0].FilePkgName
}
