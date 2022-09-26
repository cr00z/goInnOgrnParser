package innogrn

import (
	"strconv"
)

func CheckRegion(region int64) bool {
	if region == 0 || (region > 79 && region < 83) || region == 84 || region == 85 ||
		region == 88 || region == 90 || (region > 92 && region < 99) {
		return false
	}
	return true
}

func CheckINN(candidate string) bool {
	// https: //ru.wikipedia.org/wiki/Идентификационный_номер_налогоплательщика
	toInt := func(num byte) int64 { return int64(num) - 0x30 }
	checker := func(candidate string, kN []int64, checksum int64) bool {
		var sum int64
		for pos := 0; pos < len(candidate); pos++ {
			sum += toInt(candidate[pos]) * kN[pos]
		}
		return (sum%11%10 == checksum)
	}
	cLen := len(candidate)
	if cLen == 10 || cLen == 12 {
		// КК НН — код налогового органа, который присвоил ИНН
		// TODO: найти перечень налоговых для проверки НН
		if !CheckRegion(toInt(candidate[0])*10 + toInt(candidate[1])) {
			return false
		}
		// контрольное число
		kN := []int64{3, 7, 2, 4, 10, 3, 5, 9, 4, 6, 8}
		if cLen == 10 {
			kN1 := kN[2:]
			return checker(candidate[:cLen-1], kN1, toInt(candidate[cLen-1]))
		}
		if cLen == 12 {
			kN1 := kN[1:]
			return checker(candidate[:cLen-2], kN1, toInt(candidate[cLen-2])) &&
				checker(candidate[:cLen-1], kN, toInt(candidate[cLen-1]))
		}
	}
	return false
}

func CheckOGRN(candidate string) bool {
	// https://ru.wikipedia.org/wiki/Основной_государственный_регистрационный_номер
	// ОГРН:   С ГГ КК НН ХХХХХ Ч
	// ОГРНИП: С ГГ КК ХХХХХХХХХ Ч
	toInt := func(num byte) int64 { return int64(num) - 0x30 }
	cLen := len(candidate)
	if cLen == 13 || cLen == 15 {
		// С (1-й знак) — признак отнесения государственного регистрационного номера записи
		if (cLen == 13 && candidate[0] != '1' && candidate[0] != '5') || // ОГРН
			(cLen == 15 && candidate[0] != '3') { // ОГРНИП
			return false
		}
		// ГГ (со 2-го по 3-й знак) — две последние цифры года внесения записи в госреестр
		// TODO: работает с 00 по 29 год
		if candidate[1] > '2' {
			return false
		}
		// КК (4-й и 5-й знаки) - кодовое обозначение субъекта Российской Федерации,
		// установленное ФНС
		// https://www.buxprofi.ru/spravochnik/kody-regionov-ili-subektov-rf-dlja-nalogovoj
		if !CheckRegion(toInt(candidate[3])*10 + toInt(candidate[4])) {
			return false
		}
		// TODO: найти перечень налоговых для проверки НН
		// Ч - контрольное число
		num, _ := strconv.ParseInt(candidate[:cLen-1], 10, 64)
		return num%int64(cLen-2)%10 == toInt(candidate[cLen-1])
	}
	return false
}
