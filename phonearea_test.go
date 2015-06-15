package phonearea

import "testing"

var (
	testCases = [][]string{
		{"13761921111", "上海"},
		{"13564387521", "上海"},
		{"18616832345", "上海"},
	}
)

func TestQuery(t *testing.T) {
	err := Init("./phonenum.txt", "./data")
	if err != nil {
		t.Fatal("数据库初始化失败", err)
	}
	for _, cs := range testCases {
		area, err := Query(cs[0])
		if err != nil {
			t.Fatal(err)
			return
		}
		if area.City != cs[1] {
			t.Fatal("城市不对")
			return
		}
	}
}
