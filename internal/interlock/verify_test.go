package interlock_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/wyw14/cry-120/internal/interlock"
	"github.com/wyw14/cry-120/internal/weather"
)

func TestReleasingWeatherHoldPreservesPropellantInhibit(t *testing.T) {
	now := time.Date(2026, 8, 25, 9, 20, 0, 0, time.UTC)
	holds := interlock.NewHoldAggregate()
	weatherHolds := weather.NewHoldService(holds)
	weatherHolds.Raise("区域不安全", now)
	holds.Publish("propellant:fill-7", "区域不安全", now)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if !weatherHolds.Clear("区域不安全") {
			http.Error(writer, "weather hold missing", http.StatusNotFound)
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	request, err := http.NewRequest(http.MethodDelete, server.URL, http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("weather release failed with %d", response.StatusCode)
	}
	if !holds.Has("propellant:fill-7", "区域不安全") {
		t.Fatal("weather release removed propellant-owned inhibit")
	}
	if holds.Has("weather", "区域不安全") {
		t.Fatal("weather-owned hold remained active")
	}
	sources := holds.Sources("区域不安全")
	if len(sources) != 1 || sources[0] != "propellant:fill-7" {
		t.Fatalf("unexpected remaining owners: %+v", sources)
	}
}
