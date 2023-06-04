package modelinst

import (
	"encoding/json"
	"testing"

	"github.com/pelletier/go-toml/v2"
)

const (
	content = `
    a = [
			["http", "", [
				["body[]", "posts[]", [
					["post_id", "post_id"],
					["parent_post_id", "parent_post_id"],
					["post_type", "post_type"],
					["author_id", "author_id"],
					["forum_id", "forum_id"],
					["title", "title"],
					["content", "content"],
				]],
			]]
    ]`
	benchContent = `
    a = [
			["http", "", [
				["body[]", "posts[]", [
					["post_id", "post_id"],
				]],
			]]
    ]`

	contentNamedTopLevelPrimitiveArray = `
    a = [
			["http[]", "http[]", [
			]]
    ]
`
)

func TestToConverter(t *testing.T) {

	a := struct {
		A MappingRuleRaw
	}{}

	if err := toml.Unmarshal([]byte(content), &a); err != nil {
		t.Fatal(err)
	}

	if c, err := a.A.ToConverter(); err != nil {
		t.Fatal(err)
	} else {
		data, err := json.MarshalIndent(c, "", "  ")
		if err != nil {
			panic(err)
		}
		t.Log(string(data))
	}
}

func TestConverterData0(t *testing.T) {
	a := struct {
		A MappingRuleRaw
	}{}

	if err := toml.Unmarshal([]byte(content), &a); err != nil {
		t.Fatal(err)
	}
	c, err := a.A.ToConverter()
	if err != nil {
		t.Fatal(err)
	}

	src := ModelInstHelper{}.NewInst()
	{
		httpObj, err := src.ensureSubObject("http")
		if err != nil {
			t.Fatal(err)
		}
		bodyArray, err := httpObj.ensureSubArrayWithObjectElem("body")
		if err != nil {
			t.Fatal(err)
		}
		bodyElemObj, err := bodyArray.ensureArrayElement()
		if err != nil {
			t.Fatal(err)
		}
		if err := bodyElemObj.putPrimitiveValue("post_id", mustConvertPrimitive(135)); err != nil {
			t.Fatal(err)
		}
		if err := bodyElemObj.putPrimitiveValue("parent_post_id", mustConvertPrimitive(135)); err != nil {
			t.Fatal(err)
		}
		if err := bodyElemObj.putPrimitiveValue("post_type", mustConvertPrimitive(1)); err != nil {
			t.Fatal(err)
		}
		if err := bodyElemObj.putPrimitiveValue("author_id", mustConvertPrimitive("ABCDEFG")); err != nil {
			t.Fatal(err)
		}
		if err := bodyElemObj.putPrimitiveValue("forum_id", mustConvertPrimitive("1000001")); err != nil {
			t.Fatal(err)
		}
		if err := bodyElemObj.putPrimitiveValue("title", mustConvertPrimitive("TITLE001000000000")); err != nil {
			t.Fatal(err)
		}
		if err := bodyElemObj.putPrimitiveValue("content", mustConvertPrimitive("789")); err != nil {
			t.Fatal(err)
		}
		if err := bodyElemObj.putPrimitiveValue("audit_time", mustConvertPrimitive(123123123)); err != nil {
			t.Fatal(err)
		}
	}
	dst := ModelInstHelper{}.NewInst()
	if err := c.Transfer(src, dst); err != nil {
		t.Fatal(err)
	}
	t.Log(dst)
}

func BenchmarkConverterData0(b *testing.B) {
	a := struct {
		A MappingRuleRaw
	}{}

	if err := toml.Unmarshal([]byte(content), &a); err != nil {
		b.Fatal(err)
	}
	c, err := a.A.ToConverter()
	if err != nil {
		b.Fatal(err)
	}

	src := ModelInstHelper{}.NewInst()
	{
		httpObj, err := src.ensureSubObject("http")
		if err != nil {
			b.Fatal(err)
		}
		bodyArray, err := httpObj.ensureSubArrayWithObjectElem("body")
		if err != nil {
			b.Fatal(err)
		}
		bodyElemObj, err := bodyArray.ensureArrayElement()
		if err != nil {
			b.Fatal(err)
		}
		if err := bodyElemObj.putPrimitiveValue("post_id", mustConvertPrimitive(135)); err != nil {
			b.Fatal(err)
		}
		if err := bodyElemObj.putPrimitiveValue("parent_post_id", mustConvertPrimitive(135)); err != nil {
			b.Fatal(err)
		}
		if err := bodyElemObj.putPrimitiveValue("post_type", mustConvertPrimitive(1)); err != nil {
			b.Fatal(err)
		}
		if err := bodyElemObj.putPrimitiveValue("author_id", mustConvertPrimitive("ABCDEFG")); err != nil {
			b.Fatal(err)
		}
		if err := bodyElemObj.putPrimitiveValue("forum_id", mustConvertPrimitive("1000001")); err != nil {
			b.Fatal(err)
		}
		if err := bodyElemObj.putPrimitiveValue("title", mustConvertPrimitive("TITLE001000000000")); err != nil {
			b.Fatal(err)
		}
		if err := bodyElemObj.putPrimitiveValue("content", mustConvertPrimitive("789")); err != nil {
			b.Fatal(err)
		}
		if err := bodyElemObj.putPrimitiveValue("audit_time", mustConvertPrimitive(123123123)); err != nil {
			b.Fatal(err)
		}
	}
	dst := ModelInstHelper{}.NewInst()
	if err := c.Transfer(src, dst); err != nil {
		b.Fatal(err)
	}
	b.Log(dst)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dst := ModelInstHelper{}.NewInst()
		if err := c.Transfer(src, dst); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkConverterData0WithNative(b *testing.B) {
	src := map[string]interface{}{}
	{
		bodyElem := map[string]interface{}{
			"post_id":    135,
			"author_id":  "ABCDEFG",
			"content":    "789",
			"audit_time": 123123123,
		}
		body := []interface{}{bodyElem}
		src["http"] = map[string]interface{}{
			"body": body,
		}
	}

	for i := 0; i < b.N; i++ {
		dst := map[string]interface{}{}
		{
			http := src["http"].(map[string]interface{})
			body := http["body"].([]interface{})
			dstBodyArr := make([]interface{}, len(body))
			for i, v := range body {
				srcMap := v.(map[string]interface{})
				elem := map[string]interface{}{
					"post_id":   srcMap["post_id"],
					"author_id": srcMap["author_id"],
					"content":   srcMap["content"],
				}
				dstBodyArr[i] = elem
			}
			dst["posts"] = dstBodyArr
		}
	}
}

func TestConverterData1(t *testing.T) {
	a := struct {
		A MappingRuleRaw
	}{}

	if err := toml.Unmarshal([]byte(content), &a); err != nil {
		t.Fatal(err)
	}
	c, err := a.A.ToConverter()
	if err != nil {
		t.Fatal(err)
	}

	m := make(map[string]interface{})
	src := ModelInstHelper{}.WrapMap(m)
	dst := ModelInstHelper{}.NewInst()
	if err := c.Transfer(src, dst); err != nil {
		t.Fatal(err)
	}
	t.Log(dst)
}
