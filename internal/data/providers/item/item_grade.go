package item

type ItemGrade int32

const (
	ItemGradeNormal ItemGrade = 0
	ItemGradeRare   ItemGrade = 1
	ItemGradeEpic   ItemGrade = 2
	ItemGradeUnique ItemGrade = 3
)

func (g ItemGrade) MatchesOptionID(optionID int32) bool {
	switch g {
	case ItemGradeNormal:
		return optionID < 10000
	case ItemGradeRare:
		return optionID >= 10000 && optionID < 20000
	case ItemGradeEpic:
		return optionID >= 20000 && optionID < 30000
	case ItemGradeUnique:
		return optionID >= 30000 && optionID < 40000
	}
	return false
}
