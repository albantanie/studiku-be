package repositories

import "testing"

func ptr(value float64) *float64 { return &value }

func TestComputeNGain(t *testing.T) {
	cases := []struct {
		name     string
		pre      *float64
		post     *float64
		wantG    float64
		wantCat  string
		wantStat string
		wantOK   bool
	}{
		{"kenaikan tinggi", ptr(40), ptr(85), 0.75, "Tinggi", NGainStatusOK, true},
		{"kenaikan sedang", ptr(40), ptr(70), 0.5, "Sedang", NGainStatusOK, true},
		{"kenaikan rendah", ptr(40), ptr(50), 0.1667, "Rendah", NGainStatusOK, true},
		{"batas bawah tinggi", ptr(0), ptr(70), 0.7, "Tinggi", NGainStatusOK, true},
		{"batas bawah sedang", ptr(0), ptr(30), 0.3, "Sedang", NGainStatusOK, true},
		{"nilai turun", ptr(80), ptr(60), -1, "Rendah", NGainStatusRegression, true},
		{"pretest maksimum", ptr(100), ptr(100), 0, "Tidak Terdefinisi", NGainStatusCeiling, false},
		{"post-test kosong", ptr(40), nil, 0, "Belum Lengkap", NGainStatusIncomplete, false},
		{"pretest kosong", nil, ptr(80), 0, "Belum Lengkap", NGainStatusIncomplete, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ComputeNGain(tc.pre, tc.post, 100)
			if got.NGain != tc.wantG {
				t.Errorf("NGain = %v, mau %v", got.NGain, tc.wantG)
			}
			if got.Category != tc.wantCat {
				t.Errorf("Category = %q, mau %q", got.Category, tc.wantCat)
			}
			if got.Status != tc.wantStat {
				t.Errorf("Status = %q, mau %q", got.Status, tc.wantStat)
			}
			if got.Valid != tc.wantOK {
				t.Errorf("Valid = %v, mau %v", got.Valid, tc.wantOK)
			}
		})
	}
}

func TestComputeNGainEffectiveness(t *testing.T) {
	cases := []struct {
		pre, post float64
		want      string
	}{
		{20, 85, "Efektif"},        // 0,8125 -> 81,25%
		{40, 80, "Cukup Efektif"},  // 0,6667 -> 66,67%
		{50, 75, "Kurang Efektif"}, // 0,5    -> 50%
		{50, 65, "Tidak Efektif"},  // 0,3    -> 30%
	}
	for _, tc := range cases {
		got := ComputeNGain(&tc.pre, &tc.post, 100)
		if got.Effectiveness != tc.want {
			t.Errorf("pre=%v post=%v: Effectiveness = %q (%.2f%%), mau %q", tc.pre, tc.post, got.Effectiveness, got.Percent, tc.want)
		}
	}
}

func TestSummarizeNGainValuesExcludesCeilingAndIncomplete(t *testing.T) {
	values := []NGainValue{
		ComputeNGain(ptr(40), ptr(85), 100),   // 0,75 Tinggi
		ComputeNGain(ptr(40), ptr(70), 100),   // 0,50 Sedang
		ComputeNGain(ptr(100), ptr(100), 100), // ceiling, dikeluarkan
		ComputeNGain(ptr(40), nil, 100),       // belum lengkap, dikeluarkan
	}
	summary := SummarizeNGainValues(values, 100)

	if summary.StudentCount != 4 {
		t.Errorf("StudentCount = %d, mau 4", summary.StudentCount)
	}
	if summary.ComputedCount != 2 {
		t.Errorf("ComputedCount = %d, mau 2", summary.ComputedCount)
	}
	if summary.CeilingCount != 1 {
		t.Errorf("CeilingCount = %d, mau 1", summary.CeilingCount)
	}
	if summary.IncompleteCnt != 1 {
		t.Errorf("IncompleteCount = %d, mau 1", summary.IncompleteCnt)
	}
	// Rata-rata N-Gain hanya dari dua mahasiswa yang terhitung: (0,75 + 0,50) / 2.
	if summary.AverageNGain != 0.625 {
		t.Errorf("AverageNGain = %v, mau 0.625", summary.AverageNGain)
	}
	if summary.Category != "Sedang" {
		t.Errorf("Category = %q, mau \"Sedang\"", summary.Category)
	}

	// Sebaran dihitung terhadap ComputedCount, bukan seluruh mahasiswa.
	for _, bucket := range summary.Distribution {
		switch bucket.Category {
		case "Tinggi", "Sedang":
			if bucket.Count != 1 || bucket.Percent != 50 {
				t.Errorf("%s: count=%d percent=%v, mau 1 dan 50", bucket.Category, bucket.Count, bucket.Percent)
			}
		case "Rendah":
			if bucket.Count != 0 {
				t.Errorf("Rendah: count=%d, mau 0", bucket.Count)
			}
		}
	}
}
