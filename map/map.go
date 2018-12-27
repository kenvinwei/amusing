package main

import "fmt"


//map[index] = value
func testMap1() {
	ret := make([]map[string]interface{}, 0)

	ret1 := make(map[string]interface{})

	ret1["aa"] = "bb"
	//ret = append(ret, ret1)
	ret[0] = ret1

	fmt.Println("ret ====", ret)
}

//map[index] = value
func testMap2() {
	ret := make([]map[string]interface{}, 1)

	ret1 := make(map[string]interface{})

	ret1["aa"] = "bb"
	//ret = append(ret, ret1)
	ret[0] = ret1

	fmt.Println("ret ====", ret)
}

//append
func testMap3() {
	ret := make([]map[string]interface{}, 0)

	ret1 := make(map[string]interface{})

	ret1["aa"] = "bb"
	ret = append(ret, ret1)

	fmt.Println("ret ====", ret)
}

//golang 中map直接对map[index]赋值，map不会自动扩容，如果容量不够会报(panic: runtime error: index out of range) 错误
//append 会自动扩容

func main() {
	testMap1() //failed
	testMap2() //succ
	testMap3() //succ
}
