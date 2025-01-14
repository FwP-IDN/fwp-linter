package somepackage

import "fmt"

// @fwp.linter
type MyStruct struct {
	MyField1    int
	MyField2    int
	MyField3    int
	UnusedField int
}

func (m *MyStruct) Init() {
	m.MyField1 = 0
	m.MyField2 = 0
	m.MyField3 = 0
}

func (m *MyStruct) PrintSum() {
	sumNum := 0
	sumNum += m.MyField1
	sumNum += m.MyField2
	sumNum += m.MyField3
	fmt.Println(sumNum)
}

/*Lah*/
func (m *MyStruct) DoUniqueThings() {
	m.MyField2 = m.MyField1
	m.MyField3 = m.MyField1
	m.MyField2 += m.MyField3
}
