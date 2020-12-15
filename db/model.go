package db

import "github.com/lib/pq"

type Condition struct {
	ID        int64  `db:"id"` //this is IOC_ID
	Condition string `db:"condition"`
}

type Attribute struct {
	ID   int64         `db:"id"`
	Refs pq.Int64Array `gorm:"type:integer[]"`
}

type HashData struct {
	Ioc               map[int64]string
	RelatedAttributes []int64
}

type HashDataFromDB struct {
	Ioc               string        `gorm:"type:json"`
	RelatedAttributes pq.Int64Array `gorm:"type:integer[]"`
}

func (HashDataFromDB) TableName() string {
	return "hash_table"
}

type IocFromHashTable struct {
	Ioc       int64  `json:"ioc_id"`
	Condition string `json:"condition"`
}

type AttributeType struct {
	ID          int64  `db:"id"`
	Description string `db:"description"`
}
