package database

import (
	"nameserver/cad"
)

func init() {
	db.AutoMigrate(&cad.Entry{})
}

func AddCadEntry(e *cad.Entry) error {
	return db.Create(e).Error
}

func GetEntries() ([]cad.Entry, error) {
	var entries []cad.Entry
	err := db.Model(&cad.Entry{}).Find(entries).Error
	return entries, err
}
