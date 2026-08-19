package parsevalidateasn1

import (
	"fmt"
	"testing"
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
)

func TestParseMftFile(t *testing.T) {
	path := `F:\share\我的坚果云\RPKI\20241009_用openssl作为CA生成各种RPKI证书\files\ripencc\subcert\`
	m := make(map[string]string, 0)
	m[".mft"] = ".mft"
	files, _ := osutil.GetFilesInDir(path, m)
	fmt.Println("len(files):", len(files))
	start := time.Now()
	for _, file := range files {
		fileModel := &model.FileModel{
			FilePathName:     path + file,
			TempFilePathName: path + file,
		}

		fmt.Println("fileModel:", fileModel)
		mftModel := new(model.MftModel)
		stateModel := new(model.StateModel)
		err := ParseMftModelByAsn1(fileModel, mftModel, stateModel)
		if err != nil {
			fmt.Println(file, err)
			return
		}
		fmt.Println(file, "\n", jsonutil.MarshalJson(mftModel)+"\n=====================\n\n\n")
	}
	fmt.Println("end len(files):", len(files), "  time(s):", time.Since(start))
}
