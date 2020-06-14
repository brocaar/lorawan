package backend

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/brocaar/lorawan"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/require"
)

func TestHEXBytes(t *testing.T) {
	Convey("Given a HEXBytes", t, func() {
		hb := HEXBytes{1, 2, 3, 4}

		Convey("Then String returns the expected string", func() {
			So(hb.String(), ShouldEqual, "01020304")
		})

		Convey("Then MarshalText returns the expected output", func() {
			txt, err := hb.MarshalText()
			So(err, ShouldBeNil)
			So(string(txt), ShouldEqual, "01020304")
		})
	})

	Convey("Given an empty HEXBytes", t, func() {
		hb := HEXBytes{}

		Convey("Then UnmarshalText(\"01020304\") results in the expected HEXBytes", func() {
			So(hb.UnmarshalText([]byte("01020304")), ShouldBeNil)
			So(hb, ShouldResemble, HEXBytes{1, 2, 3, 4})
		})
	})
}

func TestFrequency(t *testing.T) {
	Convey("Given a Frequency instance", t, func() {
		f := Frequency(868100000)

		Convey("Then MarshalJSON returns the expected value", func() {
			b, err := f.MarshalJSON()
			So(err, ShouldBeNil)
			So(string(b), ShouldEqual, "868.1")
		})

		Convey("Then UnmarshalJSON unmarshals to the expected value", func() {
			So(f.UnmarshalJSON([]byte("868.2")), ShouldBeNil)
			So(f, ShouldEqual, Frequency(868200000))
		})
	})
}

func TestPercentage(t *testing.T) {
	Convey("Given a Percentage instance", t, func() {
		p := Percentage(1)

		Convey("Then MarshalJSON returns the exepcted value", func() {
			b, err := p.MarshalJSON()
			So(err, ShouldBeNil)
			So(string(b), ShouldEqual, "0.01")
		})

		Convey("Then UnmarshalJSON unmarshals to the expected value", func() {
			So(p.UnmarshalJSON([]byte("0.02")), ShouldBeNil)
			So(p, ShouldEqual, Percentage(2))
		})
	})
}

func TestISO8601Time(t *testing.T) {
	Convey("Given an ISO8601Time instance", t, func() {
		ts := time.Date(2017, 12, 27, 17, 6, 35, 0, time.UTC)
		isoTS := ISO8601Time(ts)

		Convey("Then MarshalJSON returns the expected value", func() {
			b, err := json.Marshal(isoTS)
			So(err, ShouldBeNil)
			So(string(b), ShouldEqual, `"2017-12-27T17:06:35Z"`)
		})

		Convey("Then UnmarshalJSON unmarshals to the expected value", func() {
			var ts2 time.Time
			So(json.Unmarshal([]byte(`"2017-12-27T17:06:35Z"`), &ts2), ShouldBeNil)
			So(ts2.Equal(ts), ShouldBeTrue)
		})
	})
}

func TestKeyEnvelope(t *testing.T) {
	key := lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8}
	kek := lorawan.AES128Key{8, 7, 6, 5, 4, 3, 2, 1, 8, 7, 6, 5, 4, 3, 2, 1}

	t.Run("No kek label", func(t *testing.T) {
		assert := require.New(t)

		ke, err := NewKeyEnvelope("", nil, key)
		assert.NoError(err)
		assert.Equal(&KeyEnvelope{
			AESKey: HEXBytes(key[:]),
		}, ke)
	})

	t.Run("With kek label", func(t *testing.T) {
		assert := require.New(t)

		ke, err := NewKeyEnvelope("test-kek", kek[:], key)
		assert.NoError(err)
		assert.Equal(&KeyEnvelope{
			KEKLabel: "test-kek",
			AESKey:   HEXBytes([]byte{0xe3, 0xd5, 0xa4, 0x7b, 0xa2, 0x5c, 0xbe, 0x6e, 0x5d, 0xa8, 0x20, 0x84, 0x6e, 0xc, 0xb6, 0xa8, 0x2b, 0x75, 0xc, 0x59, 0xd8, 0x48, 0xec, 0x7a}),
		}, ke)

		t.Run("Unwrap", func(t *testing.T) {
			assert := require.New(t)

			keyRet, err := ke.Unwrap(kek[:])
			assert.NoError(err)
			assert.Equal(key, keyRet)
		})
	})
}
