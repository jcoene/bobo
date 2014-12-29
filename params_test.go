package bobo

import (
	"net/url"
	"reflect"
	"testing"
)

func TestParamsGet(t *testing.T) {
	vals := url.Values{}
	vals.Add("id", "101405")
	params := Params(vals)

	if params.Get("id") != "101405" {
		t.Errorf("unexpected value: %s", params.Get("id"))
	}

	if params.Get("z") != "" {
		t.Errorf("unexpected value: %s", params.Get("z"))
	}
}

func TestParamsInt64(t *testing.T) {
	vals := url.Values{}
	vals.Add("id", "101405")
	params := Params(vals)

	if params.Int64("id") != int64(101405) {
		t.Errorf("unexpected value: %d", params.Int64("id"))
	}

	if params.Int64("z") != int64(0) {
		t.Errorf("unexpected value: %d", params.Int64("z"))
	}
}

func TestParamsInt(t *testing.T) {
	vals := url.Values{}
	vals.Add("id", "101405")
	params := Params(vals)

	if params.Int("id") != int(101405) {
		t.Errorf("unexpected value: %d", params.Int("id"))
	}

	if params.Int("z") != int(0) {
		t.Errorf("unexpected value: %d", params.Int("z"))
	}
}

func TestParamsInt32(t *testing.T) {
	vals := url.Values{}
	vals.Add("id", "101405")
	params := Params(vals)

	if params.Int32("id") != int32(101405) {
		t.Errorf("unexpected value: %d", params.Int32("id"))
	}

	if params.Int32("z") != int32(0) {
		t.Errorf("unexpected value: %d", params.Int32("z"))
	}
}

func TestParamsMap(t *testing.T) {
	vals := url.Values{}
	vals.Add("id", "101405")
	vals.Add("name", "Jason")
	params := Params(vals)

	expect := map[string]string{"id": "101405", "name": "Jason"}
	if !reflect.DeepEqual(params.Map(), expect) {
		t.Errorf("expected %+v, got %+v", expect, params.Map())
	}
}

func TestParamsInt64s(t *testing.T) {
	vals := url.Values{}
	vals.Add("ids", "101405,90210,99999")
	params := Params(vals)

	expect := []int64{101405, 90210, 99999}
	if !reflect.DeepEqual(params.Int64s("ids"), expect) {
		t.Errorf("expected %+v, got %+v", expect, params.Int64s("ids"))
	}

	vals.Set("ids", "101405,john,monkey,banana,32124")
	expect = []int64{101405, 32124}
	if !reflect.DeepEqual(params.Int64s("ids"), expect) {
		t.Errorf("expected %+v, got %+v", expect, params.Int64s("ids"))
	}

	expect = []int64{}
	if !reflect.DeepEqual(params.Int64s("z"), expect) {
		t.Errorf("expected %+v, got %+v", expect, params.Int64s("z"))
	}
}
