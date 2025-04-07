package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const sampleXML = `<?xml version="1.0" encoding="windows-1251"?>
<ValCurs Date="05.04.2025" name="Foreign Currency Market">
    <Valute ID="R01010">
        <Nominal>1</Nominal>
        <Name>Australian Dollar</Name>
        <Value>52,5385</Value>
    </Valute>
    <Valute ID="R01020A">
        <Nominal>1</Nominal>
        <Name>Azerbaijan Manat</Name>
        <Value>49,5749</Value>
    </Valute>
</ValCurs>
`

// TestFetchRatesForDate проверяет корректность парсинга XML-ответа.
func TestFetchRatesForDate(t *testing.T) {
	// Создаем тестовый HTTP-сервер, который возвращает sampleXML.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Указываем правильный Content-Type с кодировкой.
		w.Header().Set("Content-Type", "application/xml; charset=windows-1251")
		fmt.Fprint(w, sampleXML)
	}))
	defer ts.Close()

	// Переопределяем базовый URL для тестирования.
	originalAPIURL := apiURL
	apiURL = ts.URL + "?" // добавляем "?" для разделения параметров
	defer func() {
		apiURL = originalAPIURL
	}()

	// Используем фиксированную дату.
	date, err := time.Parse("02/01/2006", "05/04/2025")
	if err != nil {
		t.Fatalf("Не удалось распарсить дату: %v", err)
	}

	rates, err := fetchRatesForDate(date)
	if err != nil {
		t.Fatalf("fetchRatesForDate вернул ошибку: %v", err)
	}

	// Ожидаем получить 2 валюты.
	if len(rates) != 2 {
		t.Fatalf("Ожидалось 2 валюты, получено %d", len(rates))
	}

	// Проверяем первую валюту.
	if rates[0].Currency != "Australian Dollar" {
		t.Errorf("Ожидалось 'Australian Dollar', получено '%s'", rates[0].Currency)
	}
	if rates[0].Rate != 52.5385 {
		t.Errorf("Ожидалось значение 52.5385, получено %f", rates[0].Rate)
	}

	// Проверяем вторую валюту.
	if rates[1].Currency != "Azerbaijan Manat" {
		t.Errorf("Ожидалось 'Azerbaijan Manat', получено '%s'", rates[1].Currency)
	}
	if rates[1].Rate != 49.5749 {
		t.Errorf("Ожидалось значение 49.5749, получено %f", rates[1].Rate)
	}

	// Проверяем дату в структурах.
	for _, r := range rates {
		if r.Date != "05.04.2025" {
			t.Errorf("Ожидалась дата '05.04.2025', получено '%s'", r.Date)
		}
	}
}

// TestCalculationLogic проверяет вычисление максимального, минимального и среднего курса.
func TestCalculationLogic(t *testing.T) {
	// Создаем тестовый набор данных.
	rates := []CurrencyRate{
		{Rate: 50, Currency: "CurrencyA", Date: "01.01.2025"},
		{Rate: 60, Currency: "CurrencyB", Date: "02.01.2025"},
		{Rate: 55, Currency: "CurrencyC", Date: "03.01.2025"},
	}

	// Инициализируем переменные для вычислений.
	maxRate := rates[0].Rate
	minRate := rates[0].Rate
	maxCurrency := rates[0].Currency
	minCurrency := rates[0].Currency
	maxDate := rates[0].Date
	minDate := rates[0].Date
	var sum float64

	for _, r := range rates {
		sum += r.Rate
		if r.Rate > maxRate {
			maxRate = r.Rate
			maxCurrency = r.Currency
			maxDate = r.Date
		}
		if r.Rate < minRate {
			minRate = r.Rate
			minCurrency = r.Currency
			minDate = r.Date
		}
	}
	avgRate := sum / float64(len(rates))

	// Проверяем полученные значения.
	if maxRate != 60 {
		t.Errorf("Максимальный курс: ожидается 60, получено %f", maxRate)
	}
	if minRate != 50 {
		t.Errorf("Минимальный курс: ожидается 50, получено %f", minRate)
	}
	if avgRate != 55 {
		t.Errorf("Средний курс: ожидается 55, получено %f", avgRate)
	}
	if maxCurrency != "CurrencyB" {
		t.Errorf("Ожидалось, что максимальный курс у 'CurrencyB', получено '%s'", maxCurrency)
	}
	if minCurrency != "CurrencyA" {
		t.Errorf("Ожидалось, что минимальный курс у 'CurrencyA', получено '%s'", minCurrency)
	}
	if maxDate != "02.01.2025" {
		t.Errorf("Ожидалась дата '02.01.2025' для максимального курса, получено '%s'", maxDate)
	}
	if minDate != "01.01.2025" {
		t.Errorf("Ожидалась дата '01.01.2025' для минимального курса, получено '%s'", minDate)
	}
}
