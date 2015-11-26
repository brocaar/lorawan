package lorawan

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

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
		Convey("ChMaskACK, DataRateACK and PowerACK should be false", func() {
			So(p.ChMaskACK(), ShouldBeFalse)
			So(p.DataRateACK(), ShouldBeFalse)
			So(p.PowerACK(), ShouldBeFalse)
		})
	})

	Convey("Given I use NewLinkADRAnsPayload to create a new LinkADRAnsPayload", t, func() {
		Convey("Given I call NewLinkADRAnsPayload(true, false, false)", func() {
			p := NewLinkADRAnsPayload(true, false, false)
			Convey("ChMaskACK should be true", func() {
				So(p.ChMaskACK(), ShouldBeTrue)
				So(p.DataRateACK(), ShouldBeFalse)
				So(p.PowerACK(), ShouldBeFalse)
			})
		})

		Convey("Given I call NewLinkADRAnsPayload(true, true, false)", func() {
			p := NewLinkADRAnsPayload(true, true, false)
			Convey("ChMaskACK and DataRateACK should be true", func() {
				So(p.ChMaskACK(), ShouldBeTrue)
				So(p.DataRateACK(), ShouldBeTrue)
				So(p.PowerACK(), ShouldBeFalse)
			})
		})

		Convey("Given I call NewLinkADRAnsPayload(true, true, true)", func() {
			p := NewLinkADRAnsPayload(true, true, true)
			Convey("ChMaskACK DataRateACK and PowerACK should be true", func() {
				So(p.ChMaskACK(), ShouldBeTrue)
				So(p.DataRateACK(), ShouldBeTrue)
				So(p.PowerACK(), ShouldBeTrue)
			})
		})
	})
}

func TestDutyCycleReqPayload(t *testing.T) {
	Convey("Given I use NewDutyCycleReqPayload to create a new DutyCycleReqPayload", t, func() {
		Convey("A value > 15 should return an error", func() {
			_, err := NewDutyCycleReqPayload(16)
			So(err, ShouldNotBeNil)
		})
		Convey("A value < 255 should return an error", func() {
			_, err := NewDutyCycleReqPayload(254)
			So(err, ShouldNotBeNil)
		})
		Convey("A value < 15 should not return an error", func() {
			p, err := NewDutyCycleReqPayload(14)
			So(err, ShouldBeNil)
			So(p, ShouldEqual, DutyCycleReqPayload(14))
		})
	})
}

func TestDLsettings(t *testing.T) {
	Convey("Given an empty DLsettings", t, func() {
		var s DLsettings
		Convey("RX2DataRate and RX1DRoffset should both be 0", func() {
			So(s.RX1DRoffset(), ShouldEqual, 0)
			So(s.RX2DataRate(), ShouldEqual, 0)
		})

	})

	Convey("Given I use NewDLsettings to create a new NewDLsettings", t, func() {
		Convey("When calling NewDLsettings(15, 7)", func() {
			s, err := NewDLsettings(15, 7)
			So(err, ShouldBeNil)

			Convey("Then RX2DataRate should be 15", func() {
				So(s.RX2DataRate(), ShouldEqual, 15)
			})
			Convey("Then RX1DRoffset should be 7", func() {
				So(s.RX1DRoffset(), ShouldEqual, 7)
			})
		})

		Convey("A RX2DataRate > 15 should return an error", func() {
			_, err := NewDLsettings(16, 0)
			So(err, ShouldNotBeNil)
		})
		Convey("A RX1DRoffset > 7 should return an error", func() {
			_, err := NewDLsettings(0, 8)
			So(err, ShouldNotBeNil)
		})
	})

}

func TestFrequency(t *testing.T) {
	Convey("Given an empty Frequency", t, func() {
		var f Frequency
		Convey("It's uint32 representation should be 0", func() {
			So(f.Uint32(), ShouldEqual, 0)
		})
	})

	Convey("Given I use NewFrequency to create a new Frequency", t, func() {
		Convey("When calling NewFrequency(2^24-1)", func() {
			f, err := NewFrequency(2 ^ 24 - 1)
			So(err, ShouldBeNil)
			Convey("Then it's uint32 representation should equal 2^24-1", func() {
				So(f.Uint32(), ShouldEqual, 2^24-1)
			})
		})
		Convey("A frequency >= 2^24 returns an error", func() {
			_, err := NewFrequency(2 ^ 24)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestRX2SetupAnsPayload(t *testing.T) {
	Convey("Given an empty RX2SetupAnsPayload", t, func() {
		var p RX2SetupAnsPayload
		Convey("Then ChannelACK, RX2DataRateACK and RX1DRoffsetACK are false", func() {
			So(p.ChannelACK(), ShouldBeFalse)
			So(p.RX2DataRateACK(), ShouldBeFalse)
			So(p.RX1DRoffsetACK(), ShouldBeFalse)
		})
	})

	Convey("Given I use NewRX2SetupAnsPayload to create a new RX2SetupAnsPayload", t, func() {
		testTable := [][3]bool{
			{false, false, false},
			{true, false, false},
			{false, true, false},
			{false, false, true},
			{true, true, true},
		}

		for _, test := range testTable {
			Convey(fmt.Sprintf("When calling NewRX2SetupAnsPayload(%v, %v, %v)", test[0], test[1], test[2]), func() {
				p := NewRX2SetupAnsPayload(test[0], test[1], test[2])
				Convey(fmt.Sprintf("Then ChannelACK=%v, RX2DataRateACK=%v and RX1DRoffsetACK=%v", test[0], test[1], test[2]), func() {
					So(p.ChannelACK(), ShouldEqual, test[0])
					So(p.RX2DataRateACK(), ShouldEqual, test[1])
					So(p.RX1DRoffsetACK(), ShouldEqual, test[2])
				})
			})
		}
	})
}
