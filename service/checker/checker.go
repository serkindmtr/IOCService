package checker

import (
	"IOCService/db"
	"encoding/json"
	"github.com/jinzhu/gorm"
	"log"
	"sort"
	"strconv"
)

func Check(attributeValues []string) ([]int64, error) {
	database, err := db.Open("user=dmitry dbname=postgres host=localhost port=5432 sslmode=disable")
	if err != nil {
		return []int64{}, err
	}

	attributes, err := GetAttributesAssociatedWithIOCs(database, attributeValues)
	if err != nil {
		return []int64{}, err
	}

	iocIdToRelatedAttributeIdsMap := map[int64][]int64{}
	for _, attribute := range attributes {
		for _, iocId := range attribute.Refs {
			iocIdToRelatedAttributeIdsMap[iocId] = append(iocIdToRelatedAttributeIdsMap[iocId], attribute.ID)
		}
	}

	var iocs []int64
	for _, attribute := range attributes {
		for _, associatedIoc := range attribute.Refs {
			iocs = append(iocs, associatedIoc)
		}
	}

	conditions, err := GetConditionAssociatedWithIOCs(database, iocs)
	if err != nil {
		return []int64{}, err
	}

	var hashData []db.HashData
	for _, condition := range conditions {
		hashData = append(hashData, db.HashData{
			Ioc:               map[int64]string{condition.ID: condition.Condition},
			RelatedAttributes: iocIdToRelatedAttributeIdsMap[condition.ID],
		})
	}

	err = InsertHashData(database, hashData)
	if err != nil {
		return []int64{}, err
	}

	result, err := ProvideVerdict(database)
	if err != nil {
		return []int64{}, err
	}

	return result, nil
}

//SELECT id, refs FROM attribute WHERE value IN ([]string attributes)
func GetAttributesAssociatedWithIOCs(database *gorm.DB, attributes []string) ([]db.Attribute, error) {
	var rows []db.Attribute

	err := database.
		Select("id, refs").
		Table("attribute").
		Where("value IN (?)", attributes).
		Scan(&rows).Error
	if err != nil {
		log.Printf(err.Error())
	}

	return rows, nil
}

//SELECT id, condition FROM ioc WHERE ioc IN ([]int64 iocIds)
func GetConditionAssociatedWithIOCs(database *gorm.DB, iocIds []int64) ([]db.Condition, error) {
	var rows []db.Condition

	err := database.
		Select("id, condition").
		Table("ioc").
		Where("id IN (?)", iocIds).
		Scan(&rows).Error
	if err != nil {
		log.Printf(err.Error())
	}

	return rows, nil
}

//INSERT INTO hash_table VALUES ('{ "ioc_id": 1, "condition": "nil"}', '{1,2,3}')
func InsertHashData(database *gorm.DB, hashData []db.HashData) error {
	sqlInsert := "INSERT INTO hash_table VALUES "
	firstRaw := true
	for _, hashDatum := range hashData {
		if firstRaw {
			sqlInsert = sqlInsert + "("
			for iocId, conditionValue := range hashDatum.Ioc {
				sqlInsert = sqlInsert + "'{ \"ioc_id\": " + strconv.FormatInt(iocId, 10) + ", "
				sqlInsert = sqlInsert + "\"condition\": " + conditionValue + "}', "
			}
			sqlInsert = sqlInsert + "'{"
			firstAttributeId := true
			for _, attributeId := range hashDatum.RelatedAttributes {
				if firstAttributeId {
					sqlInsert = sqlInsert + strconv.FormatInt(attributeId, 10)
					firstAttributeId = false
				} else {
					sqlInsert = sqlInsert + ", " + strconv.FormatInt(attributeId, 10)
				}
			}
			sqlInsert = sqlInsert + "}')"
			firstRaw = false
		} else {
			sqlInsert = sqlInsert + ", ("
			for iocId, conditionValue := range hashDatum.Ioc {
				sqlInsert = sqlInsert + "'{ \"ioc_id\": " + strconv.FormatInt(iocId, 10) + ", "
				sqlInsert = sqlInsert + "\"condition\": " + conditionValue + "}', "
			}
			sqlInsert = sqlInsert + "'{"
			firstAttributeId := true
			for _, attributeId := range hashDatum.RelatedAttributes {
				if firstAttributeId {
					sqlInsert = sqlInsert + strconv.FormatInt(attributeId, 10)
					firstAttributeId = false
				} else {
					sqlInsert = sqlInsert + ", " + strconv.FormatInt(attributeId, 10)
				}
			}
			sqlInsert = sqlInsert + "}')"
		}
	}

	err := database.Exec(sqlInsert).Error
	if err != nil {
		log.Printf(err.Error())

		return err
	}

	return nil
}

func ProvideVerdict(database *gorm.DB) ([]int64, error) {
	var iocs []int64
	var hashData []db.HashDataFromDB

	err := database.Find(&hashData).Error
	if err != nil {
		log.Printf(err.Error())

		return []int64{}, err
	}

	for _, hashDatum := range hashData {
		var iocFromHashTable db.IocFromHashTable
		err = json.Unmarshal([]byte(hashDatum.Ioc), &iocFromHashTable)
		if err != nil {
			log.Printf("Unexpected error")
		}
		result, err := CheckCondition(iocFromHashTable.Condition, hashDatum.RelatedAttributes)
		if err != nil {
			return []int64{}, err
		}

		if result {
			iocs = append(iocs, iocFromHashTable.Ioc)
		}
	}

	return iocs, nil
}

func CheckCondition(condition string, attributeIds []int64) (bool, error) {
	if attributeId, err := strconv.ParseInt(condition, 10, 64); err == nil {
		i := sort.Search(len(attributeIds), func(i int) bool { return attributeId == attributeIds[i] })
		if i < len(attributeIds) && attributeIds[i] == attributeId {
			return true, nil
		} else {
			return false, nil
		}
	}

	return false, nil
}
