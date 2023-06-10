package models

type Sequence struct {
	Key string `gorm:"primaryKey"`
	Val uint
}
