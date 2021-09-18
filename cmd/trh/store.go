package main

import (
	"github.com/ejfhp/ddb"
)

type Store struct {
	diary *ddb.FBranch
	env   *Environment
}

func NewStore(env *Environment, diary *ddb.FBranch) *Store {
	store := Store{diary: diary, env: env}
	return &store
}

// func (s *Store) Store(filename string) ([]string, error) {
// 	tr := trace.New().Source("store.go", "Store", "Store")

// 	entry, err := ddb.NewEntryFromFile(filepath.Base(filename), filename)

// 	if err != nil {
// 		trail.Println(trace.Alert("error while opening file").Append(tr).UTC().Add("filename", filename).Error(err))
// 		return nil, fmt.Errorf("error opening file '%s': %v", filename, err)
// 	}
// 	txids, err := s.diary.CastEntry(entry)
// 	if err != nil {
// 		trail.Println(trace.Alert("error while storing file").Append(tr).UTC().Add("filename", filename).Add("address", s.diary.BitcoinPublicAddress()).Error(err))
// 		return nil, fmt.Errorf("error while storing file '%s' on address '%s': %w", filename, s.diary.BitcoinPublicAddress(), err)
// 	}
// 	return txids, nil
// }

// func (s *Store) Estimate(filename string) (ddb.Satoshi, error) {
// 	tr := trace.New().Source("store.go", "Store", "Estimate")

// 	entry, err := ddb.NewEntryFromFile(filepath.Base(filename), filename)

// 	if err != nil {
// 		trail.Println(trace.Alert("error while opening file").Append(tr).UTC().Add("filename", filename).Error(err))
// 		return 0, fmt.Errorf("error opening file '%s': %v", filename, err)
// 	}
// 	fee, err := s.diary.EstimateFee(entry)
// 	if err != nil {
// 		trail.Println(trace.Alert("error while storing file").Append(tr).UTC().Add("filename", filename).Add("address", s.diary.BitcoinPublicAddress()).Error(err))
// 		return 0, fmt.Errorf("error while storing file '%s' on address '%s': %w", filename, s.diary.BitcoinPublicAddress(), err)
// 	}
// 	return fee, nil
// }
