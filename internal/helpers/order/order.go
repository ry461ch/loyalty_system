package orderhelper

import "strconv"

func ValidateOrderID(orderID string) bool {
	_, err := strconv.Atoi(orderID)
	if err != nil {
		return false
	}

	sum := 0
	idx := 0
	if len(orderID)%2 == 1 {
		num, _ := strconv.Atoi(string(orderID[0]))
		sum += num
		idx = 1
	}

	for ; idx < len(orderID); idx += 2 {
		num, _ := strconv.Atoi(string(orderID[idx]))
		num *= 2
		if num > 9 {
			num -= 9
		}
		sum += num
		num, _ = strconv.Atoi(string(orderID[idx+1]))
		sum += num
	}

	return sum%10 == 0
}
