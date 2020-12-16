package conditionVerdictProvider

import (
	"fmt"
	"gval"
	"log"
	"regexp"
	"strconv"
	"strings"
)

func GetVerdict(condition string, attributeIds []int64) (bool, error) {
	rgxp, _ := regexp.Compile(`[0-9]+`)
	var attributeIdsInCondition []int64
	for _, match := range rgxp.FindAllString(condition, -1) {
		attributeId, err := strconv.ParseInt(match, 10, 64)
		if err != nil {
			return false, nil
		}
		attributeIdsInCondition = append(attributeIdsInCondition, attributeId)
	}

	for _, attributeIdInCondition := range attributeIdsInCondition {
		result := findElementInArray(attributeIdInCondition, attributeIds)

		condition = strings.Replace(condition, strconv.FormatInt(attributeIdInCondition, 10), strconv.FormatBool(result), -1)
	}

	condition = strings.Replace(condition, "OR", "||", -1)
	condition = strings.Replace(condition, "AND", "&&", -1)

	vars := map[string]interface{}{}
	value, err := gval.Evaluate(condition, &vars)
	if err != nil {
		return false, err
	}

	typeName := fmt.Sprintf("%T", value)
	if typeName != "bool" {
		return false, nil
	}

	result, err := getConditionResult(condition)
	if err != nil {
		return false, err
	}

	return result, nil
}

func findElementInArray(searchingElement int64, elements []int64) bool {
	elementFound := false
	for _, element := range elements {
		if searchingElement == element {
			elementFound = true
		}
	}

	return elementFound
}

func getConditionResult(condition string) (bool, error) {
	vars := map[string]interface{}{}
	value, err := gval.Evaluate(condition, &vars)
	if err != nil {
		log.Printf(err.Error())

		return false, err
	}

	result, ok := value.(interface{}).(bool)
	if ok == false {
		log.Printf("Unexpected type")

		return false, nil
	}

	return result, nil
}
