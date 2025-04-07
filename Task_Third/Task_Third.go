package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html/charset"
)

var apiURL = "http://www.cbr.ru/scripts/XML_daily_eng.asp?date_req="

// ValCurs представляет корневой элемент XML-ответа Центробанка.
type ValCurs struct {
	XMLName xml.Name `xml:"ValCurs"`
	// Дата, указанная в ответе (формат dd.mm.yyyy)
	Date   string   `xml:"Date,attr"`
	Valute []Valute `xml:"Valute"`
}

// Valute представляет информацию по отдельной валюте.
type Valute struct {
	Nominal string `xml:"Nominal"` // Количество единиц валюты
	Name    string `xml:"Name"`    // Название валюты
	Value   string `xml:"Value"`   // Курс (с запятой в качестве десятичного разделителя)
}

// CurrencyRate хранит рассчитанный курс, название валюты и дату.
type CurrencyRate struct {
	Rate     float64
	Currency string
	Date     string
}

// fetchRatesForDate получает курсы валют для заданной даты.
func fetchRatesForDate(date time.Time) ([]CurrencyRate, error) {
	// Формат даты для запроса: dd/mm/yyyy

	dateStr := date.Format("02/01/2006")
	url := apiURL + dateStr

	// Создаем новый HTTP-запрос с заголовком User-Agent.
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %v", err)
	}
	// Имитируем запрос из браузера.
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "+
		"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка HTTP: статус %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %v", err)
	}

	var data ValCurs
	// Используем xml.Decoder с заданным CharsetReader для поддержки Windows-1251.
	decoder := xml.NewDecoder(bytes.NewReader(body))
	decoder.CharsetReader = charset.NewReaderLabel
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("ошибка разбора XML: %v", err)
	}

	var rates []CurrencyRate
	for _, v := range data.Valute {
		// Преобразуем номинал в число.
		nominal, err := strconv.Atoi(strings.TrimSpace(v.Nominal))
		if err != nil || nominal == 0 {
			continue
		}
		// Заменяем запятую на точку для корректного парсинга.
		valueStr := strings.Replace(strings.TrimSpace(v.Value), ",", ".", 1)
		rateVal, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			continue
		}
		actualRate := rateVal / float64(nominal)
		rates = append(rates, CurrencyRate{
			Rate:     actualRate,
			Currency: v.Name,
			Date:     data.Date, // Дата из XML (формат dd.mm.yyyy)
		})
	}
	return rates, nil
}

func main() {
	fmt.Println("Please Wait, Program is working!")
	var allRates []CurrencyRate
	today := time.Now()

	// Получаем данные за последние 90 дней (включая сегодня).
	for i := 0; i < 90; i++ {
		date := today.AddDate(0, 0, -i)
		rates, err := fetchRatesForDate(date)
		if err != nil {
			fmt.Printf("Ошибка для даты %s: %v\n", date.Format("02/01/2006"), err)
			continue
		}
		allRates = append(allRates, rates...)
	}

	if len(allRates) == 0 {
		fmt.Println("Данные курсов не получены.")
		return
	}

	// Инициализируем максимальное и минимальное значение первыми данными.
	maxRate := allRates[0].Rate
	minRate := allRates[0].Rate
	maxCurrency := allRates[0].Currency
	minCurrency := allRates[0].Currency
	maxDate := allRates[0].Date
	minDate := allRates[0].Date
	var sum float64

	for _, r := range allRates {
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

	avgRate := sum / float64(len(allRates))

	fmt.Printf("Максимальный курс: %.4f руб. за 1 %s (Дата: %s)\n", maxRate, maxCurrency, maxDate)
	fmt.Printf("Минимальный курс: %.4f руб. за 1 %s (Дата: %s)\n", minRate, minCurrency, minDate)
	fmt.Printf("Средний курс по всем валютам за период: %.4f руб.\n", avgRate)
}
