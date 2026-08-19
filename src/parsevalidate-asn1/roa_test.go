package parsevalidateasn1

import (
	"fmt"
	"testing"
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
)

func TestParseRoaFile(t *testing.T) {
	path := `G:\Download\cert\roa\`
	m := make(map[string]string, 0)
	m[".roa"] = ".roa"
	files, _ := osutil.GetFilesInDir(path, m)
	fmt.Println("len(files):", len(files))
	start := time.Now()
	for _, file := range files {
		fileModel := &model.FileModel{
			FilePathName:     path + file,
			TempFilePathName: path + file,
		}

		fmt.Println("fileModel:", fileModel)
		roaModel := new(model.RoaModel)
		stateModel := new(model.StateModel)
		err := ParseRoaModelByAsn1(fileModel, roaModel, stateModel)
		if err != nil {
			fmt.Println(file, err)
			return
		}
		fmt.Println(file, "\n", jsonutil.MarshalJson(roaModel)+"\n=====================\n\n\n")
	}
	fmt.Println("end len(files):", len(files), "  time(s):", time.Since(start))
}
