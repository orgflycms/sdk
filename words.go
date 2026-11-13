package sdk

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// InWords — сумма прописью для счетов и актов: 123456 → «Одна тысяча двести
// тридцать четыре рубля 56 копеек». Отрицательная — со словом «минус».
//
// Функция в sdk, а не в каждом плагине документов: склонения рубля и
// копейки, «одна/две тысячи», «одиннадцать — девятнадцать» — классическое
// место ошибок бухгалтерского документа, а у кода плагина в магазине тестов
// нет. Ядро отдаёт её плагину нативной, её проверяет обычный тест ядра.
func InWords(kopecks int64) string {
	// Модуль — в uint64: у самого малого int64 нет положительной пары.
	abs := uint64(kopecks)
	if kopecks < 0 {
		abs = -abs
	}
	rub, kop := abs/100, abs%100

	var b strings.Builder
	if kopecks < 0 {
		b.WriteString("минус ")
	}
	if rub == 0 {
		b.WriteString("ноль")
	} else {
		b.WriteString(number(rub))
	}
	fmt.Fprintf(&b, " %s %02d %s", plural(rub, "рубль", "рубля", "рублей"),
		kop, plural(kop, "копейка", "копейки", "копеек"))

	s := b.String()
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[n:]
}

// Разряды по тысяче: у тысячи женский род («одна тысяча», «две тысячи»), у
// остальных мужской. Квадриллиона хватает на весь int64 копеек.
var scales = [...]struct {
	one, few, many string
	fem            bool
}{
	{},
	{"тысяча", "тысячи", "тысяч", true},
	{"миллион", "миллиона", "миллионов", false},
	{"миллиард", "миллиарда", "миллиардов", false},
	{"триллион", "триллиона", "триллионов", false},
	{"квадриллион", "квадриллиона", "квадриллионов", false},
}

var (
	hundreds = [...]string{"", "сто", "двести", "триста", "четыреста", "пятьсот", "шестьсот", "семьсот", "восемьсот", "девятьсот"}
	tens     = [...]string{"", "", "двадцать", "тридцать", "сорок", "пятьдесят", "шестьдесят", "семьдесят", "восемьдесят", "девяносто"}
	teens    = [...]string{"десять", "одиннадцать", "двенадцать", "тринадцать", "четырнадцать", "пятнадцать", "шестнадцать", "семнадцать", "восемнадцать", "девятнадцать"}
	units    = [...]string{"", "один", "два", "три", "четыре", "пять", "шесть", "семь", "восемь", "девять"}
	unitsFem = [...]string{"", "одна", "две", "три", "четыре", "пять", "шесть", "семь", "восемь", "девять"}
)

// number — целое больше нуля словами: 1234 → «одна тысяча двести тридцать четыре».
func number(n uint64) string {
	var parts []string
	for i := 0; n > 0; i++ {
		g := n % 1000
		n /= 1000
		if g == 0 {
			continue
		}
		words := triple(g, scales[i].fem)
		if i > 0 {
			words = append(words, plural(g, scales[i].one, scales[i].few, scales[i].many))
		}
		parts = append(words, parts...)
	}
	return strings.Join(parts, " ")
}

// triple — от 1 до 999 словами.
func triple(g uint64, fem bool) []string {
	var w []string
	if h := g / 100; h > 0 {
		w = append(w, hundreds[h])
	}
	t, u := g/10%10, g%10
	if t == 1 {
		return append(w, teens[u])
	}
	if t > 1 {
		w = append(w, tens[t])
	}
	if u > 0 {
		if fem {
			w = append(w, unitsFem[u])
		} else {
			w = append(w, units[u])
		}
	}
	return w
}

// plural — форма слова при числе: 1 рубль, 2 рубля, 5 рублей, 11 рублей,
// 21 рубль.
func plural(n uint64, one, few, many string) string {
	n %= 100
	if n >= 11 && n <= 19 {
		return many
	}
	switch n % 10 {
	case 1:
		return one
	case 2, 3, 4:
		return few
	}
	return many
}
