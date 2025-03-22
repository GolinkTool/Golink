package extractCall

type funCall struct {
	XdotF   string
	ImpName string
}

type FunCallArray []funCall

func (fcs FunCallArray) Len() int {
	return len(fcs)
}

func (fcs FunCallArray) Less(i, j int) bool {
	return fcs[i].ImpName < fcs[j].ImpName
}

func (fcs FunCallArray) Swap(i, j int) {
	fcs[i], fcs[j] = fcs[j], fcs[i]
}

func removeDupFunCall(funCalls []funCall) []funCall {

	seen := make(map[funCall]bool)

	var uniqueFC []funCall

	for _, FC := range funCalls {
		if !seen[FC] {
			seen[FC] = true
			uniqueFC = append(uniqueFC, FC)
		}
	}

	return uniqueFC
}

type impt struct {
	ImpName string
	PkgName string
}

type imptArr []impt

func (impts imptArr) Len() int {
	return len(impts)
}

func (impts imptArr) Less(i, j int) bool {
	return impts[i].ImpName < impts[j].ImpName
}

func (impts imptArr) Swap(i, j int) {
	impts[i], impts[j] = impts[j], impts[i]
}

func removeDupImp(apiList []impt) []impt {

	seen := make(map[impt]bool)

	var uniqueAPI []impt

	for _, num := range apiList {
		if !seen[num] {
			seen[num] = true
			uniqueAPI = append(uniqueAPI, num)
		}
	}

	return uniqueAPI
}
