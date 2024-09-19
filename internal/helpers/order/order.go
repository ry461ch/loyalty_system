package orderhelper

import "strconv"

func ValidateOrderId(orderId string) bool {
	_, err := strconv.Atoi(orderId)
	if err != nil {
		return false
	}

	sum := 0
	idx := 0
	if len(orderId)%2 == 1 {
		num, _ := strconv.Atoi(string(orderId[0]))
		sum += num
		idx = 1
	}

	for ; idx < len(orderId); idx += 2 {
		num, _ := strconv.Atoi(string(orderId[idx]))
		num *= 2
		if num > 9 {
			num -= 9
		}
		sum += num
		num, _ = strconv.Atoi(string(orderId[idx+1]))
		sum += num
	}

	return sum%10 == 0
}
