package tools

import (
	"encoding/json"
	"fmt"
	"github.com/sari3l/requests"
	"github.com/sari3l/requests/ext"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
)

func ConvertGbkToUtf8(str string) string {
	reader := transform.NewReader(strings.NewReader(str), simplifiedchinese.GBK.NewDecoder())
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		fmt.Print(err)
		return ""
	}
	return string(data)
}

func ConvertUtf8ToGbk(str string) string {
	reader := transform.NewReader(strings.NewReader(str), simplifiedchinese.GBK.NewEncoder())
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		fmt.Print(err)
		return ""
	}
	return string(data)
}

func HookResponseGbkToUtf8(response any) (error, any) {
	resp := response.(requests.Response)
	respRef := reflect.ValueOf(&resp).Elem()
	respContent := respRef.FieldByName("Content")
	chineseContent := ConvertGbkToUtf8(respContent.String())
	respContent.SetString(chineseContent)
	return nil, resp
}

func HookResponseUtf8ToGbk(response any) (error, any) {
	resp := response.(requests.Response)
	respRef := reflect.ValueOf(&resp).Elem()
	respContent := respRef.FieldByName("Content")
	chineseContent := ConvertUtf8ToGbk(respContent.String())
	respContent.SetString(chineseContent)
	return nil, resp
}

// ConvertStructToJson 需要结构体字段Tag设置为 `json:"目标键名"` 或 `json:"目标键名,omitempty"`
// 如 Text string `json:"text,omitempty"`
func ConvertStructToJson(obj any) map[string]any {
	result := make(map[string]any, 0)
	tmpBytes, _ := json.Marshal(&obj)
	json.Unmarshal(tmpBytes, &result)
	return result
}

// ConvertStructToDict 需要结构体字段Tag设置为 `dict:"目标键名"` 或 `dict:"目标键名,omitempty"`
// 如 Text string `dict:"text,omitempty"`
func ConvertStructToDict(obj any) ext.Dict {
	dict := ext.Dict{}
	ref := reflect.ValueOf(obj)
	for i := 0; i < ref.NumField(); i++ {
		tagOpt := ref.Type().Field(i).Tag.Get("dict")
		if tagOpt == "" {
			continue
		}
		var tagName string
		var omitemptyOpt bool
		if strings.Contains(tagOpt, ",") {
			tagSlice := strings.Split(tagOpt, ",")
			tagName = tagSlice[0]
			omitemptyOpt = tagSlice[1] == "omitempty"
		} else {
			tagName = tagOpt
		}
		var key string
		if tagName == "" {
			key = ref.Type().Field(i).Name
		} else {
			key = tagName
		}
		var value string
		value = ConvertValueToString(ref.Field(i))
		if omitemptyOpt == true && value == "" {
			continue
		}
		dict[key] = value
	}
	return dict
}

func ConvertValueToString(obj reflect.Value) string {
	switch obj.Kind() {
	case reflect.String:
		return obj.String()
	case reflect.Bool:
		return strconv.FormatBool(obj.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(obj.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(obj.Uint(), 10)
	case reflect.Float64, reflect.Float32:
		return strconv.FormatFloat(obj.Float(), 'f', -1, 64)
	case reflect.Complex64, reflect.Complex128:
		return strconv.FormatComplex(obj.Complex(), 'f', 0, 128)
	case reflect.Slice, reflect.Array:
		tmp := make([]string, 0)
		for i := 0; i < obj.Len(); i++ {
			tmp = append(tmp, ConvertValueToString(obj.Index(i)))
		}
		return strings.Join(tmp, ",")
	case reflect.Ptr, reflect.UnsafePointer:
		if obj.IsNil() {
			return ""
		} else {
			return ConvertValueToString(obj.Elem())
		}
	default:
		return ""
	}
}
