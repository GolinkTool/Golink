package DB

// Api
type Api struct {
	ApiName  string
	ApiValue string
	ImpName  string
	FileName string
	VerList  string
}

// Pkg
type Pkg struct {
	PkgId   int
	PkgName string
	RepoURL string
}

// PkgUrl
type PkgUrl struct {
	Id         int
	ImportURL  string
	ModulePath string
	PkgId      string
	VerId      string
}

// Versions
type Versions struct {
	//ID          int
	PkgId       int
	VerID       int
	Version     string
	VN          int
	VersionTime string
}

type ImpToPkg struct {
	FileName    string
	FilePkgName string
	ImpUrl      string
}
