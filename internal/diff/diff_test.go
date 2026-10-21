package diff

import "testing"

func TestSeparateHunks(t *testing.T) {
	old := []byte("one\ntwo\nthree\nfour\nfive\nsix\nseven\neight\nnine\nten\neleven\ntwelve\n")
	new := []byte("one\nTWO\nthree\nfour\nfive\nsix\nseven\neight\nnine\nTEN\neleven\ntwelve\n")
	got := string(Diff("source.go", old, "source.go", new))
	want := "diff source.go source.go\n" +
		"--- source.go\n" +
		"+++ source.go\n" +
		"@@ -1,5 +1,5 @@\n" +
		" one\n" +
		"-two\n" +
		"+TWO\n" +
		" three\n" +
		" four\n" +
		" five\n" +
		"@@ -7,6 +7,6 @@\n" +
		" seven\n" +
		" eight\n" +
		" nine\n" +
		"-ten\n" +
		"+TEN\n" +
		" eleven\n" +
		" twelve\n"
	if got != want {
		t.Fatalf("diff output:\n%s\nwant:\n%s", got, want)
	}
}
