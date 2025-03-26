package DB

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	DbConn *sql.DB
	err    error
)

func init() {
	DbConn, err = sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/GOLINK?charset=utf8")
	if err != nil {
		fmt.Println(err)
	}

	DbConn.SetMaxOpenConns(200)

	DbConn.SetMaxIdleConns(20)

	DbConn.SetConnMaxLifetime(5 * time.Minute)

	if err := DbConn.Ping(); err != nil {
		fmt.Println("open DB fail")
	}

	fmt.Println("open DB SUCCESS!")
}

func QueryOptVersion(pkgId []int) map[int]int {
	var res = make(map[int]int)
	for _, id := range pkgId {
		pkgName := QueryPackageName(id)
		optVer, _ := QueryVersionMPMatch(id, pkgName)
		// choose the closest Version
		if len(optVer) > 0 {
			res[id] = optVer[len(optVer)-1]
		}
	}
	return res
}

func QueryVersionMPMatch(pkgId int, pkgName string) ([]int, error) {
	query := `
        SELECT version_id
        FROM package_url 
        WHERE package_id = ? and module_path = ?
    `

	rows, err := DbConn.Query(query, pkgId, pkgName)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	//var results []map[string]interface{}
	var result []int
	for rows.Next() {
		var packageVersionId int

		if err := rows.Scan(&packageVersionId); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		// Construct rowData map with appropriate checks for NULL values
		/*rowData := map[string]interface{}{
			"package_version_id": nilIfNull(packageVersionId),
		}*/
		result = append(result, packageVersionId)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return result, nil
}

func QueryPackageName(PkgId int) string {
	PackageName := ""
	stmtOut, err := DbConn.Prepare("SELECT package_name FROM go_packages WHERE id = ?")
	if err != nil {
		fmt.Errorf("QueryPackageName failed: %v", err)
		return ""
	}
	defer stmtOut.Close()

	rows, err := stmtOut.Query(PkgId)
	if err != nil {
		fmt.Errorf("QueryPackageName failed: %v", err)
		return ""
	}
	defer rows.Close()

	if rows.Next() {

		if err := rows.Scan(&PackageName); err != nil {
			fmt.Errorf("QueryPackageName failed: %v", err)
			return ""
		}
	}

	return PackageName
}

func QueryPackageId(PackageName string) int {
	PackageId := 0
	stmtOut, err := DbConn.Prepare("SELECT id FROM go_packages WHERE package_name = ?")
	if err != nil {
		fmt.Errorf("prepare insert packages_url statement failed: %v", err)
		return 0
	}
	defer stmtOut.Close()

	rows, err := stmtOut.Query(PackageName)
	if err != nil {
		fmt.Errorf("prepare insert packages_url statement failed: %v", err)
		return 0
	}
	defer rows.Close()

	if rows.Next() {

		if err := rows.Scan(&PackageId); err != nil {
			fmt.Errorf("prepare insert packages_url statement failed: %v", err)
			return 0
		}
	}

	return PackageId
}

func ReadModulePath(PkgId, versionId int) string {

	stmtOut, err := DbConn.Prepare("SELECT module_path FROM package_url WHERE package_id = ? and version_id = ?")
	if err != nil {
		fmt.Printf("Fail to Read ModulePath: %v\n", err)
		os.Exit(1)

		return ""
	}
	defer stmtOut.Close()

	rows, err := stmtOut.Query(PkgId, versionId)
	if err != nil {
		fmt.Printf("Fail to Read ModulePath: %v\n", err)
		os.Exit(1)

		return ""
	}
	defer rows.Close()

	var PkgArr []PkgUrl

	for rows.Next() {
		var ModulePath string

		if err := rows.Scan(&ModulePath); err != nil {
			fmt.Printf("Fail to Read ModulePath: %v\n", err)
			os.Exit(1)

			return ""
		}

		pkg := PkgUrl{
			ModulePath: ModulePath,
		}
		PkgArr = append(PkgArr, pkg)
	}

	if err := rows.Err(); err != nil {
		fmt.Printf("Fail to Read ModulePath: %v\n", err)
		os.Exit(1)

		return ""
	}

	if len(PkgArr) < 1 {
		return ""
	}

	return PkgArr[0].ModulePath
}

func ReadAPI(PkgId int, apiName string, impName string) ([]Api, error) {

	stmtOut, err := DbConn.Prepare("SELECT api_value, import_name, file_name, version_list FROM api_new WHERE package_id = ? and api_name= ? and import_name=? and api_value not like 'func (%'")
	if err != nil {
		return nil, err
	}
	defer stmtOut.Close()

	rows, err := stmtOut.Query(PkgId, apiName, impName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var APIArr []Api

	for rows.Next() {
		var APIValue, ImpName, FileName, VerList string

		if err := rows.Scan(&APIValue, &ImpName, &FileName, &VerList); err != nil {
			return nil, err
		}

		api := Api{
			ApiName:  apiName,
			ApiValue: APIValue,
			ImpName:  ImpName,
			FileName: FileName,
			VerList:  VerList,
		}
		APIArr = append(APIArr, api)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return APIArr, nil
}

func ReadPkgName(importUrl string) ([]ImpToPkg, error) {

	stmtOut, err := DbConn.Prepare("SELECT file_name, file_package_name FROM import_to_file_package WHERE import_url = ?")
	if err != nil {
		return nil, err
	}
	defer stmtOut.Close()

	rows, err := stmtOut.Query(importUrl)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var impToPkgArr []ImpToPkg

	for rows.Next() {
		var fileName, filePkgName string

		if err := rows.Scan(&fileName, &filePkgName); err != nil {
			return nil, err
		}

		impToPkg := ImpToPkg{
			FileName:    fileName,
			FilePkgName: filePkgName,
			ImpUrl:      importUrl,
		}
		impToPkgArr = append(impToPkgArr, impToPkg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return impToPkgArr, nil
}

func SelectV0pkg(PkgId int) string {

	stmtOut, err := DbConn.Prepare("SELECT version_id FROM package_versions WHERE package_id = ? and VN>=0 and VN<=1")
	if err != nil {
		fmt.Printf("SelectV0pkg: %s\n", err)
		os.Exit(1)
	}
	defer stmtOut.Close()

	rows, err := stmtOut.Query(PkgId)
	if err != nil {
		fmt.Printf("SelectV0pkg: %s\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	var voversions []string
	for rows.Next() {
		var versionId int

		if err := rows.Scan(&versionId); err != nil {
			fmt.Printf("SelectV0pkg: %s\n", err)
			os.Exit(1)
		}
		voversions = append(voversions, strconv.Itoa(versionId))
	}

	if err := rows.Err(); err != nil {
		fmt.Printf("SelectV0pkg: %s\n", err)
		os.Exit(1)
	}

	return strings.Join(voversions, ", ")
}

func QueryDepsByPId(packageId int) ([]map[string]interface{}, error) {

	query := `
        SELECT package_version_id,dep_name_id, dep_version_list 
        FROM deps 
        WHERE package_name_id = ?
    `

	rows, err := DbConn.Query(query, packageId)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var packageVersionId sql.NullInt64
		var depNameId sql.NullInt64
		var depVersionList sql.NullString

		if err := rows.Scan(&packageVersionId, &depNameId, &depVersionList); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		// Construct rowData map with appropriate checks for NULL values
		rowData := map[string]interface{}{
			"package_version_id": nilIfNull(packageVersionId),
			"dep_name_id":        nilIfNull(depNameId),
			"dep_version_list":   nilIfNull(depVersionList),
		}
		results = append(results, rowData)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return results, nil
}

func QueryPackageNameAndVersionByIds(packageId int, versionId int) (map[string]interface{}, error) {

	query := `
        SELECT gp.package_name, pv.version, pv.VN
        FROM go_packages gp
        JOIN package_versions pv ON gp.id = pv.package_id
        WHERE pv.version_id = ? AND gp.id = ?
    `

	row := DbConn.QueryRow(query, versionId, packageId)

	var packageName sql.NullString
	var version sql.NullString
	var VN int

	if err := row.Scan(&packageName, &version, &VN); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no package found with given package_id: %d  and version_id:%d ", packageId, versionId)
		}
		return nil, fmt.Errorf("error scanning row: %w", err)
	}

	// Construct result map with appropriate checks for NULL values
	result := map[string]interface{}{
		"package_name": nilIfNull(packageName),
		"version":      nilIfNull(version),
	}

	return result, nil
}

// nilIfNull is a helper function that checks for SQL null values and returns the value or nil
func nilIfNull(val interface{}) interface{} {
	switch v := val.(type) {
	case sql.NullInt64:
		if v.Valid {
			return v.Int64
		}
		return nil
	case sql.NullString:
		if v.Valid {
			return v.String
		}
		return nil
	default:
		return nil
	}
}

func QueryVN(packageId int, versionId int) (int, error) {

	query := `
        SELECT pv.VN
        FROM go_packages gp
        JOIN package_versions pv ON gp.id = pv.package_id
        WHERE pv.version_id = ? AND gp.id = ?
    `

	row := DbConn.QueryRow(query, versionId, packageId)

	var VN int

	if err := row.Scan(&VN); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Printf("Fail to Query VN: %v", err)
			return 0, fmt.Errorf("no package found with given package_id: %d  and version_id:%d ", packageId, versionId)
		}
		fmt.Printf("Fail to Query VN: %v", err)
		return 0, fmt.Errorf("error scanning row: %w", err)
	}

	// Construct result map with appropriate checks for NULL values

	return VN, nil
}
