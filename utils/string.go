package utils

import (
	"strings"
)

func SplitHBaseRegionStr(data string) []string {
	// Split the string just like: Namespace_n1_table_t1_region_r1_metric_m1
	// return: [n1 t1 r1 m1]

	var res []string
	flagList := []string{"Namespace_", "_table_", "_region_", "_metric_"}
	temp1 := strings.SplitN(data, flagList[0], 2)

	temp2 := strings.SplitN(temp1[1], flagList[1], 2)
	res = append(res, temp2[0])

	temp3 := strings.SplitN(temp2[1], flagList[2], 2)
	res = append(res, temp3[0])

	temp4 := strings.SplitN(temp3[1], flagList[3], 2)
	res = append(res, temp4[0])

	res = append(res, temp4[1])

	return res
}
