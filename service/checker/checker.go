package checker

import (
	"IOCService/db"
	"IOCService/service/conditionVerdictProvider"
	"github.com/jinzhu/gorm"
	"log"
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

	result, err := provideVerdict(hashData)
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

func provideVerdict(hashData []db.HashData) ([]int64, error) {
	var iocIds []int64
	for _, hashDatum := range hashData {
		for iocId, condition := range hashDatum.Ioc {
			result, err := conditionVerdictProvider.GetVerdict(condition, hashDatum.RelatedAttributes)
			if err != nil {
				return []int64{}, err
			}

			if result {
				iocIds = append(iocIds, iocId)
			}
		}
	}

	return iocIds, nil
}
