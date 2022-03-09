package tool

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
)

func PickDomainFromUrl(s string) (string, error) {
	if !strings.Contains(s, "http") {
		s = "https://" + s
	}

	u, err := url.Parse(s)
	if err != nil {
		return "", err
	}

	return u.Host, nil
}

func ToJson(val interface{}) string {
	bytes, _ := jsoniter.Marshal(val)
	return string(bytes)
}

func Interface2String(value interface{}) string {
	key := ""
	if value == nil {
		return key
	}

	switch v := value.(type) {
	case float64:
		key = strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		key = strconv.FormatFloat(float64(v), 'f', -1, 64)
	case int:
		key = strconv.Itoa(v)
	case uint:
		key = strconv.Itoa(int(v))
	case int8:
		key = strconv.Itoa(int(v))
	case uint8:
		key = strconv.Itoa(int(v))
	case int16:
		key = strconv.Itoa(int(v))
	case uint16:
		key = strconv.Itoa(int(v))
	case int32:
		key = strconv.Itoa(int(v))
	case uint32:
		key = strconv.Itoa(int(v))
	case int64:
		key = strconv.FormatInt(v, 10)
	case uint64:
		key = strconv.FormatUint(v, 10)
	case string:
		key = v
	case []byte:
		key = string(v)
	case json.Number:
		key = v.String()
	}

	return key
}

// StrAppend 字符串拼接
func StrAppend(str1 string, str2 ...string) string {
	var builder strings.Builder
	builder.WriteString(str1)
	for _, str := range str2 {
		builder.WriteString(str)
	}
	return builder.String()
}

func Bytes2str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func SubStr(s string, length int) string {
	var size, n int
	for i := 0; i < length && n < len(s); i++ {
		_, size = utf8.DecodeRuneInString(s[n:])
		n += size
	}

	return s[:n]
}
