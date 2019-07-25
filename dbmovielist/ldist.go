package dbmovielist

import (
	"strings"
	"unicode/utf8"
)

func LevenshteinDist(s1 string, s2 string) int {
	uniS1 := unicodeFromUtf8(replaceAlias(s1))
	uniS2 := unicodeFromUtf8(replaceAlias(s2))
	len1 := len(uniS1)
	len2 := len(uniS2)

	table := make([][]int, len1+1)
	for i := range table {
		table[i] = make([]int, len2+1)
	}

	for i := 0; i <= len1; i++ {
		table[i][0] = i
	}
	for i := 0; i <= len2; i++ {
		table[0][i] = i
	}

	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			var cost int
			if uniS1[i-1] == uniS2[j-1] {
				cost = 0
			} else {
				cost = 1
			}
			del := table[i-1][j] + 1
			ins := table[i][j-1] + 1
			sub := table[i-1][j-1] + cost
			table[i][j] = min(del, ins, sub)
		}
	}
	return table[len1][len2]
}

func unicodeFromUtf8(s string) []uint16 {
	length := utf8.RuneCountInString(s)
	uniS := make([]uint16, length)
	var i int // while using range on string, index is count by byte and char is count by rune
	for _, char := range s {
		uniS[i] = uint16(char)
		i++
	}
	return uniS
}

func min(first int, rest ...int) int {
	min := first
	for _, n := range rest {
		if n < min {
			min = n
		}
	}
	return min
}

func replaceAlias(s string) string {
	roman := strings.NewReplacer("Ⅰ", "1", "Ⅱ", "2", "Ⅲ", "3", "Ⅳ", "4", "Ⅴ", "5", "Ⅵ", "6", "Ⅶ", "7", "Ⅷ", "8", "Ⅸ", "9", "Ⅹ", "10")
	numberFull := strings.NewReplacer("１", "1", "２", "2", "３", "3", "４", "4", "５", "5", "６", "6", "７", "7", "８", "8", "９", "9", "０", "0")
	lowerLetterFull := strings.NewReplacer("ａ", "a", "ｂ", "b", "ｃ", "c", "ｄ", "d", "ｅ", "e", "ｆ", "f", "ｇ", "g", "ｈ", "h", "ｉ", "i", "ｊ", "j", "ｋ", "k", "ｌ", "l", "ｍ", "m", "ｎ", "n", "ｏ", "o", "ｐ", "p", "ｑ", "q", "ｒ", "r", "ｓ", "s", "ｔ", "t", "ｕ", "u", "ｖ", "v", "ｗ", "w", "ｘ", "x", "ｙ", "y", "ｚ", "z")
	upperLetterFull := strings.NewReplacer("Ａ", "A", "Ｂ", "B", "Ｃ", "C", "Ｄ", "D", "Ｅ", "E", "Ｆ", "F", "Ｇ", "G", "Ｈ", "H", "Ｉ", "I", "Ｊ", "J", "Ｋ", "K", "Ｌ", "L", "Ｍ", "M", "Ｎ", "N", "Ｏ", "O", "Ｐ", "P", "Ｑ", "Q", "Ｒ", "R", "Ｓ", "S", "Ｔ", "T", "Ｕ", "U", "Ｖ", "V", "Ｗ", "W", "Ｘ", "X", "Ｙ", "Y", "Ｚ", "Z")
	symbolFull := strings.NewReplacer("～", "~", "！", "!", "＠", "@", "＃", "#", "＄", "$", "％", "%", "＾", "^", "＆", "&", "＊", "*", "（", "(", "）", ")", "＿", "_", "－", "-", "＝", "=", "＋", "+", "＼", "\\", "［", "[", "］", "]", "｛", "{", "｝", "}", "：", ":", "；", ";", "＇", "'", "＂", "\"", "＜", "<", "＞", ">", "？", "?", "，", ",", "．", ".", "／", "/")
	chineseSymbol := strings.NewReplacer("『", "[", "』", "]", "〖", "[", "〗", "]", "〈", "<", "〉", ">", "“", "\"", "。", ".", "、", ",")
	return chineseSymbol.Replace(symbolFull.Replace(upperLetterFull.Replace(lowerLetterFull.Replace(numberFull.Replace(roman.Replace(s))))))
}
