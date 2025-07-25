package application

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/MrM2025/rpforcalc/tree/master/calc_go/pkg/errorStore"
)

type DCalc struct {
}

type TCalc struct {
	history map[time.Time]map[string]string
}

type IHistory interface {
	Init()
	Calc(Expression string) (float64, error)
	GetCalcHistory() map[time.Time]map[string]string
	RemoveHistory()
}

const isLeftParenthesis = 1
const isRightParenthesis = 2
const isNotParenthesis = 0
const isMultiplication = 10
const isDivision = 20
const isAddition = 30
const isSubtraction = 40
const isNotOperation = 0
const isPoint = 100
const isNotSeparator = 0

func (d *DCalc) IsNumber(char byte) bool {
	const numbers = "1234567890"
	for index, _ := range numbers {
		if numbers[index] == char {
			return true
		}
	}
	return false
}

func (d *DCalc) IsParenthesis(char byte) int {
	if string(char) == "(" {
		return isLeftParenthesis
	}
	if string(char) == ")" {
		return isRightParenthesis
	}
	return isNotParenthesis
}

func (d *DCalc) IsOperator(char byte) int {
	if string(char) == "*" {
		return isMultiplication
	} else if string(char) == "/" {
		return isDivision
	} else if string(char) == "+" {
		return isAddition
	} else if string(char) == "-" {
		return isSubtraction
	}
	return isNotOperation
}

func (d *DCalc) IsSeparator(char byte) int {
	if string(char) == "." {
		return isPoint
	}
	return isNotSeparator
}

func getPryority(operator int) int {
	mapofoperators := map[int]int{
		isMultiplication: 2,
		isDivision:       2,
		isAddition:       1,
		isSubtraction:    1,
	}
	pryority := mapofoperators[operator]
	return pryority
}

func extractNum(Expression string, indexofnum int, sliceofnums []float64, negative bool) ([]float64, int, error) {
	var d DCalc
	var num string
	var index int
	var length int = len(Expression)
	var numfloat64 float64
	var converr error

	for nextnotnumindex := indexofnum; nextnotnumindex < length; nextnotnumindex++ {
		if d.IsNumber(Expression[nextnotnumindex]) || d.IsSeparator(Expression[nextnotnumindex]) != 0 {
			num += string(Expression[nextnotnumindex])
		}
		if !d.IsNumber(Expression[nextnotnumindex]) && d.IsSeparator(Expression[nextnotnumindex]) == 0 {
			numfloat64, converr = strconv.ParseFloat(num, 64)
			if numfloat64 == 0 && converr != nil {
				return nil, indexofnum, converr
			}
			if negative && d.IsParenthesis(Expression[nextnotnumindex]) != isRightParenthesis {
				numfloat64 = -numfloat64
			} else if negative && d.IsParenthesis(Expression[nextnotnumindex]) == isRightParenthesis {
				numfloat64 = -numfloat64
				nextnotnumindex += 1
			}
			sliceofnums = append(sliceofnums, numfloat64)
			return sliceofnums, nextnotnumindex, nil
		}
		index = nextnotnumindex
	}

	numfloat64, converr = strconv.ParseFloat(num, 64)
	if numfloat64 == 0 && converr != nil {
		return nil, indexofnum, converr
	}
	if negative && d.IsParenthesis(Expression[index]) != isRightParenthesis {
		numfloat64 = -numfloat64
	} else if negative && d.IsParenthesis(Expression[index]) == isRightParenthesis {
		numfloat64 = -numfloat64
		index += 1
	}
	sliceofnums = append(sliceofnums, numfloat64)

	return sliceofnums, index, nil
}

func popNum(sliceofnums []float64, numtopop int) ([]float64, []float64, error) {

	var poppednum, newsliceofnums []float64

	if numtopop > len(sliceofnums) {
		return poppednum, sliceofnums, errorStore.NumToPopMErr // NumToPopMErr
	}
	if numtopop <= 0 {
		return poppednum, sliceofnums, errorStore.NumToPopMErr // NumToPopZeroErr
	}

	poppednum = append(sliceofnums[len(sliceofnums)-numtopop:])
	newsliceofnums = append(sliceofnums[:len(sliceofnums)-numtopop], sliceofnums[len(sliceofnums):]...)

	return poppednum, newsliceofnums, nil
}

func popOp(opslice []int) (int, []int, error) {

	var newopslice []int

	if len(opslice) == 0 {
		return 0, opslice, errorStore.NthToPopErr // NthToPopErr
	}

	poppedop := opslice[len(opslice)-1]
	newopslice = append(opslice[:len(opslice)-1], opslice[len(opslice):]...)

	return poppedop, newopslice, nil
}

/*
	func (s *TCalc) GetLightExpression(taskStore []Task, exprID string, sliceofnums []float64, opslice []int, prioritynum int, operator int, addop bool) (Task, []float64, []int, error) {
		var (
			poppedop            int
			poppednums          []float64
			popnumerr, popoperr error
			lightexpr           Task
		)
		//log.Println("nums", sliceofnums)
		//log.Println("operators", opslice)
		log.Println("tasks", taskStore)

		switch {
		case len(taskStore) == 0:

			poppedop, opslice, popoperr = popOp(opslice)
			poppednums, sliceofnums, popnumerr = popNum(sliceofnums, 2)

			log.Println(3.3, opslice, poppedop)

			if poppedop == 0 && popoperr != nil {
				return Task{}, sliceofnums, opslice, popoperr

			}
			if poppednums == nil && popnumerr != nil {
				return Task{}, sliceofnums, opslice, popnumerr
			}

			ID := strconv.Itoa(prioritynum) // prioritynum * -1 ???

			switch {
			case poppedop == isAddition:

				lightexpr = Task{
					ID:             ID,
					ExprID:         exprID,
					Arg1:           poppednums[0],
					Arg2:           poppednums[1],
					Operation:      "+",
					Operation_time: ConfigFromEnv().TimeAddition,
				}

			case poppedop == isSubtraction:
				lightexpr = Task{
					ID:             ID,
					ExprID:         exprID,
					Arg1:           poppednums[0],
					Arg2:           poppednums[1],
					Operation:      "-",
					Operation_time: ConfigFromEnv().TimeSubtraction,
				}
			case poppedop == isMultiplication:
				lightexpr = Task{
					ID:             ID,
					ExprID:         exprID,
					Arg1:           poppednums[0],
					Arg2:           poppednums[1],
					Operation:      "*",
					Operation_time: ConfigFromEnv().TimeMultiplications,
				}

			case poppedop == isDivision:
				if poppednums[1] == 0 {
					return Task{}, sliceofnums, opslice, errorStore.DvsByZeroErr //DvsByZeroErr
				}
				lightexpr = Task{
					ID:             ID,
					ExprID:         exprID,
					Arg1:           poppednums[0],
					Arg2:           poppednums[1],
					Operation:      "/",
					Operation_time: ConfigFromEnv().TimeDivisions,
				}

				//sliceofnums = append(sliceofnums, resultOfTask)
			}

			// Operator --> {  0 0 0}  {  20 20 0} вместо нуля нужен оператор
			if addop {
				opslice = append(opslice, operator)
			}
		default:
			if taskStore[prioritynum-1].Result == "" {
				ID := strconv.Itoa(prioritynum)
				arg1ID := strconv.Itoa(prioritynum - 1)

				poppedop, opslice, popoperr = popOp(opslice)
				poppednums, sliceofnums, popnumerr = popNum(sliceofnums, 1)

				if poppedop == 0 && popoperr != nil {
					return Task{}, sliceofnums, opslice, popoperr
				}
				if poppednums == nil && popnumerr != nil {
					return Task{}, sliceofnums, opslice, popnumerr
				}

				switch {
				case poppedop == isAddition:

					lightexpr = Task{
						ID:             ID,
						ExprID:         exprID,
						IDArg1:         arg1ID,
						Arg2:           poppednums[0],
						Operation:      "+",
						Operation_time: ConfigFromEnv().TimeAddition,
					}

				case poppedop == isSubtraction:
					lightexpr = Task{
						ID:             ID,
						ExprID:         exprID,
						IDArg1:         arg1ID,
						Arg2:           poppednums[0],
						Operation:      "-",
						Operation_time: ConfigFromEnv().TimeSubtraction,
					}
				case poppedop == isMultiplication:
					lightexpr = Task{
						ID:             ID,
						ExprID:         exprID,
						IDArg1:         arg1ID,
						Arg2:           poppednums[0],
						Operation:      "*",
						Operation_time: ConfigFromEnv().TimeMultiplications,
					}

				case poppedop == isDivision:
					if poppednums[0] == 0 {
						return Task{}, sliceofnums, opslice, errorStore.DvsByZeroErr //DvsByZeroErr
					}
					lightexpr = Task{
						ID:             ID,
						ExprID:         exprID,
						IDArg1:         arg1ID,
						Arg2:           poppednums[0],
						Operation:      "/",
						Operation_time: ConfigFromEnv().TimeDivisions,
					}
					return lightexpr, sliceofnums, opslice, nil
				}
			}
			if addop {
				opslice = append(opslice, operator)
			}
		}

		/*switch {
		case len(sliceofnums) == 1 && len(total) > 0:
			if total[prioritynum-1].Result == "" {
				ID := strconv.Itoa(prioritynum)
				arg1ID := strconv.Itoa(prioritynum - 1)

				poppedop, opslice, popoperr = popOp(opslice)
				poppednums, sliceofnums, popnumerr = popNum(sliceofnums, 1)

				if poppedop == 0 && popoperr != nil {
					return Task{}, sliceofnums, opslice, popoperr
				}
				if poppednums == nil && popnumerr != nil {
					return Task{}, sliceofnums, opslice, popnumerr
				}

				switch {
				case poppedop == isAddition:

					lightexpr = Task{
						ID:             ID,
						ExprID:         exprID,
						IDArg1:         arg1ID,
						Arg2:           poppednums[0],
						Operation:      "+",
						Operation_time: ConfigFromEnv().TimeAddition,
					}

				case poppedop == isSubtraction:
					lightexpr = Task{
						ID:             ID,
						ExprID:         exprID,
						IDArg1:         arg1ID,
						Arg2:           poppednums[0],
						Operation:      "-",
						Operation_time: ConfigFromEnv().TimeSubtraction,
					}
				case poppedop == isMultiplication:
					lightexpr = Task{
						ID:             ID,
						ExprID:         exprID,
						IDArg1:         arg1ID,
						Arg2:           poppednums[0],
						Operation:      "*",
						Operation_time: ConfigFromEnv().TimeMultiplications,
					}

				case poppedop == isDivision:
					if poppednums[0] == 0 {
						return Task{}, sliceofnums, opslice, errorStore.DvsByZeroErr //DvsByZeroErr
					}
					lightexpr = Task{
						ID:             ID,
						ExprID:         exprID,
						IDArg1:         arg1ID,
						Arg2:           poppednums[0],
						Operation:      "/",
						Operation_time: ConfigFromEnv().TimeDivisions,
					}
					return lightexpr, sliceofnums, opslice, nil
				}
			}
			if addop {
				opslice = append(opslice, operator)
			}

		case len(sliceofnums) == 0 && len(total) > 1:
			if total[prioritynum-1].Result == "" {

				ID := strconv.Itoa(prioritynum)
				arg1ID := strconv.Itoa(prioritynum - 1)
				arg2ID := strconv.Itoa(prioritynum - 2)

				poppedop, opslice, popoperr = popOp(opslice)

				if poppedop == 0 && popoperr != nil {
					return Task{}, sliceofnums, opslice, popoperr
				}

				switch {
				case poppedop == isAddition:

					lightexpr = Task{
						ID:             ID,
						ExprID:         exprID,
						IDArg1:         arg1ID,
						IDArg2:         arg2ID,
						Operation:      "+",
						Operation_time: ConfigFromEnv().TimeAddition,
					}

				case poppedop == isSubtraction:
					lightexpr = Task{
						ID:             ID,
						ExprID:         exprID,
						IDArg1:         arg1ID,
						IDArg2:         arg2ID,
						Operation:      "-",
						Operation_time: ConfigFromEnv().TimeSubtraction,
					}
				case poppedop == isMultiplication:
					lightexpr = Task{
						ID:             ID,
						ExprID:         exprID,
						IDArg1:         arg1ID,
						IDArg2:         arg2ID,
						Operation:      "*",
						Operation_time: ConfigFromEnv().TimeMultiplications,
					}

				case poppedop == isDivision:
					if sliceofnums[0] == 0 {
						return Task{}, sliceofnums, opslice, errorStore.DvsByZeroErr //DvsByZeroErr
					}
					lightexpr = Task{
						ID:             ID,
						ExprID:         exprID,
						IDArg1:         arg1ID,
						IDArg2:         arg2ID,
						Operation:      "/",
						Operation_time: ConfigFromEnv().TimeDivisions,
					}
					return lightexpr, sliceofnums, opslice, nil
				}

			}
			if addop {
				opslice = append(opslice, operator)
			}
		default:

			log.Println(3, "operators", opslice)

			poppedop, opslice, popoperr = popOp(opslice)
			poppednums, sliceofnums, popnumerr = popNum(sliceofnums, 2)

			log.Println(3.3, opslice, poppedop)


			if poppedop == 0 && popoperr != nil {
				return Task{}, sliceofnums, opslice, popoperr

			}
			if poppednums == nil && popnumerr != nil {
				return Task{}, sliceofnums, opslice, popnumerr
			}

			ID := strconv.Itoa(prioritynum) // prioritynum * -1 ???

			switch {
			case poppedop == isAddition:

				lightexpr = Task{
					ID:             ID,
					ExprID:         exprID,
					Arg1:           poppednums[0],
					Arg2:           poppednums[1],
					Operation:      "+",
					Operation_time: ConfigFromEnv().TimeAddition,
				}

			case poppedop == isSubtraction:
				lightexpr = Task{
					ID:             ID,
					ExprID:         exprID,
					Arg1:           poppednums[0],
					Arg2:           poppednums[1],
					Operation:      "-",
					Operation_time: ConfigFromEnv().TimeSubtraction,
				}
			case poppedop == isMultiplication:
				lightexpr = Task{
					ID:             ID,
					ExprID:         exprID,
					Arg1:           poppednums[0],
					Arg2:           poppednums[1],
					Operation:      "*",
					Operation_time: ConfigFromEnv().TimeMultiplications,
				}

			case poppedop == isDivision:
				if poppednums[1] == 0 {
					return Task{}, sliceofnums, opslice, errorStore.DvsByZeroErr //DvsByZeroErr
				}
				lightexpr = Task{
					ID:             ID,
					ExprID:         exprID,
					Arg1:           poppednums[0],
					Arg2:           poppednums[1],
					Operation:      "/",
					Operation_time: ConfigFromEnv().TimeDivisions,
				}

				//sliceofnums = append(sliceofnums, resultOfTask)
			}

			// Operator --> {  0 0 0}  {  20 20 0} вместо нуля нужен оператор
			if addop {
				opslice = append(opslice, operator)
			}

		}

		return lightexpr, sliceofnums, opslice, nil
	}
*/
func transact(sliceofnums []float64, opslice []int, operator int, addop bool) (float64, []float64, []int, error) {
	var result float64
	var poppedop int
	var poppednums []float64
	var popnumerr, popoperr error

	poppedop, opslice, popoperr = popOp(opslice)
	//poppednums, sliceofnums, popnumerr = popNum(sliceofnums, 2)

	if poppedop == 0 && popoperr != nil {
		return 0, sliceofnums, opslice, popoperr
	}

	if poppednums == nil && popnumerr != nil {
		return 0, sliceofnums, opslice, popnumerr
	}

	switch {
	case poppedop == isAddition:
		result = poppednums[0] + poppednums[1]
	case poppedop == isSubtraction:
		result = poppednums[0] - poppednums[1]
	case poppedop == isMultiplication:
		result = poppednums[0] * poppednums[1]
	case poppedop == isDivision:
		if poppednums[1] == 0 {
			return 0, sliceofnums, opslice, errorStore.DvsByZeroErr //DvsByZeroErr
		}
		result = poppednums[0] / poppednums[1]
	}

	sliceofnums = append(sliceofnums, result)
	if addop {
		opslice = append(opslice, operator)
	}

	return result, sliceofnums, opslice, nil
}

func (s *TCalc) IsCorrectExpression(Expression string) (bool, error) { //Проверка на правильность заданной строки
	var d DCalc
	var errorstring string

	if Expression == "" {
		return false, errorStore.EmptyExpressionErr //Пустое выражение
	}

	if errorStore.IncorrectExpressionErr.Error() != "incorrect expression" {
		errorStore.IncorrectExpressionErr = fmt.Errorf(`incorrect expression`)
	}

	Expression = strings.ReplaceAll(Expression, " ", "")

	correctexpression := true
	expressionlength := len(Expression)
	countleftparenthesis := 0
	countrightparenthesis := 0
	for index, _ := range Expression {
		if index < expressionlength-1 {
			switch {
			case !d.IsNumber(Expression[index]) && d.IsParenthesis(Expression[index]) == 0 && d.IsOperator(Expression[index]) == 0 && d.IsSeparator(Expression[index]) == 0: //Недопустимые символы
				correctexpression = false
				errorstring += fmt.Sprintf("| incorrect symbol, char %d. Allowed only: %s ", index, "1234567890.*/+-()")
			case index == 0 && !d.IsNumber(Expression[index]) && d.IsParenthesis(Expression[index]) == 0 && d.IsOperator(Expression[index]) != isSubtraction: //Запрещенная последовательность "выражение начинается не числом и не скобкой"
				correctexpression = false
				errorstring += fmt.Sprintf("| wrong sequence `non-number character`: char %d ", index)
			case d.IsOperator(Expression[index]) != 0 && d.IsOperator(Expression[index+1]) != 0: //Запрещенная последовательность "оператор->оператор"
				correctexpression = false
				errorstring += fmt.Sprintf("| wrong sequence `operation sign->operation sign`: chars %d, %d ", index, index+1)
			case d.IsSeparator(Expression[index]) != 0 && d.IsSeparator(Expression[index+1]) != 0: //Запрещенная последовательность "разделитель->разделитель"
				correctexpression = false
				errorstring += fmt.Sprintf("| wrong sequence `multiple separators together`: starting from char %d ", index)
			case d.IsParenthesis(Expression[index]) != 0 && d.IsSeparator(Expression[index+1]) != 0: //Запрещенная последовательность "скобка->разделитель дроби"
				correctexpression = false
				errorstring += fmt.Sprintf("| wrong sequence `parenthesis->separator`: chars %d, %d ", index, index+1)
			case d.IsParenthesis(Expression[index+1]) != 0 && d.IsSeparator(Expression[index]) != 0: //Запрещенная последовательность "разделитель дроби->скобка"
				correctexpression = false
				errorstring += fmt.Sprintf("| wrong sequence `separator->parenthesis`: chars %d, %d ", index, index+1)
			case d.IsSeparator(Expression[index]) != 0 && d.IsOperator(Expression[index+1]) != 0: //Запрещенная последовательность "разделитель дроби->оператор
				correctexpression = false
				errorstring += fmt.Sprintf("| wrong sequence `separator->operation sign`: chars %d, %d ", index, index+1)
			case d.IsSeparator(Expression[index+1]) != 0 && d.IsOperator(Expression[index]) != 0: //Запрещенная последовательность "оператор->разделитель дроби"
				correctexpression = false
				errorstring += fmt.Sprintf("| wrong sequence `operation sign->separator`: chars %d, %d ", index, index+1)
			case d.IsParenthesis(Expression[index]) == isRightParenthesis && d.IsOperator(Expression[index+1]) == 0 && d.IsParenthesis(Expression[index+1]) != isRightParenthesis:
				correctexpression = false
				errorstring += fmt.Sprintf("| wrong sequence `right parenthesys -> non operation sign or non right parenthesys character`: chars %d, %d ", index, index+1)
			case Expression[index] == '/' && Expression[index+1] == '0': // Запрещенная последовательность "Деление на ноль"
				correctexpression = false
				errorstring += fmt.Sprintf("| wrong sequence `division by zero`")
			case d.IsSeparator(Expression[index]) != 0 && d.IsNumber(Expression[index+1]) && d.IsNumber(Expression[index-1]): //Запрещенная последовательность "множественные разделители дроби в числе"
				for nextcharindex := index + 1; nextcharindex < expressionlength; nextcharindex++ {
					if !d.IsNumber(Expression[nextcharindex]) {
						if d.IsSeparator(Expression[nextcharindex]) != 0 {
							correctexpression = false
							errorstring += fmt.Sprintf("| wrong sequence `multiple separators within number`: starting from char %d ", index)
							break
						} else {
							break
						}
					}
				}
			case d.IsParenthesis(Expression[index]) == isLeftParenthesis && d.IsParenthesis(Expression[index+1]) == isRightParenthesis: //Запрещенная последовательность "пустые скобки"
				correctexpression = false
				errorstring += fmt.Sprintf("| wrong sequence `empty parentheses`: chars %d, %d ", index, index+1)
			case d.IsParenthesis(Expression[index]) == isRightParenthesis && countleftparenthesis == 0: // Запрещенная последовательность "подвыражение начинается с правой скобки"
				countrightparenthesis++
				correctexpression = false
				errorstring += fmt.Sprintf("| wrong sequence `beginning form right parenthesis`: on char %d ", index)
			case d.IsParenthesis(Expression[index]) == isLeftParenthesis && countleftparenthesis == 0: // Считаем левые и правые скобки
				countleftparenthesis++
				for nextcharindex := index + 1; nextcharindex < expressionlength; nextcharindex++ {
					if d.IsParenthesis(Expression[nextcharindex]) == isLeftParenthesis {
						countleftparenthesis++
					} else if d.IsParenthesis(Expression[nextcharindex]) == isRightParenthesis {
						countrightparenthesis++
					}

				}
			}
		} else if !d.IsNumber(Expression[index]) && d.IsParenthesis(Expression[index]) == 0 && d.IsOperator(Expression[index]) == 0 && d.IsSeparator(Expression[index]) == 0 { //Недопустимые символы
			correctexpression = false
			errorstring += fmt.Sprintf("| incorrect symbol, char %d. Allowed only: %s", index, "1234567890.*/+-()")
		} else if !d.IsNumber(Expression[index]) && d.IsParenthesis(Expression[index]) != isRightParenthesis && index == expressionlength-1 {
			correctexpression = false
			errorstring += "| wrong sequence `non-numeric last character`"
		} else if !d.IsNumber(Expression[index]) && d.IsParenthesis(Expression[index]) == isRightParenthesis && index == expressionlength-1 && countleftparenthesis != countrightparenthesis {
			correctexpression = false
			errorstring += "| wrong sequence `non-numeric last character`"
		}
	}

	if countleftparenthesis < countrightparenthesis { // Не хватает левых скобок
		correctexpression = false
		errorstring += "| wrong sequence `insufficient number of left parentheses`"
	} else if countleftparenthesis > countrightparenthesis { // Не хватает правых скобок
		correctexpression = false
		errorstring += "| wrong sequence `insufficient number of right parentheses`"
	}

	if !correctexpression { //Некорректное выражение
		errorStore.IncorrectExpressionErr = fmt.Errorf("%s %s", errorStore.IncorrectExpressionErr, errorstring)
		return false, errorStore.IncorrectExpressionErr
	}
	return true, nil
}

/*
func tokenizeandCalc(Expression string) (float64, error) {
	var s TCalc
	var d DCalc
	var result float64
	var operatorsslice []int
	var numsslice []float64
	var priority, countdown int
	var matherr, numconverr error

	Expression = strings.ReplaceAll(Expression, " ", "")

	check, checkerr := s.IsCorrectExpression(Expression)
	if !check && checkerr != nil {
		return 0, checkerr
	}
	length := len(Expression)
	for indexoftokenizer := 0; indexoftokenizer < length; indexoftokenizer++ {
		operatorslicelength := len(operatorsslice)
		if d.IsNumber(Expression[indexoftokenizer]) {
			numsslice, indexoftokenizer, numconverr = extractNum(Expression, indexoftokenizer, numsslice, false) //Положительное число
		} else if !d.IsNumber(Expression[indexoftokenizer]) && d.IsOperator(Expression[indexoftokenizer]) == isSubtraction && d.IsNumber(Expression[indexoftokenizer+1]) && indexoftokenizer == 0 { // Отрицательное число в начале выражения
			numsslice, indexoftokenizer, numconverr = extractNum(Expression, indexoftokenizer+1, numsslice, true)
		} else if d.IsParenthesis(Expression[indexoftokenizer]) == isLeftParenthesis && d.IsOperator(Expression[indexoftokenizer+1]) == isSubtraction && d.IsNumber(Expression[indexoftokenizer+2]) { // Отрицательное число после открывающей скобки
			numsslice, indexoftokenizer, numconverr = extractNum(Expression, indexoftokenizer+2, numsslice, true)
			if d.IsNumber(Expression[indexoftokenizer-1]) { // Добавляем в стек операторов открывающую скобку если она не часть выражения вида (-1), описывающего отрицательное число
				operatorsslice = append(operatorsslice, 1)
				operatorslicelength++
			}
			if indexoftokenizer == length { // Конец строки после закрывающей скобкой, перед которой отрицательное число
				break
			}
		}
		if numsslice == nil && numconverr != nil {
			return 0, numconverr
		}
		if !d.IsNumber(Expression[indexoftokenizer]) && d.IsSeparator(Expression[indexoftokenizer]) == 0 {
			switch {
			case d.IsOperator(Expression[indexoftokenizer]) != 0:
				if operatorslicelength-1 >= 0 {
					priority = getPryority(d.IsOperator(Expression[indexoftokenizer]))
					if getPryority(operatorsslice[operatorslicelength-1]) == priority {
						result, numsslice, operatorsslice, matherr = transact(numsslice, operatorsslice, d.IsOperator(Expression[indexoftokenizer]), true)
						if result == 0 && matherr != nil {
							return 0, matherr

						}
					} else if getPryority(operatorsslice[operatorslicelength-1]) < priority {
						operatorsslice = append(operatorsslice, d.IsOperator(Expression[indexoftokenizer]))
					} else if getPryority(operatorsslice[operatorslicelength-1]) > priority {
						result, numsslice, operatorsslice, matherr = transact(numsslice, operatorsslice, d.IsOperator(Expression[indexoftokenizer]), true)
						if result == 0 && matherr != nil {
							return 0, matherr

						}
					}

				} else {
					operatorsslice = append(operatorsslice, d.IsOperator(Expression[indexoftokenizer]))
				}
			case d.IsParenthesis(Expression[indexoftokenizer]) == isLeftParenthesis:
				operatorsslice = append(operatorsslice, isLeftParenthesis)
			case d.IsParenthesis(Expression[indexoftokenizer]) == isRightParenthesis:
				for {
					if (operatorsslice[len(operatorsslice)-1]) == isLeftParenthesis {
						_, operatorsslice, _ = popOp(operatorsslice)
						break
					}
					result, numsslice, operatorsslice, matherr = transact(numsslice, operatorsslice, 0, false)
					if result == 0 && matherr != nil {
						return 0, matherr

					}
				}
			}
		}
	}

	countdown = len(operatorsslice) - 1
	for {
		if countdown < 0 {
			break
		} else {
			result, numsslice, operatorsslice, matherr = transact(numsslice, operatorsslice, 0, false)
			if result == 0 && matherr != nil {
				return 0, matherr

			}
		}
		countdown--
	}
	return numsslice[0], nil
}
*/

/*
func (s *TCalc) ExprtolightExprs(Expression, exprID string) ([]Task, error) { // Функция разбивает выражение на подвыражения
	var d DCalc
	var exprpryority int
	var result Task
	var operatorsslice []int
	var numsslice []float64
	var priority, countdown int
	var matherr, numconverr error

	Expression = strings.ReplaceAll(Expression, " ", "")

	check, checkerr := s.IsCorrectExpression(Expression)
	if !check && checkerr != nil {
		return nil, checkerr
	}
	length := len(Expression)
	for indexoftokenizer := 0; indexoftokenizer < length; indexoftokenizer++ {
		operatorslicelength := len(operatorsslice)
		if d.IsNumber(Expression[indexoftokenizer]) {
			numsslice, indexoftokenizer, numconverr = extractNum(Expression, indexoftokenizer, numsslice, false) //Положительное число
		} else if !d.IsNumber(Expression[indexoftokenizer]) && d.IsOperator(Expression[indexoftokenizer]) == isSubtraction && d.IsNumber(Expression[indexoftokenizer+1]) && indexoftokenizer == 0 { // Отрицательное число в начале выражения
			numsslice, indexoftokenizer, numconverr = extractNum(Expression, indexoftokenizer+1, numsslice, true)
		} else if d.IsParenthesis(Expression[indexoftokenizer]) == isLeftParenthesis && d.IsOperator(Expression[indexoftokenizer+1]) == isSubtraction && d.IsNumber(Expression[indexoftokenizer+2]) { // Отрицательное число после открывающей скобки
			numsslice, indexoftokenizer, numconverr = extractNum(Expression, indexoftokenizer+2, numsslice, true)
			if d.IsNumber(Expression[indexoftokenizer-1]) { // Добавляем в стек операторов открывающую скобку если она не часть выражения вида (-1), описывающего отрицательное число
				operatorsslice = append(operatorsslice, 1)
				operatorslicelength++
			}
			if indexoftokenizer == length { // Конец строки после закрывающей скобкой, перед которой отрицательное число
				break
			}
		}
		if numsslice == nil && numconverr != nil {
			return nil, numconverr
		}
		if !d.IsNumber(Expression[indexoftokenizer]) && d.IsSeparator(Expression[indexoftokenizer]) == 0 {
			switch {
			case d.IsOperator(Expression[indexoftokenizer]) != 0:
				if operatorslicelength-1 >= 0 {
					priority = getPryority(d.IsOperator(Expression[indexoftokenizer]))
					if getPryority(operatorsslice[operatorslicelength-1]) == priority {
						result, numsslice, operatorsslice, matherr = s.GetLightExpression(taskStore, exprID, numsslice, operatorsslice, exprpryority, d.IsOperator(Expression[indexoftokenizer]), true)
						taskStore = append(taskStore, result)
						for taskStore[exprpryority].Result == "" {
						}
						res, cerr := strconv.ParseFloat(taskStore[exprpryority].Result, 32)
						if cerr != nil {
							log.Printf("Error of type conversion the result of a task %v", cerr)
						}
						numsslice = append(numsslice, res)
						exprpryority++
						if matherr != nil {
							return nil, matherr

						}
					} else if getPryority(operatorsslice[operatorslicelength-1]) < priority {
						operatorsslice = append(operatorsslice, d.IsOperator(Expression[indexoftokenizer]))
					} else if getPryority(operatorsslice[operatorslicelength-1]) > priority {
						result, numsslice, operatorsslice, matherr = s.GetLightExpression(taskStore, exprID, numsslice, operatorsslice, exprpryority, d.IsOperator(Expression[indexoftokenizer]), true)
						taskStore = append(taskStore, result)
						for taskStore[exprpryority].Result == "" {
						}
						res, cerr := strconv.ParseFloat(taskStore[exprpryority].Result, 32)
						if cerr != nil {
							log.Printf("Error of type conversion the result of a task %v", cerr)
						}
						numsslice = append(numsslice, res)
						exprpryority++
						if matherr != nil {
							return nil, matherr

						}
					}

				} else {
					operatorsslice = append(operatorsslice, d.IsOperator(Expression[indexoftokenizer]))
				}
			case d.IsParenthesis(Expression[indexoftokenizer]) == isLeftParenthesis:
				operatorsslice = append(operatorsslice, isLeftParenthesis)
			case d.IsParenthesis(Expression[indexoftokenizer]) == isRightParenthesis:
				for {
					if (operatorsslice[len(operatorsslice)-1]) == isLeftParenthesis {
						_, operatorsslice, _ = popOp(operatorsslice)
						break
					}
					result, numsslice, operatorsslice, matherr = s.GetLightExpression(taskStore, exprID, numsslice, operatorsslice, exprpryority, 0, false)
					taskStore = append(taskStore, result)
					for taskStore[exprpryority].Result == "" {
					}
					res, cerr := strconv.ParseFloat(taskStore[exprpryority].Result, 32)
					if cerr != nil {
						log.Printf("Error of type conversion the result of a task %v", cerr)
					}
					numsslice = append(numsslice, res)
					exprpryority++
					if matherr != nil {
						return nil, matherr

					}
				}
			}
		}
	}

	countdown = len(operatorsslice) - 1
	for {
		if countdown < 0 {
			break
		} else {
			result, numsslice, operatorsslice, matherr = s.GetLightExpression(taskStore, exprID, numsslice, operatorsslice, exprpryority, 0, false)
			taskStore = append(taskStore, result)
			for taskStore[exprpryority].Result == "" {
			}
			res, cerr := strconv.ParseFloat(taskStore[exprpryority].Result, 32)
			if cerr != nil {
				log.Printf("Error of type conversion the result of a task %v", cerr)
			}
			numsslice = append(numsslice, res)
			exprpryority++
			if matherr != nil {
				return nil, matherr

			}
		}
		countdown--
	}
	return taskStore, nil
}
*/

func (s TCalc) Init() TCalc {
	s.history = make(map[time.Time]map[string]string)
	return s
}

func (s TCalc) RemoveHistory() {
	for t := range s.history {
		delete(s.history, t)
	}

}

func (s TCalc) GetCalcHistory() map[time.Time]map[string]string {

	return s.history
}

/*
func (s TCalc) Calc(Expression string) (float64, error) {

	resultmap := make(map[string]string)

	if s.history == nil {
		fmt.Println(" ")
		s.history = make(map[time.Time]map[string]string)
	}

	result, calcerr := tokenizeandCalc(Expression)
	if result == 0 && calcerr != nil {
		resultmap[Expression] = calcerr.Error()
		s.history[time.Now()] = resultmap
		return 0, calcerr
	} else {
		resultmap[Expression] = strconv.FormatFloat(result, 'g', 8, 64)
		s.history[time.Now()] = resultmap
		return result, nil
	}

}
*/
