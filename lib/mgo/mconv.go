package mgo

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

func vint(val reflect.Value) int64 {
	switch val.Kind() {
	case reflect.Bool:
		if val.Bool() {
			return 1
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return int64(val.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return int64(val.Uint())
	case reflect.Float32, reflect.Float64:
		return int64(val.Float())
	case reflect.String:
		s := val.String()
		if s == "" || s == "None" {
			return 0
		}
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			Fatalf(err)
		}
		return int64(v)
	default:
		Fatalf("unsupported type %v", val.Kind())
	}
	return 0
}
func vfloat(val reflect.Value) float64 {
	switch val.Kind() {
	case reflect.Bool:
		if val.Bool() {
			return 1
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(val.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return float64(val.Uint())
	case reflect.Float32, reflect.Float64:
		return val.Float()
	case reflect.String:
		s := val.String()
		if s == "" || s == "None" {
			return 0.0
		}
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			Fatalf(err)
		}
		return v
	default:
		Fatalf("unsupported type %v", val.Kind())
	}
	return 0.0
}
// Int convert int/float/string to int
func Int(a interface{}) int64 {
	val := reflect.ValueOf(a)
	return vint(val)
}

// Float convert int/float/string to float
func Float(a interface{}) float64 {
	val := reflect.ValueOf(a)
	return vfloat(val)
}

// Floats convert []a to []a.name
func Floats(a interface{}, name string) []float64 {
	val := reflect.ValueOf(a)
	switch val.Kind() {
	case reflect.Slice:
		fs := []float64{}
		for i := 0; i < val.Len(); i++ {
			v := val.Index(i)
			if v.Kind() == reflect.Struct {
				v = v.FieldByName(name)
			} else {
				v = v.Elem().FieldByName(name)
			}
			fs = append(fs, vfloat(v))
		}
		return fs
	default:
		panic("invalid type")
	}
}

// Range convert range string to float value
func Range(s string) (float64, float64) {
	a := strings.Split(s, ",")
	if len(a) == 1 {
		return Float(a[0]), Float(a[0])
	}
	return Float(a[0]), Float(a[1])
}

// Time convert string/int to time
// string format 'YY-mm-dd HH:MM:SS.XXX'
// int format unix-timestamp
func Time(a interface{}) time.Time {
	val := reflect.ValueOf(a)
	switch val.Kind() {
	case reflect.String:
		s := val.String()
		if !strings.Contains(s, "-") {
			s += "-01-01"
		}
		if !strings.Contains(s, " ") {
			s += " 00:00:00"
		}
		if !strings.Contains(s, ".") {
			s += ".000"
		}

		t, err := time.Parse(TIME_FORMAT, s)
		if err != nil {
			Fatalf(err)
		}
		return t
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v := val.Int()
		return time.Unix(v, 0)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		v := val.Uint()
		return time.Unix(int64(v), 0).UTC()
	default:
		Fatalf("unsupported type %v", val.Kind())
	}
	return time.Unix(0, 0)
}

// Today returns date of today
func Today() time.Time {
	return Time(DateStr(time.Now()))
}

// YearStr returns time's year string
func YearStr(t time.Time) string {
	return t.Format("2006")
}

// DateStr returns time's date string
func DateStr(t time.Time) string {
	return t.Format("2006-01-02")
}

// TimeStr returns time's date-time string
func TimeStr(t time.Time) string {
	return t.Format(TIME_FORMAT)
}

// Ife looks like t ? a : b
func Ife(t bool, a, b interface{}) interface{} {
	if t {
		return a
	}
	return b
}

// StrAlign returns string with fixed width
func StrAlign(s string, width int) string {
	blen := len(s)
	ulen := len([]rune(s))
	// b + 3u = blen
	// b + u = ulen
	// w = b + 2u
	b := (3*ulen - blen) / 2
	u := (blen - ulen) / 2
	w := 2*u + b

	if width > w {
		return s + strings.Repeat(" ", width-w)
	}

	for bi := width / 3; bi < blen; bi++ {
		if !utf8.ValidString(s[:bi]) {
			continue
		}
		ui := len([]rune(s[:bi]))
		if (3*ui-bi)/2+(bi-ui) >= width {
			return s[:bi]
		}
	}
	return s
}

// StrsTrimSpace returns string array without blank string
func StrsTrimSpace(ss []string) []string {
	vs := []string{}
	for _, s := range ss {
		if strings.TrimSpace(s) == "" {
			continue
		}
		vs = append(vs, s)
	}
	return vs
}

// StrsCountMap returns map of string's occurrence number
func StrsCountMap(ss []string) map[string]int {
	cm := make(map[string]int)
	for _, s := range ss {
		cm[s]++
	}
	return cm
}

// IntRound returns float's rounding number
func IntRound(num float64) int64 {
	return int64(math.Floor(num + 0.5))
}

// IntLimit returns num in range [min, max]
func IntLimit(num, min, max int64) int64 {
	if num < min {
		num = min
	}
	if num > max {
		num = max
	}
	return num
}

// FloatLimit returns num in range [min, max]
func FloatLimit(num, min, max float64) float64 {
	if num < min {
		num = min
	}
	if num > max {
		num = max
	}
	return num
}

// MapKeys returns sorted keys of a map
func MapKeys(mm interface{}) []interface{} {
	keys := []interface{}{}

	ref := reflect.ValueOf(mm)
	if ref.Kind() != reflect.Map {
		Fatalf("unsupported type %v", ref.Kind())
		return keys
	}

	kind := ref.Type().Key().Kind()
	switch kind {
	case reflect.String:
		tkeys := []string{}
		for _, k := range ref.MapKeys() {
			tkeys = append(tkeys, k.String())
		}
		sort.Strings(tkeys)
		for _, k := range tkeys {
			keys = append(keys, k)
		}
	case reflect.Int, reflect.Int64:
		tkeys := []int{}
		for _, k := range ref.MapKeys() {
			tkeys = append(tkeys, int(k.Int()))
		}
		sort.Ints(tkeys)
		for _, k := range tkeys {
			keys = append(keys, k)
		}
	case reflect.Float64:
		tkeys := []float64{}
		for _, k := range ref.MapKeys() {
			tkeys = append(tkeys, k.Float())
		}
		sort.Float64s(tkeys)
		for _, k := range tkeys {
			keys = append(keys, k)
		}
	default:
		Fatalf("unsupported key type %v", kind)
	}

	return keys
}

// JsonFormat returns json formatted string
func JsonFormat(data interface{}) string {
	mi, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return fmt.Sprintf("%s: %v", err, data)
	}

	return string(mi)
}

// JsonSave save var into file
func JsonSave(fname string, v interface{}) {
	f, err := os.Create(fname)
	if err != nil {
		Fatalf(err)
	}
	defer f.Close()

	e := json.NewEncoder(f)
	err = e.Encode(v)
	if err != nil {
		Fatalf(err)
	}
}

// JsonLoad load var from file
func JsonLoad(fname string, v interface{}) {
	f, err := os.Open(fname)
	if err != nil {
		Fatalf(err)
	}
	defer f.Close()

	d := json.NewDecoder(f)
	err = d.Decode(v)
	if err != nil {
		Fatalf(err)
	}
}

// Base64Encode return encode string
func Base64Encode(a string) string {
	return base64.StdEncoding.EncodeToString([]byte(a))
}

// Base64Decode return decode string
func Base64Decode(a string) string {
	s, err := base64.StdEncoding.DecodeString(a)
	if err != nil {
		return ""
	}
	return string(s)
}
