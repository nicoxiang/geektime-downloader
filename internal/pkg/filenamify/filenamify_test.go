package filenamify

import (
	"testing"
)

func TestFilenamify_RemoveEmpty(t *testing.T) {
	want := "ab"
	fileName := Filenamify("a  b")
	if fileName != want {
		t.Fatalf(`want %s, but got %s`, want, fileName)
	}
}

func TestFilenamify_SpecialCharacters(t *testing.T) {
	want := "-"
	fileName := Filenamify(".<>|?")
	if fileName != want {
		t.Fatalf(`want %s, but got %s`, want, fileName)
	}
}

func TestFilenamify_TooLongChinese(t *testing.T) {
	want := "一二三四五六七八九十一二三四五六七八九十一二三四五六七八九十一二三四五六七八九十一二三四五六七八九十一二三四五六七八九十一二三四五六七八九十一二三四五六七八九十一二三四五六七八九十一二三四五六七八九十"
	fileName := Filenamify("一二三四五六七八九十一二三四五六七八九十一二三四五六七八九十一二三四五六七八九十一二三四五六七八九十一二三四五六七八九十一二三四五六七八九十一二三四五六七八九十一二三四五六七八九十一二三四五六七八九十一二三四五六七八九十")
	if fileName != want {
		t.Fatalf(`want %s, but got %s`, want, fileName)
	}
}

func TestFilenamify_TooLongEnglish(t *testing.T) {
	want := "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuv"
	fileName := Filenamify("abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz")
	if fileName != want {
		t.Fatalf(`want %s, but got %s`, want, fileName)
	}
}
