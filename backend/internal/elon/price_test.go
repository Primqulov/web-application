package elon

import "testing"

func TestComputePrice(t *testing.T) {
	cases := []struct {
		name             string
		req              upsertReq
		wantType         string
		wantTotal        int64
		wantPer          int64
	}{
		{"per_worker", upsertReq{PricingType: "per_worker", PriceAmount: 100000, WorkersNeeded: 3}, "per_worker", 300000, 100000},
		{"total", upsertReq{PricingType: "total", PriceAmount: 600000, WorkersNeeded: 4}, "total", 600000, 150000},
		{"empty -> negotiable", upsertReq{PricingType: "", PriceAmount: 0, WorkersNeeded: 2}, "negotiable", 0, 0},
		{"empty + amount falls back to per_worker", upsertReq{PricingType: "", PriceAmount: 200000, WorkersNeeded: 2}, "per_worker", 400000, 200000},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			pt, total, per := c.req.computePrice()
			if pt != c.wantType || total != c.wantTotal || per != c.wantPer {
				t.Fatalf("got (%s,%d,%d) want (%s,%d,%d)", pt, total, per, c.wantType, c.wantTotal, c.wantPer)
			}
		})
	}
}
