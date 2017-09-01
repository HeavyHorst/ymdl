package ytdl

import (
	"reflect"
	"testing"
)

var formats = FormatList{
	FORMATS[18],
	FORMATS[22],
	FORMATS[34],
	FORMATS[37],
	FORMATS[133],
	FORMATS[139],
}

type formatListTestCase struct {
	Key             FormatKey
	FilterValues    interface{}
	ExpectedFormats FormatList
}

func TestFilter(t *testing.T) {
	cases := []formatListTestCase{
		formatListTestCase{
			FormatExtensionKey,
			[]interface{}{"mp4"},
			FormatList{formats[0], formats[1], formats[3], formats[4], formats[5]},
		},
		formatListTestCase{
			FormatResolutionKey,
			[]interface{}{"360p", "720p"},
			FormatList{formats[0], formats[1]},
		},
		formatListTestCase{
			FormatItagKey,
			[]interface{}{"22", "37"},
			FormatList{formats[1], formats[3]},
		},
		formatListTestCase{
			FormatAudioBitrateKey,
			[]interface{}{"96", "128"},
			FormatList{formats[0], formats[2]},
		},
		formatListTestCase{
			FormatResolutionKey,
			[]interface{}{""},
			FormatList{formats[5]},
		},
		formatListTestCase{
			FormatAudioBitrateKey,
			[]interface{}{"0"},
			FormatList{formats[4]},
		},
		formatListTestCase{
			FormatResolutionKey,
			[]interface{}{},
			nil,
		},
	}

	for _, v := range cases {
		f := formats.Filter(v.Key, v.FilterValues.([]interface{}))
		if !reflect.DeepEqual(v.ExpectedFormats, f) {
			t.Error("Format filter test case failed expected", v.ExpectedFormats, "got", f)
		}
	}
}

func TestExtremes(t *testing.T) {

	cases := []formatListTestCase{
		formatListTestCase{
			FormatResolutionKey,
			true,
			FormatList{formats[3]},
		},
		formatListTestCase{
			FormatResolutionKey,
			false,
			FormatList{formats[5]},
		},
		formatListTestCase{
			FormatAudioBitrateKey,
			true,
			FormatList{formats[1], formats[3]},
		},
		formatListTestCase{
			FormatAudioBitrateKey,
			false,
			FormatList{formats[4]},
		},
		formatListTestCase{
			FormatItagKey,
			true,
			formats,
		},
	}
	for _, v := range cases {
		f := formats.Extremes(v.Key, v.FilterValues.(bool))
		if !reflect.DeepEqual(v.ExpectedFormats, f) {
			t.Error("Format filter test case failed expected", v.ExpectedFormats, "got", f)
		}
	}
}

func TestSubtract(t *testing.T) {
	cases := []formatListTestCase{
		formatListTestCase{
			FormatExtensionKey,
			[]interface{}{"mp4"},
			FormatList{formats[2]},
		},
		formatListTestCase{
			FormatResolutionKey,
			[]interface{}{"480p", "360p", "240p", ""},
			FormatList{formats[1], formats[3]},
		},
		formatListTestCase{
			FormatResolutionKey,
			[]interface{}{},
			formats,
		},
	}
	for _, v := range cases {
		f := formats.Subtract(formats.Filter(v.Key, v.FilterValues.([]interface{})))
		if !reflect.DeepEqual(v.ExpectedFormats, f) {
			t.Error("Format filter test case failed expected", v.ExpectedFormats, "got", f)
		}
	}
}

func TestSort(t *testing.T) {
	cases := []formatListTestCase{
		formatListTestCase{
			FormatResolutionKey,
			formats,
			FormatList{
				formats[5],
				formats[4],
				formats[0],
				formats[2],
				formats[1],
				formats[3],
			},
		},
		formatListTestCase{
			FormatAudioBitrateKey,
			formats,
			FormatList{
				formats[4],
				formats[5],
				formats[0],
				formats[2],
				formats[1],
				formats[3],
			},
		},
	}

	for _, v := range cases {
		sorted := v.FilterValues.(FormatList).Copy()
		sorted.Sort(v.Key, false)
		if !reflect.DeepEqual(v.ExpectedFormats, sorted) {
			t.Error("FormatList sort failed")
		}
	}
}

func TestCopy(t *testing.T) {
	if !reflect.DeepEqual(formats, formats.Copy()) {
		t.Error("Copying format list failed")
	}
}
