package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	netatmo "github.com/tipok/netatmo_exporter/netatmo-api"
)

const tokenFile = "netatmo_token.json"

func saveToken(token string) error {
	return os.WriteFile(tokenFile, []byte(token), 0600)
}

func loadToken() (string, error) {
	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func main() {
	var clientID, clientSecret, listen string
	flag.StringVar(&clientID, "client-id", "", "Netatmo API client ID")
	flag.StringVar(&clientSecret, "client-secret", "", "Netatmo API client secret")
	flag.StringVar(&listen, "listen", ":2112", "Address to listen on")
	flag.Parse()

	if clientID == "" || clientSecret == "" {
		log.Fatal("Netatmo API client ID and secret must be provided.")
	}

	refreshToken, err := loadToken()
	if err != nil || refreshToken == "" {
		log.Fatalf("Could not load refresh token from %s: %v", tokenFile, err)
	}

	config := &netatmo.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RefreshToken: refreshToken,
		Scopes:       []string{netatmo.ReadStation, netatmo.ReadThermostat},
	}

	client, err := netatmo.NewClient(context.Background(), config)
	if err != nil {
		log.Fatal(err)
	}

	// Salvataggio automatico del nuovo token aggiornato
	if client.Config.RefreshToken != refreshToken {
		if err := saveToken(client.Config.RefreshToken); err != nil {
			log.Printf("Error saving new token: %v", err)
		} else {
			log.Println("New refresh token saved.")
		}
	}

	prometheus.MustRegister(version.NewCollector("netatmo_exporter"))
	collector := newCollector(client)
	prometheus.MustRegister(collector)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(sig)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{Addr: listen, Handler: mux}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	<-sig

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
}
