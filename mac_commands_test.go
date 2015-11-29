package lorawan

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetMACPayloadAndSize(t *testing.T) {
	Convey("Given uplink=false and CID=LinkADRReq", t, func() {
		uplink := false
		c := LinkADRReq
		Convey("Then getMACPayloadAndSize returns LinkADRAnsPayload{} with size 4", func() {
			p, s, err := getMACPayloadAndSize(uplink, c)
			So(err, ShouldBeNil)
			So(p, ShouldHaveSameTypeAs, &LinkADRAnsPayload{})
			So(s, ShouldEqual, 4)
		})
	})

	Convey("Given uplink=true and CID=LinkADRAns", t, func() {
		uplink := true
		c := LinkADRAns
		Convey("Then getMACPayloadAndSize returns LinkADRAnsPayload{} with size 1", func() {
			p, s, err := getMACPayloadAndSize(uplink, c)
			So(err, ShouldBeNil)
			So(p, ShouldHaveSameTypeAs, &LinkADRAnsPayload{})
			So(s, ShouldEqual, 1)
		})
	})
}

func TestMACCommand(t *testing.T) {
	Convey("Given an empty MACCommand", t, func() {
		var m MACCommand
		Convey("Then MarshalBinary returns []byte{0}", func() {
			b, err := m.MarshalBinary()
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte{0})
		})

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
				m.uplink = false
				Convey("Then UnmarshalBinary should return a MACCommand with CID=LinkCheckAns", func() {
					err := m.UnmarshalBinary(b)
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
				m.uplink = true
				Convey("Then UnmarshalBinary should return an error", func() {
					err := m.UnmarshalBinary(b)
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
			p.MaxDCCycle = 16
			Convey("Then MarshalBinary returns an error", func() {
				_, err := p.MarshalBinary()
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Given a MaxDCCycle=254", func() {
			p.MaxDCCycle = 254
			Convey("Then MarshalBinary returns an error", func() {
				_, err := p.MarshalBinary()
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Given A MaxDCCycle=13", func() {
			p.MaxDCCycle = 13
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

func TestDLsettings(t *testing.T) {
	Convey("Given an empty DLsettings", t, func() {
		var s DLsettings
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

		Convey("Given a RX1DRoffset > 7", func() {
			s.RX1DRoffset = 8
			Convey("Then MarshalBinary returns an error", func() {
				_, err := s.MarshalBinary()
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Given RX2DataRate=15 and RX1DRoffset=7", func() {
			s.RX2DataRate = 15
			s.RX1DRoffset = 7
			Convey("Then MarshalBinary returns []byte{127}", func() {
				b, err := s.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{127})
			})
		})

		Convey("Given a slice []byte{127}", func() {
			b := []byte{127}
			Convey("Then UnmarshalBinary returns a DLsettings with RX2DataRate=15 and RX1DRoffset=7", func() {
				err := s.UnmarshalBinary(b)
				So(err, ShouldBeNil)
				So(s, ShouldResemble, DLsettings{RX2DataRate: 15, RX1DRoffset: 7})
			})
		})
	})
}

func TestRX2SetupReqPayload(t *testing.T) {
	Convey("Given an empty RX2SetupReqPayload", t, func() {
		var p RX2SetupReqPayload
		Convey("Then MarshalBinary returns []byte{0, 0, 0, 0}", func() {
			b, err := p.MarshalBinary()
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte{0, 0, 0, 0})
		})

		Convey("Given Frequency > 2^24-1", func() {
			p.Frequency = 16777216 // 2^24
			Convey("Then MarshalBinary returns an error", func() {
				_, err := p.MarshalBinary()
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Given Frequency=262657 and DLsettings(RX2DataRate=11, RX1DRoffset=3)", func() {
			p.Frequency = 262657
			p.DLsettings = DLsettings{RX2DataRate: 11, RX1DRoffset: 3}
			Convey("Then MarshalBinary returns []byte{1, 2, 4, 59}", func() {
				b, err := p.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{59, 1, 2, 4})
			})
		})

		Convey("Given a slice []byte{59, 1, 2, 4}", func() {
			b := []byte{59, 1, 2, 4}
			Convey("Then UnmarshalBinary returns a RX2SetupReqPayload with Frequency=262657 and DLsettings(RX2DataRate=11, RX1DRoffset=3)", func() {
				exp := RX2SetupReqPayload{
					Frequency:  262657,
					DLsettings: DLsettings{RX2DataRate: 11, RX1DRoffset: 3},
				}
				err := p.UnmarshalBinary(b)
				So(err, ShouldBeNil)
				So(p, ShouldResemble, exp)
			})
		})
	})
}

func TestRX2SetupAnsPayload(t *testing.T) {
	Convey("Given an empty RX2SetupAnsPayload", t, func() {
		var p RX2SetupAnsPayload
		Convey("Then MarshalBinary returns []byte{0}", func() {
			b, err := p.MarshalBinary()
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte{0})
		})

		testTable := []struct {
			ChannelACK     bool
			RX2DataRateACK bool
			RX1DRoffsetACK bool
			Bytes          []byte
		}{
			{true, false, false, []byte{1}},
			{false, true, false, []byte{2}},
			{false, false, true, []byte{4}},
			{true, true, true, []byte{7}},
		}

		for _, test := range testTable {
			Convey(fmt.Sprintf("Given ChannelACK=%v, RX2DataRateACK=%v, RX1DRoffsetACK=%v", test.ChannelACK, test.RX2DataRateACK, test.RX1DRoffsetACK), func() {
				p.ChannelACK = test.ChannelACK
				p.RX2DataRateACK = test.RX2DataRateACK
				p.RX1DRoffsetACK = test.RX1DRoffsetACK
				Convey(fmt.Sprintf("Then marshalBinary returns %v", test.Bytes), func() {
					b, err := p.MarshalBinary()
					So(err, ShouldBeNil)
					So(b, ShouldResemble, test.Bytes)
				})
			})

			Convey(fmt.Sprintf("Given slice %v", test.Bytes), func() {
				b := test.Bytes
				Convey(fmt.Sprintf("Then UnmarshalBinary returns a RX2SetupAnsPayload with ChannelACK=%v, RX2DataRateACK=%v, RX1DRoffsetACK=%v", test.ChannelACK, test.RX2DataRateACK, test.RX1DRoffsetACK), func() {
					exp := RX2SetupAnsPayload{
						ChannelACK:     test.ChannelACK,
						RX2DataRateACK: test.RX2DataRateACK,
						RX1DRoffsetACK: test.RX1DRoffsetACK,
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
			p.Freq = 16777216
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

		Convey("Given ChIndex=3, Freq=262657, MaxDR=5, MinDR=10", func() {
			p.ChIndex = 3
			p.Freq = 262657
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
			So(p, ShouldResemble, NewChannelReqPayload{ChIndex: 3, Freq: 262657, MaxDR: 5, MinDR: 10})
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
