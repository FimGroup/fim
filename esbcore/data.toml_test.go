package esbcore

import (
	"encoding/json"
	"testing"

	"github.com/pelletier/go-toml/v2"
)

type A struct {
	String string
	Int    int
	Name   string
	Addr   struct {
		Country string
		City    string
		Address string
		Name    string
	}
	Email   string
	Company struct {
		Name        string
		Address     string
		Description string
		CEO         struct {
			Name   string
			Salary int
		}
		CFO struct {
			Name    string
			Salary  int
			Partial bool
		}
	}
}

var a = A{
	String: "hello world",
	Int:    1000,
	Name:   "哈哈",
	Addr: struct {
		Country string
		City    string
		Address string
		Name    string
	}{Country: "AAA", City: "sldfjlsjf", Address: "skljvbowjboi sofjowejgvo woefjwoegvj sfvweg", Name: "哈哈 是东方购物逛街哦"},
	Email: "slfdjkowbj@sljvlwejbo.com",
	Company: struct {
		Name        string
		Address     string
		Description string
		CEO         struct {
			Name   string
			Salary int
		}
		CFO struct {
			Name    string
			Salary  int
			Partial bool
		}
	}{
		Name:        "lsjkdfo sdjfosjf",
		Address:     "slfjowvjgvjow wovjwoejvb wovjweovj",
		Description: "slkjvlwejkjvwovjweojgvojbowjoebjwobjwjbwoebjiowbiwofiwoejgojeg",
		CEO: struct {
			Name   string
			Salary int
		}{Name: "呵呵", Salary: 10000000},
		CFO: struct {
			Name    string
			Salary  int
			Partial bool
		}{Name: "嘿嘿", Salary: 1000000, Partial: true},
	},
}

func init() {
	s := "a"
	for i := 0; i < 1024; i++ {
		s += "bcdefghibcdefghibcdefghibcdefghibcdefghibcdefghibcdefghibcdefghi"
	}
	a.String = s
}

func TestExample(t *testing.T) {
	checkPrint("alfjdslfj", t)
	checkPrint("_alfjdslfj", t)
	checkPrint("王王王王", t)
	checkPrint(".alfjdslfj", t)
	checkPrint("alfjd/slfj", t)
	checkPrint("alfjdslfj[]", t)
	checkPrint("alfjdslfj[1]", t)
	checkPrint("alfjdslfj[111111111]", t)
	checkPrint("alfjdslfj[", t)
	checkPrint("alfjdslfj]", t)
	checkPrint("111alfjdslfj", t)
	checkPrint("alfjd1111slfj", t)
}

func TestExample2(t *testing.T) {
	checkPrint2("alfjdslfj", t)
	checkPrint2("_alfjdslfj", t)
	checkPrint2("王王王王", t)
	checkPrint2(".alfjdslfj", t)
	checkPrint2("alfjd/slfj", t)
	checkPrint2("alfjdslfj[]", t)
	checkPrint2("alfjdslfj[1]", t)
	checkPrint2("alfjdslfj[111111111]", t)
	checkPrint2("alfjdslfj[", t)
	checkPrint2("alfjdslfj]", t)
	checkPrint2("111alfjdslfj", t)
	checkPrint2("alfjd1111slfj", t)
}

func checkPrint(str string, t *testing.T) {
	typeOfElem, ok := checkElementKey(str, false)
	t.Log(typeOfElem, ok, str)
}

func checkPrint2(str string, t *testing.T) {
	typeOfElem, ok := checkElementKey(str, true)
	t.Log(typeOfElem, ok, str)
}

func BenchmarkTomlMarshal(b *testing.B) {
	data, err := toml.Marshal(a)
	if err != nil {
		b.Fatal(err)
	}

	_ = data
	//b.Log(string(data))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		toml.Marshal(a)
	}
}

func BenchmarkTomlUnmarshal(b *testing.B) {
	data, err := toml.Marshal(a)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var a A
		toml.Unmarshal(data, &a)
	}
}

func BenchmarkJsonMarshall(b *testing.B) {
	data, err := json.Marshal(a)
	if err != nil {
		b.Fatal(err)
	}

	_ = data
	//b.Log(string(data))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(a)
	}
}

func BenchmarkJsonUnmarshal(b *testing.B) {
	data, err := json.Marshal(a)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var a A
		json.Unmarshal(data, &a)
	}
}
