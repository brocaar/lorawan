package lorawan

import (
	"errors"
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetMACPayloadAndSize(t *testing.T) {
	Convey("Given uplink=false and CID=LinkADRReq", t, func() {
		uplink := false
		c := LinkADRReq
		Convey("Then getMACPayloadAndSize returns LinkADRReqPayload{} with size 4", func() {
			p, s, err := GetMACPayloadAndSize(uplink, c)
			So(err, ShouldBeNil)
			So(p, ShouldHaveSameTypeAs, &LinkADRReqPayload{})
			So(s, ShouldEqual, 4)
		})

		Convey("Then running getMACPayloadAndSize twice should return different objects", func() {
			p1, _, err := GetMACPayloadAndSize(uplink, c)
			So(err, ShouldBeNil)
			p2, _, err := GetMACPayloadAndSize(uplink, c)
			So(err, ShouldBeNil)

			So(fmt.Sprintf("%p", p1.(*LinkADRReqPayload)), ShouldNotEqual, fmt.Sprintf("%p", p2.(*LinkADRReqPayload)))
		})
	})

	Convey("Given uplink=true and CID=LinkADRAns", t, func() {
		uplink := true
		c := LinkADRAns
		Convey("Then getMACPayloadAndSize returns LinkADRAnsPayload{} with size 1", func() {
			p, s, err := GetMACPayloadAndSize(uplink, c)
			So(err, ShouldBeNil)
			So(p, ShouldHaveSameTypeAs, &LinkADRAnsPayload{})
			So(s, ShouldEqual, 1)
		})
	})

	Convey("When testing mac commands within the proprietary range", t, func() {
		Convey("Then getting an unregistered returns an error", func() {
			_, _, err := GetMACPayloadAndSize(true, CID(128))
			So(err, ShouldNotBeNil)
		})

		Convey("When registering CID 128 with a payload-size of 12 for uplink", func() {
			So(RegisterProprietaryMACCommand(true, CID(128), 12), ShouldBeNil)

			Convey("Then getting the payload-size for this CID returns 12 and a ProprietaryMACCommandPayload type", func() {
				pl, size, err := GetMACPayloadAndSize(true, CID(128))
				So(err, ShouldBeNil)
				So(size, ShouldEqual, 12)
				So(pl, ShouldHaveSameTypeAs, &ProprietaryMACCommandPayload{})
			})
		})
	})
}

func TestMACCommand(t *testing.T) {
	Convey("Given an empty MACCommand", t, func() {
		var m MACCommand

		Convey("Given CID=LinkCheckAns, Payload=LinkCheckAnsPayload(Margin=10, GwCnt=15)", func() {
			m.CID = LinkCheckAns
			m.Payload = &LinkCheckAnsPayload{Margin: 10, GwCnt: 15}
			Convey("Then MarshalBinary returns []byte{2, 10, 15}", func() {
				b, err := m.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{2, 10, 15})
			})
		})

		Convey("Given the slice []byte{2, 10, 15}", func() {
			b := []byte{2, 10, 15}
			Convey("Given the direction is downlink", func() {
				Convey("Then UnmarshalBinary should return a MACCommand with CID=LinkCheckAns", func() {
					err := m.UnmarshalBinary(false, b)
					So(err, ShouldBeNil)
					So(m.CID, ShouldEqual, LinkCheckAns)
					Convey("And Payload should be of type *LinkCheckAnsPayload", func() {
						p, ok := m.Payload.(*LinkCheckAnsPayload)
						So(ok, ShouldBeTrue)
						Convey("And Margin=10, GwCnt=15", func() {
							So(p, ShouldResemble, &LinkCheckAnsPayload{Margin: 10, GwCnt: 15})
						})
					})
				})
			})

			Convey("Given the direction is uplink", func() {
				Convey("Then UnmarshalBinary should return an error", func() {
					err := m.UnmarshalBinary(true, b)
					So(err, ShouldNotBeNil)
				})
			})
		})
	})
}

func TestLinkCheckAnsPayload(t *testing.T) {
	Convey("Given a LinkCheckAnsPayload with Margin=123 and GwCnt=234", t, func() {
		p := LinkCheckAnsPayload{Margin: 123, GwCnt: 234}
		Convey("Then MarshalBinary returns []byte{123, 234}", func() {
			b, err := p.MarshalBinary()
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte{123, 234})
		})
	})

	Convey("Given the slice []byte{123, 234}", t, func() {
		b := []byte{123, 234}
		p := LinkCheckAnsPayload{}
		Convey("Then UnmarshalBinary returns a LinkCheckAnsPayload with Margin=123 and GwCnt=234", func() {
			err := p.UnmarshalBinary(b)
			So(err, ShouldBeNil)
			So(p, ShouldResemble, LinkCheckAnsPayload{Margin: 123, GwCnt: 234})
		})
	})
}

func TestChMask(t *testing.T) {
	Convey("Given an empty ChMask", t, func() {
		var m ChMask

		Convey("Then MarshalBinary returns []byte{0, 0}", func() {
			b, err := m.MarshalBinary()
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte{0, 0})
		})

		Convey("Given channel 1, 2 and 13 are set to true", func() {
			m[0] = true
			m[1] = true
			m[12] = true
			Convey("Then MarshalBinary returns []byte{3, 16}", func() {
				b, err := m.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{3, 16})
			})
		})

		Convey("Given a slice of 3 bytes", func() {
			b := []byte{1, 2, 3}
			Convey("Then UnmarshalBinary returns an error", func() {
				err := m.UnmarshalBinary(b)
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Given the slice []byte{3, 16}", func() {
			b := []byte{3, 16}
			Convey("Then UnmarshalBinary returns a ChMask with channel 1, 2 and 3 set to true", func() {
				err := m.UnmarshalBinary(b)
				var exp ChMask
				exp[0] = true
				exp[1] = true
				exp[12] = true
				So(err, ShouldBeNil)
				So(m, ShouldResemble, exp)
			})
		})
	})
}

func TestRedundancy(t *testing.T) {
	Convey("Given an empty Redundancy", t, func() {
		var r Redundancy

		Convey("Then MarshalBinary returns []byte{0}", func() {
			b, err := r.MarshalBinary()
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte{0})
		})

		Convey("Given NbRep > 15", func() {
			r.NbRep = 16
			Convey("Then MarshalBinary returns an error", func() {
				_, err := r.MarshalBinary()
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Given ChMaskCntl > 7", func() {
			r.ChMaskCntl = 8
			Convey("Then MarshalBinary returns an error", func() {
				_, err := r.MarshalBinary()
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Given ChMaskCntl=1 and NbRep=5", func() {
			r.ChMaskCntl = 1
			r.NbRep = 5
			Convey("Then MarshalBinary returns []byte{13}", func() {
				b, err := r.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{21})
			})
		})

		Convey("Given the slice []byte{21}", func() {
			b := []byte{21}
			Convey("Then UnmarshalBinary returns a Redundancy with ChMaskCntl=1 and NbRep=5", func() {
				err := r.UnmarshalBinary(b)
				So(err, ShouldBeNil)
				So(r.ChMaskCntl, ShouldEqual, 1)
				So(r.NbRep, ShouldEqual, 5)
			})
		})
	})
}

func TestLinkADRReqPayload(t *testing.T) {
	Convey("Given an empty LinkADRReqPayload", t, func() {
		var p LinkADRReqPayload

		Convey("Then MarshalBinary returns []byte{0, 0, 0, 0}", func() {
			b, err := p.MarshalBinary()
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte{0, 0, 0, 0})
		})

		Convey("Given a DataRate > 15", func() {
			p.DataRate = 16
			Convey("Then MarshalBinary returns an error", func() {
				_, err := p.MarshalBinary()
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Given a TXPower > 15", func() {
			p.TXPower = 16
			Convey("Then MarshalBinary returns an error", func() {
				_, err := p.MarshalBinary()
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Given a LinkADRReqPayload with DataRate=1, TXPower=2, ChMask(channel 3=true) and Redundancy(ChMaskCntl=4, NbRep=5)", func() {
			var cm ChMask
			cm[2] = true

			p.DataRate = 1
			p.TXPower = 2
			p.ChMask = cm
			p.Redundancy = Redundancy{ChMaskCntl: 4, NbRep: 5}

			Convey("Then MarshalBinary returns []byte{18, 4, 0, 69}", func() {
				b, err := p.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{18, 4, 0, 69})
			})
		})

		Convey("Given the slice []byte{18, 4, 0, 69}", func() {
			b := []byte{18, 4, 0, 69}
			Convey("The UnmarshalBinary returns a LinkADRReqPayload with DataRate=1, TXPower=2, ChMask(channel 3=true) and Redundancy(ChMaskCntl=4, NbRep=5)", func() {
				err := p.UnmarshalBinary(b)
				So(err, ShouldBeNil)

				var cm ChMask
				cm[2] = true
				var exp LinkADRReqPayload
				exp.DataRate = 1
				exp.TXPower = 2
				exp.ChMask = cm
				exp.Redundancy = Redundancy{ChMaskCntl: 4, NbRep: 5}

				So(p, ShouldResemble, exp)
			})
		})

	})
}

func TestLinkADRAnsPayload(t *testing.T) {
	Convey("Given an empty LinkADRAnsPayload", t, func() {
		var p LinkADRAnsPayload
		Convey("Then MarshalBinary returns []byte{0}", func() {
			b, err := p.MarshalBinary()
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte{0})
		})

		testTable := []struct {
			ChannelMaskACK bool
			DataRateACK    bool
			PowerACK       bool
			Bytes          []byte
		}{
			{true, false, false, []byte{1}},
			{false, true, false, []byte{2}},
			{false, false, true, []byte{4}},
			{true, true, true, []byte{7}},
		}

		for _, test := range testTable {
			Convey(fmt.Sprintf("Given a LinkADRAnsPayload with ChannelMaskACK=%v, DataRateACK=%v, PowerACK=%v", test.ChannelMaskACK, test.DataRateACK, test.PowerACK), func() {
				p.ChannelMaskACK = test.ChannelMaskACK
				p.DataRateACK = test.DataRateACK
				p.PowerACK = test.PowerACK
				Convey(fmt.Sprintf("Then MarshalBinary returns %v", test.Bytes), func() {
					b, err := p.MarshalBinary()
					So(err, ShouldBeNil)
					So(b, ShouldResemble, test.Bytes)
				})
			})

			Convey(fmt.Sprintf("Given a slice %v", test.Bytes), func() {
				b := test.Bytes
				Convey(fmt.Sprintf("Then UnmarshalBinary returns a LinkADRAnsPayload with ChannelMaskACK=%v, DataRateACK=%v, PowerACK=%v", test.ChannelMaskACK, test.DataRateACK, test.PowerACK), func() {
					exp := LinkADRAnsPayload{
						ChannelMaskACK: test.ChannelMaskACK,
						DataRateACK:    test.DataRateACK,
						PowerACK:       test.PowerACK,
					}
					err := p.UnmarshalBinary(b)
					So(err, ShouldBeNil)
					So(p, ShouldResemble, exp)
				})
			})
		}
	})
}

func TestDutyCycleReqPayload(t *testing.T) {
	Convey("Given an empty DutyCycleReqPayload", t, func() {
		var p DutyCycleReqPayload
		Convey("Then MarshalBinary returns []byte{0}", func() {
			b, err := p.MarshalBinary()
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte{0})
		})

		Convey("Given a MaxDCCycle=15", func() {
			p.MaxDCycle = 16
			Convey("Then MarshalBinary returns an error", func() {
				_, err := p.MarshalBinary()
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Given a MaxDCCycle=254", func() {
			p.MaxDCycle = 254
			Convey("Then MarshalBinary returns an error", func() {
				_, err := p.MarshalBinary()
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Given A MaxDCCycle=13", func() {
			p.MaxDCycle = 13
			Convey("Then MarshalBinary returns []byte{13}", func() {
				b, err := p.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{13})
			})
		})

		Convey("Given a slice []byte{13}", func() {
			b := []byte{13}
			Convey("Then UnmarshalBinary returns a DutyCycleReqPayload with MaxDCCycle=13", func() {
				err := p.UnmarshalBinary(b)
				So(err, ShouldBeNil)
				So(p, ShouldResemble, DutyCycleReqPayload{13})
			})
		})
	})
}

func TestDLSettings(t *testing.T) {
	Convey("Given an empty DLSettings", t, func() {
		var s DLSettings
		Convey("Then MarshalBinary returns []byte{0}", func() {
			b, err := s.MarshalBinary()
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte{0})
		})

		Convey("Given a RX2DataRate > 15", func() {
			s.RX2DataRate = 16
			Convey("Then MarshalBinary returns an error", func() {
				_, err := s.MarshalBinary()
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Given a RX1DROffset > 7", func() {
			s.RX1DROffset = 8
			Convey("Then MarshalBinary returns an error", func() {
				_, err := s.MarshalBinary()
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Given RX2DataRate=15, RX1DROffset=7 and OptNeg=true", func() {
			s.RX2DataRate = 15
			s.RX1DROffset = 7
			s.OptNeg = true
			Convey("Then MarshalBinary returns []byte{127}", func() {
				b, err := s.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{255})
			})

			Convey("Then MarshalText returns ff", func() {
				b, err := s.MarshalText()
				So(err, ShouldBeNil)
				So(string(b), ShouldEqual, "ff")
			})
		})

		Convey("Given the hex string ff", func() {
			h := "ff"

			Convey("Then UnmarshalText returns a DLSettings with RX2DataRate=15, RX1DROffset=7 and OptNeg=true", func() {
				So(s.UnmarshalText([]byte(h)), ShouldBeNil)
				So(s, ShouldResemble, DLSettings{
					RX2DataRate: 15,
					RX1DROffset: 7,
					OptNeg:      true,
				})
			})
		})

		Convey("Given a slice []byte{255}", func() {
			b := []byte{255}
			Convey("Then UnmarshalBinary returns a DLSettings with RX2DataRate=15, RX1DROffset=7 and OptNeg=true", func() {
				err := s.UnmarshalBinary(b)
				So(err, ShouldBeNil)
				So(s, ShouldResemble, DLSettings{RX2DataRate: 15, RX1DROffset: 7, OptNeg: true})
			})
		})
	})
}

func TestRXParamSetupReqPayload(t *testing.T) {
	Convey("Given an empty RXParamSetupReqPayload", t, func() {
		var p RXParamSetupReqPayload
		Convey("Then MarshalBinary returns []byte{0, 0, 0, 0}", func() {
			b, err := p.MarshalBinary()
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte{0, 0, 0, 0})
		})

		Convey("Given Frequency > 2^24-1", func() {
			p.Frequency = 1677721600 // 2^24
			Convey("Then MarshalBinary returns an error", func() {
				_, err := p.MarshalBinary()
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Given Frequency=26265700 and DLSettings(RX2DataRate=11, RX1DROffset=3)", func() {
			p.Frequency = 26265700
			p.DLSettings = DLSettings{RX2DataRate: 11, RX1DROffset: 3}
			Convey("Then MarshalBinary returns []byte{1, 2, 4, 59}", func() {
				b, err := p.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{59, 1, 2, 4})
			})
		})

		Convey("Given a slice []byte{59, 1, 2, 4}", func() {
			b := []byte{59, 1, 2, 4}
			Convey("Then UnmarshalBinary returns a RXParamSetupReqPayload with Frequency=26265700 and DLSettings(RX2DataRate=11, RX1DROffset=3)", func() {
				exp := RXParamSetupReqPayload{
					Frequency:  26265700,
					DLSettings: DLSettings{RX2DataRate: 11, RX1DROffset: 3},
				}
				err := p.UnmarshalBinary(b)
				So(err, ShouldBeNil)
				So(p, ShouldResemble, exp)
			})
		})
	})
}

func TestRXParamSetupAnsPayload(t *testing.T) {
	Convey("Given an empty RXParamSetupAnsPayload", t, func() {
		var p RXParamSetupAnsPayload
		Convey("Then MarshalBinary returns []byte{0}", func() {
			b, err := p.MarshalBinary()
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte{0})
		})

		testTable := []struct {
			ChannelACK     bool
			RX2DataRateACK bool
			RX1DROffsetACK bool
			Bytes          []byte
		}{
			{true, false, false, []byte{1}},
			{false, true, false, []byte{2}},
			{false, false, true, []byte{4}},
			{true, true, true, []byte{7}},
		}

		for _, test := range testTable {
			Convey(fmt.Sprintf("Given ChannelACK=%v, RX2DataRateACK=%v, RX1DROffsetACK=%v", test.ChannelACK, test.RX2DataRateACK, test.RX1DROffsetACK), func() {
				p.ChannelACK = test.ChannelACK
				p.RX2DataRateACK = test.RX2DataRateACK
				p.RX1DROffsetACK = test.RX1DROffsetACK
				Convey(fmt.Sprintf("Then marshalBinary returns %v", test.Bytes), func() {
					b, err := p.MarshalBinary()
					So(err, ShouldBeNil)
					So(b, ShouldResemble, test.Bytes)
				})
			})

			Convey(fmt.Sprintf("Given slice %v", test.Bytes), func() {
				b := test.Bytes
				Convey(fmt.Sprintf("Then UnmarshalBinary returns a RXParamSetupAnsPayload with ChannelACK=%v, RX2DataRateACK=%v, RX1DROffsetACK=%v", test.ChannelACK, test.RX2DataRateACK, test.RX1DROffsetACK), func() {
					exp := RXParamSetupAnsPayload{
						ChannelACK:     test.ChannelACK,
						RX2DataRateACK: test.RX2DataRateACK,
						RX1DROffsetACK: test.RX1DROffsetACK,
					}
					err := p.UnmarshalBinary(b)
					So(err, ShouldBeNil)
					So(p, ShouldResemble, exp)
				})
			})
		}
	})
}

func TestDevStatusAnsPayload(t *testing.T) {
	Convey("Given an empty DevStatusAnsPayload", t, func() {
		var p DevStatusAnsPayload
		Convey("Then MarshalBinary returns []byte{0, 0}", func() {
			b, err := p.MarshalBinary()
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte{0, 0})
		})

		testTable := []struct {
			Battery uint8
			Margin  int8
			Bytes   []byte
		}{
			{0, -30, []byte{0, 34}},
			{255, 30, []byte{255, 30}},
			{127, -1, []byte{127, 63}},
			{127, 0, []byte{127, 0}},
		}

		for _, test := range testTable {
			Convey(fmt.Sprintf("Given Battery=%v and Margin=%v", test.Battery, test.Margin), func() {
				p.Battery = test.Battery
				p.Margin = test.Margin
				Convey(fmt.Sprintf("Then MarshalBinary returns %v", test.Bytes), func() {
					b, err := p.MarshalBinary()
					So(err, ShouldBeNil)
					So(b, ShouldResemble, test.Bytes)
				})
			})
		}
	})
}

func TestNewChannelReqPayload(t *testing.T) {
	Convey("Given an emtpy NewChannelReqPayload", t, func() {
		var p NewChannelReqPayload
		Convey("Then MarshalBinary returns []byte{0, 0, 0, 0, 0}", func() {
			b, err := p.MarshalBinary()
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte{0, 0, 0, 0, 0})
		})

		Convey("Given Freq > 2^24 - 1", func() {
			p.Freq = 16777216 * 100
			Convey("MarshalBinary returns an error", func() {
				_, err := p.MarshalBinary()
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Given MaxDR > 15", func() {
			p.MaxDR = 16
			Convey("MarshalBinary returns an error", func() {
				_, err := p.MarshalBinary()
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Given MinDR > 15", func() {
			p.MinDR = 16
			Convey("MarshalBinary returns an error", func() {
				_, err := p.MarshalBinary()
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Given ChIndex=3, Freq=26265700, MaxDR=5, MinDR=10", func() {
			p.ChIndex = 3
			p.Freq = 26265700
			p.MaxDR = 5
			p.MinDR = 10
			Convey("Then MarshalBinary returns []byte{3, 1, 2, 4, 90}", func() {
				b, err := p.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{3, 1, 2, 4, 90})
			})
		})

		Convey("Given a slice []byte{3, 1, 2, 4, 90}", func() {
			b := []byte{3, 1, 2, 4, 90}
			err := p.UnmarshalBinary(b)
			So(err, ShouldBeNil)
			So(p, ShouldResemble, NewChannelReqPayload{ChIndex: 3, Freq: 26265700, MaxDR: 5, MinDR: 10})
		})
	})
}

func TestNewChannelAnsPayload(t *testing.T) {
	Convey("Given an empty NewChannelAnsPayload", t, func() {
		var p NewChannelAnsPayload
		Convey("Then MarshalBinary returns []byte{0}", func() {
			b, err := p.MarshalBinary()
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte{0})
		})

		testTable := []struct {
			ChannelFrequencyOK bool
			DataRateRangeOK    bool
			Bytes              []byte
		}{
			{false, false, []byte{0}},
			{true, false, []byte{1}},
			{false, true, []byte{2}},
			{true, true, []byte{3}},
		}

		for _, test := range testTable {
			Convey(fmt.Sprintf("Given ChannelFrequencyOK=%v, DataRateRangeOK=%v", test.ChannelFrequencyOK, test.DataRateRangeOK), func() {
				p.ChannelFrequencyOK = test.ChannelFrequencyOK
				p.DataRateRangeOK = test.DataRateRangeOK
				Convey(fmt.Sprintf("Then MarshalBinary returns %v", test.Bytes), func() {
					b, err := p.MarshalBinary()
					So(err, ShouldBeNil)
					So(b, ShouldResemble, test.Bytes)
				})
			})

			Convey(fmt.Sprintf("Given a slice %v", test.Bytes), func() {
				b := test.Bytes
				Convey(fmt.Sprintf("Then UnmarshalBinary returns a NewChannelAnsPayload with ChannelFrequencyOK=%v, DataRateRangeOK=%v", test.ChannelFrequencyOK, test.DataRateRangeOK), func() {
					err := p.UnmarshalBinary(b)
					So(err, ShouldBeNil)
					So(p, ShouldResemble, NewChannelAnsPayload{ChannelFrequencyOK: test.ChannelFrequencyOK, DataRateRangeOK: test.DataRateRangeOK})
				})
			})
		}
	})
}

func TestRXTimingSetupReqPayload(t *testing.T) {
	Convey("Given an emtpy RXTimingSetupReqPayload", t, func() {
		var p RXTimingSetupReqPayload
		Convey("Then MarshalBinary returns []byte{0}", func() {
			b, err := p.MarshalBinary()
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte{0})
		})

		Convey("Given Delay > 15", func() {
			p.Delay = 16
			Convey("Then MarshalBinary returns an error", func() {
				_, err := p.MarshalBinary()
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Given Delay=15", func() {
			p.Delay = 15
			Convey("Then MarshalBinary returns []byte{15}", func() {
				b, err := p.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{15})
			})
		})

		Convey("Given a slice []byte{15}", func() {
			b := []byte{15}
			Convey("Then UnmarshalBinary returns RXTimingSetupReqPayload with Delay=15", func() {
				err := p.UnmarshalBinary(b)
				So(err, ShouldBeNil)
				So(p, ShouldResemble, RXTimingSetupReqPayload{Delay: 15})
			})
		})
	})
}

func TestDLChannelReqPayload(t *testing.T) {
	Convey("Given an empty DLChannelReqPayload", t, func() {
		var p DLChannelReqPayload

		Convey("Then MarshalBinary returns []byte{0, 0, 0, 0}", func() {
			b, err := p.MarshalBinary()
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte{0, 0, 0, 0})
		})

		tests := []struct {
			ChIndex uint8
			Freq    uint32
			Bytes   []byte
			Error   error
		}{
			{0, 868100000, []byte{0, 40, 118, 132}, nil},
			{1, 868200000, []byte{1, 16, 122, 132}, nil},
			{0, 868100099, nil, errors.New("lorawan: Freq must be a multiple of 100")},
		}

		for i, test := range tests {
			Convey(fmt.Sprintf("Given ChIndex: %d, Freq: %d [%d]", test.ChIndex, test.Freq, i), func() {
				p.ChIndex = test.ChIndex
				p.Freq = test.Freq

				if test.Error != nil {
					Convey(fmt.Sprintf("Then MarshalBinary returns error %s", test.Error), func() {
						_, err := p.MarshalBinary()
						So(err, ShouldResemble, test.Error)
					})
				} else {
					Convey(fmt.Sprintf("Then MarshalBinary returns %v", test.Bytes), func() {
						b, err := p.MarshalBinary()
						So(err, ShouldBeNil)
						So(b, ShouldResemble, test.Bytes)
					})
				}
			})
		}

		for i, test := range tests {
			if test.Error != nil {
				continue
			}

			Convey(fmt.Sprintf("When unmarshaling %v [%d]", test.Bytes, i), func() {
				So(p.UnmarshalBinary(test.Bytes), ShouldBeNil)

				Convey(fmt.Sprintf("Then ChIndex: %d and Freq: %d are expected", test.ChIndex, test.Freq), func() {
					So(p.ChIndex, ShouldEqual, test.ChIndex)
					So(p.Freq, ShouldEqual, test.Freq)
				})
			})
		}

	})
}

func TestDLChannelAnsPayload(t *testing.T) {
	Convey("Given an empty DLChannelAnsPayload", t, func() {
		var p DLChannelAnsPayload

		Convey("Then MarshalBinary returns []byte{0}", func() {
			b, err := p.MarshalBinary()
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte{0})
		})

		tests := []struct {
			ChannelFrequencyOK    bool
			UplinkFrequencyExists bool
			Bytes                 []byte
		}{
			{false, false, []byte{0}},
			{true, false, []byte{1}},
			{false, true, []byte{2}},
			{true, true, []byte{3}},
		}

		for i, test := range tests {
			Convey(fmt.Sprintf("Given ChannelFrequencyOK: %t, UplinkFrequencyExists: %t [%d]", test.ChannelFrequencyOK, test.UplinkFrequencyExists, i), func() {
				p.ChannelFrequencyOK = test.ChannelFrequencyOK
				p.UplinkFrequencyExists = test.UplinkFrequencyExists

				Convey(fmt.Sprintf("Then MarshalBinary returns %v", test.Bytes), func() {
					b, err := p.MarshalBinary()
					So(err, ShouldBeNil)
					So(b, ShouldResemble, test.Bytes)
				})
			})
		}

		for i, test := range tests {
			Convey(fmt.Sprintf("When Unmarshaling %v [%d]", test.Bytes, i), func() {
				So(p.UnmarshalBinary(test.Bytes), ShouldBeNil)

				Convey(fmt.Sprintf("Then ChannelFrequencyOK: %t and UplinkFrequencyExists: %t are expected", test.ChannelFrequencyOK, test.UplinkFrequencyExists), func() {
					So(p.ChannelFrequencyOK, ShouldEqual, test.ChannelFrequencyOK)
					So(p.UplinkFrequencyExists, ShouldEqual, test.UplinkFrequencyExists)
				})
			})
		}
	})
}

type macPayloadTest struct {
	Payload       MACCommandPayload
	ExpectedBytes []byte
	ExpectedError error
}

// TestMACPayloads tests the mac-command payloads
// TODO: refactor above tests in this new framework
func TestMACPayloads(t *testing.T) {
	Convey("Testing PingSlotInfoReqPayload", t, func() {
		tests := []macPayloadTest{
			{
				Payload:       &PingSlotInfoReqPayload{Periodicity: 3},
				ExpectedBytes: []byte{3},
			},
			{
				Payload:       &PingSlotInfoReqPayload{Periodicity: 8},
				ExpectedError: errors.New("lorawan: max value of Periodicity is 7"),
			},
		}

		testMACPayloads(func() MACCommandPayload { return &PingSlotInfoReqPayload{} }, tests)
	})

	Convey("Testing BeaconFreqReqPayload", t, func() {
		tests := []macPayloadTest{
			{
				Payload:       &BeaconFreqReqPayload{Frequency: 101},
				ExpectedError: errors.New("lorawan: Frequency must be a multiple of 100"),
			},
			{
				Payload:       &BeaconFreqReqPayload{Frequency: 1677721600},
				ExpectedError: errors.New("lorawan: max value of Frequency is 2^24 - 1"),
			},
			{
				Payload:       &BeaconFreqReqPayload{Frequency: 868100000},
				ExpectedBytes: []byte{40, 118, 132},
			},
		}

		testMACPayloads(func() MACCommandPayload { return &BeaconFreqReqPayload{} }, tests)
	})

	Convey("Testing BeaconFreqAnsPayload", t, func() {
		tests := []macPayloadTest{
			{
				Payload:       &BeaconFreqAnsPayload{BeaconFrequencyOK: true},
				ExpectedBytes: []byte{1},
			},
			{
				Payload:       &BeaconFreqAnsPayload{BeaconFrequencyOK: false},
				ExpectedBytes: []byte{0},
			},
		}

		testMACPayloads(func() MACCommandPayload { return &BeaconFreqAnsPayload{} }, tests)
	})

	Convey("Testing PingSlotChannelReqPayload", t, func() {
		tests := []macPayloadTest{
			{
				Payload:       &PingSlotChannelReqPayload{Frequency: 101},
				ExpectedError: errors.New("lorawan: Frequency must be a multiple of 100"),
			},
			{
				Payload:       &PingSlotChannelReqPayload{Frequency: 1677721600},
				ExpectedError: errors.New("lorawan: max value of Frequency is 2^24 - 1"),
			},
			{
				Payload:       &PingSlotChannelReqPayload{Frequency: 868100000, DR: 16},
				ExpectedError: errors.New("lorawan: max value of DR is 15"),
			},
			{
				Payload:       &PingSlotChannelReqPayload{Frequency: 868100000, DR: 5},
				ExpectedBytes: []byte{40, 118, 132, 5},
			},
		}

		testMACPayloads(func() MACCommandPayload { return &PingSlotChannelReqPayload{} }, tests)
	})

	Convey("Testing PingSlotChannelAnsPayload", t, func() {
		tests := []macPayloadTest{
			{
				Payload:       &PingSlotChannelAnsPayload{DataRateOK: true},
				ExpectedBytes: []byte{2},
			},
			{
				Payload:       &PingSlotChannelAnsPayload{ChannelFrequencyOK: true},
				ExpectedBytes: []byte{1},
			},
			{
				Payload:       &PingSlotChannelAnsPayload{DataRateOK: true, ChannelFrequencyOK: true},
				ExpectedBytes: []byte{3},
			},
		}

		testMACPayloads(func() MACCommandPayload { return &PingSlotChannelAnsPayload{} }, tests)
	})

	Convey("Testing DeviceTimeAnsPayload", t, func() {
		tests := []macPayloadTest{
			{
				Payload:       &DeviceTimeAnsPayload{TimeSinceGPSEpoch: time.Second},
				ExpectedBytes: []byte{1, 0, 0, 0, 0},
			},
			{
				Payload:       &DeviceTimeAnsPayload{TimeSinceGPSEpoch: time.Second + (2 * 3906250)},
				ExpectedBytes: []byte{1, 0, 0, 0, 2},
			},
		}

		testMACPayloads(func() MACCommandPayload { return &DeviceTimeAnsPayload{} }, tests)
	})

	Convey("Testing ResetIndPayload", t, func() {
		tests := []macPayloadTest{
			{
				Payload:       &ResetIndPayload{DevLoRaWANVersion: Version{Minor: 1}},
				ExpectedBytes: []byte{1},
			},
			{
				Payload:       &ResetIndPayload{DevLoRaWANVersion: Version{Minor: 8}},
				ExpectedError: errors.New("lorawan: max value of Minor is 7"),
			},
		}

		testMACPayloads(func() MACCommandPayload { return &ResetIndPayload{} }, tests)
	})

	Convey("Testing ResetConfPayload", t, func() {
		tests := []macPayloadTest{
			{
				Payload:       &ResetConfPayload{ServLoRaWANVersion: Version{Minor: 1}},
				ExpectedBytes: []byte{1},
			},
			{
				Payload:       &ResetConfPayload{ServLoRaWANVersion: Version{Minor: 8}},
				ExpectedError: errors.New("lorawan: max value of Minor is 7"),
			},
		}

		testMACPayloads(func() MACCommandPayload { return &ResetConfPayload{} }, tests)
	})

	Convey("Testing RekeyIndPayload", t, func() {
		tests := []macPayloadTest{
			{
				Payload:       &RekeyIndPayload{DevLoRaWANVersion: Version{Minor: 1}},
				ExpectedBytes: []byte{1},
			},
			{
				Payload:       &RekeyIndPayload{DevLoRaWANVersion: Version{Minor: 8}},
				ExpectedError: errors.New("lorawan: max value of Minor is 7"),
			},
		}

		testMACPayloads(func() MACCommandPayload { return &RekeyIndPayload{} }, tests)
	})

	Convey("Testing RekeyConfPayload", t, func() {
		tests := []macPayloadTest{
			{
				Payload:       &RekeyConfPayload{ServLoRaWANVersion: Version{Minor: 1}},
				ExpectedBytes: []byte{1},
			},
			{
				Payload:       &RekeyConfPayload{ServLoRaWANVersion: Version{Minor: 8}},
				ExpectedError: errors.New("lorawan: max value of Minor is 7"),
			},
		}

		testMACPayloads(func() MACCommandPayload { return &RekeyConfPayload{} }, tests)
	})

	Convey("Testing ADRParamSetupReqPayload", t, func() {
		tests := []macPayloadTest{
			{
				Payload: &ADRParamSetupReqPayload{ADRParam: ADRParam{
					LimitExp: 10,
					DelayExp: 15,
				}},
				ExpectedBytes: []byte{175},
			},
			{
				Payload: &ADRParamSetupReqPayload{ADRParam: ADRParam{
					LimitExp: 16,
				}},
				ExpectedError: errors.New("lorawan: max value of LimitExp is 15"),
			},
			{
				Payload: &ADRParamSetupReqPayload{ADRParam: ADRParam{
					DelayExp: 16,
				}},
				ExpectedError: errors.New("lorawan: max value of DelayExp is 15"),
			},
		}

		testMACPayloads(func() MACCommandPayload { return &ADRParamSetupReqPayload{} }, tests)
	})

	Convey("Testing ForceRejoinReq", t, func() {
		tests := []macPayloadTest{
			{
				Payload: &ForceRejoinReqPayload{
					Period:     3,
					MaxRetries: 4,
					RejoinType: 2,
					DR:         5,
				},
				ExpectedBytes: []byte{37, 28},
			},
			{
				Payload: &ForceRejoinReqPayload{
					Period:     8,
					MaxRetries: 4,
					RejoinType: 2,
					DR:         5,
				},
				ExpectedError: errors.New("lorawan: max value of Period is 7"),
			},
			{
				Payload: &ForceRejoinReqPayload{
					Period:     3,
					MaxRetries: 8,
					RejoinType: 2,
					DR:         5,
				},
				ExpectedError: errors.New("lorawan: max value of MaxRetries is 7"),
			},
			{
				Payload: &ForceRejoinReqPayload{
					Period:     3,
					MaxRetries: 4,
					RejoinType: 3,
					DR:         5,
				},
				ExpectedError: errors.New("lorawan: RejoinType must be 0 or 2"),
			},
			{
				Payload: &ForceRejoinReqPayload{
					Period:     3,
					MaxRetries: 4,
					RejoinType: 2,
					DR:         16,
				},
				ExpectedError: errors.New("lorawan: max value of DR is 15"),
			},
		}

		testMACPayloads(func() MACCommandPayload { return &ForceRejoinReqPayload{} }, tests)
	})

	Convey("Testing RejoinParamSetupReqPayload", t, func() {
		tests := []macPayloadTest{
			{
				Payload: &RejoinParamSetupReqPayload{
					MaxTimeN:  14,
					MaxCountN: 15,
				},
				ExpectedBytes: []byte{239},
			},
			{
				Payload: &RejoinParamSetupReqPayload{
					MaxTimeN:  16,
					MaxCountN: 15,
				},
				ExpectedError: errors.New("lorawan: max value of MaxTimeN is 15"),
			},
			{
				Payload: &RejoinParamSetupReqPayload{
					MaxTimeN:  14,
					MaxCountN: 16,
				},
				ExpectedError: errors.New("lorawan: max value of MaxCountN is 15"),
			},
		}

		testMACPayloads(func() MACCommandPayload { return &RejoinParamSetupReqPayload{} }, tests)
	})

	Convey("Testing RejoinParamSetupAnsPayload", t, func() {
		tests := []macPayloadTest{
			{
				Payload: &RejoinParamSetupAnsPayload{
					TimeOK: true,
				},
				ExpectedBytes: []byte{1},
			},
		}

		testMACPayloads(func() MACCommandPayload { return &RejoinParamSetupAnsPayload{} }, tests)
	})
}

func testMACPayloads(newPLFunc func() MACCommandPayload, tests []macPayloadTest) {
	for i, t := range tests {
		Convey(fmt.Sprintf("Testing: %+v [%d]", t.Payload, i), func() {
			b, err := t.Payload.MarshalBinary()
			if t.ExpectedError != nil {
				Convey("Then the expected error is returned", func() {
					So(err, ShouldNotBeNil)
					So(err, ShouldResemble, t.ExpectedError)
				})
				return
			}
			So(err, ShouldBeNil)
			So(b, ShouldResemble, t.ExpectedBytes)

			Convey("Then unmarshal with a different byteslice size returns an error", func() {
				pl := newPLFunc()
				So(pl.UnmarshalBinary(b[0:len(b)-1]), ShouldBeError)

				b2 := make([]byte, len(b)+1)
				copy(b2, b)
				So(pl.UnmarshalBinary(b2), ShouldBeError)
			})

			Convey("Then unmarshal results in the same payload", func() {
				pl := newPLFunc()
				So(pl.UnmarshalBinary(b), ShouldBeNil)
				So(pl, ShouldResemble, t.Payload)
			})
		})
	}
}
