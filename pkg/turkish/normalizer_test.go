package turkish

import "testing"

// ─── NormalizeSQLLiterals ───────────────────────────────────────────────────

func TestNormalizeSQLLiterals_SimpleWhere(t *testing.T) {
	in := "SELECT * FROM MUSTERI WHERE ADI = 'Hüseyin'"
	want := "SELECT * FROM MUSTERI WHERE ADI = 'HUSEYIN'"
	if got := NormalizeSQLLiterals(in); got != want {
		t.Errorf("NormalizeSQLLiterals(%q)\n got  %q\n want %q", in, got, want)
	}
}

func TestNormalizeSQLLiterals_Istanbul(t *testing.T) {
	in := "SELECT * FROM SEHIR WHERE AD = 'İstanbul'"
	want := "SELECT * FROM SEHIR WHERE AD = 'ISTANBUL'"
	if got := NormalizeSQLLiterals(in); got != want {
		t.Errorf("NormalizeSQLLiterals(%q)\n got  %q\n want %q", in, got, want)
	}
}

func TestNormalizeSQLLiterals_Like(t *testing.T) {
	in := "SELECT * FROM URUN WHERE ACIKLAMA LIKE '%şeker%'"
	want := "SELECT * FROM URUN WHERE ACIKLAMA LIKE '%SEKER%'"
	if got := NormalizeSQLLiterals(in); got != want {
		t.Errorf("NormalizeSQLLiterals(%q)\n got  %q\n want %q", in, got, want)
	}
}

func TestNormalizeSQLLiterals_MultipleLiterals(t *testing.T) {
	in := "SELECT * FROM SEHIR WHERE AD = 'İstanbul' OR AD = 'Çanakkale'"
	want := "SELECT * FROM SEHIR WHERE AD = 'ISTANBUL' OR AD = 'CANAKKALE'"
	if got := NormalizeSQLLiterals(in); got != want {
		t.Errorf("NormalizeSQLLiterals(%q)\n got  %q\n want %q", in, got, want)
	}
}

func TestNormalizeSQLLiterals_PreservesKeywords(t *testing.T) {
	// SQL keywords and identifiers must NOT be modified
	in := "SELECT adi, soyadi FROM musteri WHERE tip = 'bireysel'"
	want := "SELECT adi, soyadi FROM musteri WHERE tip = 'BIREYSEL'"
	if got := NormalizeSQLLiterals(in); got != want {
		t.Errorf("NormalizeSQLLiterals(%q)\n got  %q\n want %q", in, got, want)
	}
}

func TestNormalizeSQLLiterals_AllTurkishChars(t *testing.T) {
	in := "SELECT * FROM T WHERE V = 'İışşĞğÜüÖöÇç'"
	got := NormalizeSQLLiterals(in)
	for _, r := range []rune{'İ', 'ı', 'Ş', 'ş', 'Ğ', 'ğ', 'Ü', 'ü', 'Ö', 'ö', 'Ç', 'ç'} {
		if containsRune(got, r) {
			t.Errorf("NormalizeSQLLiterals result still contains Turkish char %c: %q", r, got)
		}
	}
}

func TestNormalizeSQLLiterals_EscapedQuotes(t *testing.T) {
	in := "SELECT * FROM T WHERE V = 'O''nun adi'"
	got := NormalizeSQLLiterals(in)
	if got == "" {
		t.Error("NormalizeSQLLiterals returned empty string for escaped-quote input")
	}
}

func TestNormalizeSQLLiterals_NoLiterals(t *testing.T) {
	in := "SELECT COUNT(*) FROM MUSTERI"
	if got := NormalizeSQLLiterals(in); got != in {
		t.Errorf("NormalizeSQLLiterals should be a no-op when no literals: got %q", got)
	}
}

func TestNormalizeSQLLiterals_EmptyLiteral(t *testing.T) {
	in := "SELECT * FROM T WHERE V = ''"
	want := "SELECT * FROM T WHERE V = ''"
	if got := NormalizeSQLLiterals(in); got != want {
		t.Errorf("NormalizeSQLLiterals(%q) = %q, want %q", in, got, want)
	}
}

// ─── ToASCIIUpper ──────────────────────────────────────────────────────────

var toASCIIUpperTests = []struct {
	in   string
	want string
}{
	{"Hüseyin", "HUSEYIN"},
	{"Istanbul", "ISTANBUL"},
	{"cicek", "CICEK"},
	{"SEKER", "SEKER"},
	{"g", "G"},
	{"I", "I"},
	{"Turkiye", "TURKIYE"},
	{"ANKARA", "ANKARA"},
	{"", ""},
}

func TestToASCIIUpper(t *testing.T) {
	for _, tc := range toASCIIUpperTests {
		if got := ToASCIIUpper(tc.in); got != tc.want {
			t.Errorf("ToASCIIUpper(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestToASCIIUpper_TurkishChars(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Hüseyin", "HUSEYIN"},
		{"çiçek", "CICEK"},
		{"SAĞLIK", "SAGLIK"},
		{"ışık", "ISIK"},
	}
	for _, tc := range cases {
		if got := ToASCIIUpper(tc.in); got != tc.want {
			t.Errorf("ToASCIIUpper(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// ─── FixMojibake ───────────────────────────────────────────────────────────

var fixMojibakeTests = []struct {
	in   string
	want string
}{
	{"YSTANBUL", "YSTANBUL"},
	{"SAGLIK", "SAGLIK"},
	{"ANKARA", "ANKARA"},
	{"", ""},
}

func TestFixMojibake_NoChange(t *testing.T) {
	for _, tc := range fixMojibakeTests {
		if got := FixMojibake(tc.in); got != tc.want {
			t.Errorf("FixMojibake(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestFixMojibake_MojibakeChars(t *testing.T) {
	// Ð (0xD0 in Win1252) should become Ğ
	if got := FixMojibake("SAÐ"); got != "SAĞ" {
		t.Errorf("FixMojibake(SAÐ) = %q, want SAĞ", got)
	}
	// Þ (0xDE in Win1252) should become Ş
	if got := FixMojibake("ÞEKİL"); got != "ŞEKİL" {
		t.Errorf("FixMojibake(ÞEKİL) = %q, want ŞEKİL", got)
	}
}

// ─── FixResultValue ────────────────────────────────────────────────────────

func TestFixResultValue_String(t *testing.T) {
	v := FixResultValue("SAÐLIK")
	if got, ok := v.(string); !ok || got != "SAĞLIK" {
		t.Errorf("FixResultValue(string) = %v, want SAĞLIK", v)
	}
}

func TestFixResultValue_Bytes(t *testing.T) {
	v := FixResultValue([]byte("SAÐLIK"))
	if got, ok := v.(string); !ok || got != "SAĞLIK" {
		t.Errorf("FixResultValue([]byte) = %v, want SAĞLIK", v)
	}
}

func TestFixResultValue_Int(t *testing.T) {
	v := FixResultValue(42)
	if got, ok := v.(int); !ok || got != 42 {
		t.Errorf("FixResultValue(int) = %v, want 42", v)
	}
}

func TestFixResultValue_Nil(t *testing.T) {
	v := FixResultValue(nil)
	if v != nil {
		t.Errorf("FixResultValue(nil) = %v, want nil", v)
	}
}

func TestFixResultValue_NoChange(t *testing.T) {
	original := "ANKARA"
	v := FixResultValue(original)
	got, ok := v.(string)
	if !ok || got != original {
		t.Errorf("FixResultValue(no-mojibake string) = %v, want %q", v, original)
	}
}

// ─── helpers ───────────────────────────────────────────────────────────────

func containsRune(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}
	return false
}
