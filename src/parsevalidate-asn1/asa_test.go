package parsevalidateasn1

import (
	"fmt"
	"testing"
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
)

func TestParseAsaFile(t *testing.T) {
	path := `G:\Download\cert\asa\`
	m := make(map[string]string, 0)
	m[".asa"] = ".asa"
	files, _ := osutil.GetFilesInDir(path, m)
	fmt.Println("len(files):", len(files))
	start := time.Now()
	for _, file := range files {
		fileModel := &model.FileModel{
			FilePathName:     path + file,
			TempFilePathName: path + file,
		}

		fmt.Println("file:", file)
		asaModel := new(model.AsaModel)
		stateModel := new(model.StateModel)
		err := ParseAsaModelByAsn1(fileModel, asaModel, stateModel)
		if err != nil {
			fmt.Println(file, err)
			return
		}
		fmt.Println(file, "\n"+jsonutil.MarshalJson(asaModel)+"\n"+jsonutil.MarshalJson(stateModel)+"\n=====================\n\n\n")
	}
	fmt.Println("end len(files):", len(files), "  time(s):", time.Since(start))
}
