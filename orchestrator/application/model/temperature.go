package model

type Temperature struct {
	City                 string  `json:"city"`
	TemperatureCelsius   float64 `json:"temp_C"`
	TemperatureFarenheit float64 `json:"temp_F"`
	TemperatureKelvin    float64 `json:"temp_K"`
}
