package trh

import (
	"github.com/ejfhp/ddb"
)

type Retrieve struct {
	diary *ddb.FBranch
	// env       *Environment
	outfolder string
}

// func NewRetrieve(env *Environment, diary *ddb.FBranch) *Retrieve {
// 	tr := trace.New().Source("retrieve.go", "Retrieve", "newRetrieve")
// 	if env.outFolder == "" {
// 		trail.Println(trace.Info("Output dir not set, using local flolder").Append(tr).UTC())
// 		env.outFolder = env.workingDir
// 	}
// 	retrieve := Retrieve{diary: diary, outfolder: env.outFolder, env: env}
// 	return &retrieve

// }

// func (cr *Retrieve) RetrieveAll() (int, error) {
// 	tr := trace.New().Source("retrieve.go", "Retrieve", "cmd")
// 	n, err := cr.diary.DowloadAll(flagOutputDir)
// 	if err != nil {
// 		trail.Println(trace.Info("error while downloadingAll").Append(tr).UTC().Add("address", cr.diary.BitcoinPublicAddress()).Add("ourFolder", cr.outfolder).Error(err))
// 		return 0, fmt.Errorf("error while downloadingAll files from address '%s' to folder '%s': %w", cr.diary.BitcoinPublicAddress(), cr.outfolder, err)
// 	}
// 	return n, nil
// }