package release

import "testing"

func TestParseSemVer(t *testing.T) {
	cases := []struct {
		in      string
		want    SemVer
		wantErr bool
	}{
		{"v0.4.1", SemVer{0, 4, 1}, false},
		{"0.4.1", SemVer{0, 4, 1}, false},
		{"v1.20.300", SemVer{1, 20, 300}, false},
		{"", SemVer{}, true},
		{"v1.2", SemVer{}, true},
		{"v1.2.3.4", SemVer{}, true},
		{"v1.x.3", SemVer{}, true},
		{"v-1.2.3", SemVer{}, true},
		{"0.0.1-dev", SemVer{}, true},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got, err := ParseSemVer(c.in)
			if (err != nil) != c.wantErr {
				t.Fatalf("ParseSemVer(%q) err=%v wantErr=%v", c.in, err, c.wantErr)
			}
			if err == nil && got != c.want {
				t.Errorf("ParseSemVer(%q) = %+v, want %+v", c.in, got, c.want)
			}
		})
	}
}

func TestSemVer_Compare(t *testing.T) {
	cases := []struct {
		a, b SemVer
		want int
	}{
		{SemVer{0, 4, 1}, SemVer{0, 4, 1}, 0},
		{SemVer{0, 4, 1}, SemVer{0, 4, 2}, -1},
		{SemVer{0, 4, 2}, SemVer{0, 4, 1}, 1},
		{SemVer{0, 4, 0}, SemVer{0, 5, 0}, -1},
		{SemVer{0, 5, 0}, SemVer{0, 4, 99}, 1},
		{SemVer{1, 0, 0}, SemVer{0, 99, 99}, 1},
		{SemVer{0, 4, 10}, SemVer{0, 4, 2}, 1}, // numeric, not lexicographic
	}
	for _, c := range cases {
		got := c.a.Compare(c.b)
		if got != c.want {
			t.Errorf("(%+v).Compare(%+v) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestIsDevBuild(t *testing.T) {
	devLikes := []string{
		"0.0.1-dev",                            // sentinel
		"v0.4.2-0.20260524150735-4bd22c5caa45", // Go pseudo-version
		"v0.4.2+dirty",                         // local dirty build
		"v0.4.2-0.20260524150735-abc123+dirty", // pseudo + dirty
	}
	for _, s := range devLikes {
		if !IsDevBuild(s) {
			t.Errorf("expected %q to be classified as dev build", s)
		}
	}
	for _, s := range []string{"v0.4.1", "0.4.1", "", "dev", "v0.0.1"} {
		if IsDevBuild(s) {
			t.Errorf("did not expect %q to be classified as dev build", s)
		}
	}
}
