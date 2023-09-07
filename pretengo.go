package main

import (
	"io/ioutil"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"os"
	"os/signal"
	"syscall"
)

type Config struct {
	ListenAddress string `json:"ListenAddress"`
	ListenPort    string `json:"ListenPort"`
	GetServerlist string `json:"GetServerlist"`
	AccessToken   string `json:"access_token"`
	RefreshToken  string `json:"refresh_token"`
	ExpiresIn     string `json:"expires_in"`
	StaticKey	  string `json:"StaticKey"`
	Key  		  string `json:"Key"`
}

var userID string
var password string
var token string
var AppConfig Config

func main() {
	// Read the JSON file
	err := loadConfig("config.json")
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Access the parsed variables
	fmt.Println("AccessToken:", AppConfig.AccessToken)
	fmt.Println("RefreshToken:", AppConfig.RefreshToken)
	fmt.Println("ExpiresIn:", AppConfig.ExpiresIn)
	fmt.Println("StaticKey:", AppConfig.StaticKey)
	fmt.Println("Listen IP :", AppConfig.ListenAddress)
	fmt.Println("Listen Port :", AppConfig.ListenPort)
	if AppConfig.ListenPort != "443" {
		fmt.Println("Not Using Port 443, you need to manually proxy the https into port :", AppConfig.ListenPort)
	}

	// Set up TLS configuration
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // Skip certificate verification (for self-signed certificates)
	}

	// Load the certificate and key files
	certFile := "cert.crt"
	keyFile := "cert.pem"
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatal("Failed to load certificate and key:", err)
	}
	tlsConfig.Certificates = []tls.Certificate{cert}

	// Create an HTTP server with Custom TLS Configuration
	server := &http.Server{
		Addr:      AppConfig.ListenAddress + ":" + AppConfig.ListenPort,
		TLSConfig: tlsConfig,
	}

	// Register request handler
	http.HandleFunc("/v1/api/oauth20/access_token/generate", handleRequest)
	http.HandleFunc("/v1/api/provider/service_token/@me", handleRequest)
	http.HandleFunc("/serverlist.xml", handleRequest)

	// Create a channel to receive the interrupt signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	go func() {
		err := server.ListenAndServeTLS(certFile, keyFile)
		if err != nil {
			log.Fatal("Failed to start server:", err)
			os.Exit(1) // Terminate the program with an error exit status
		}
	}()

	go func() {
		if AppConfig.GetServerlist == "true" {
			err := http.ListenAndServe(":80", nil)
			if err != nil {
			log.Fatal("Failed to start server:", err)
			os.Exit(1) // Terminate the program with an error exit status
		}
		}
	}()

	// Wait for the interrupt signal
	<-interrupt

	// Perform cleanup and shutdown
	fmt.Println("Received interrupt signal. Shutting down...")
	server.Close()       // Close the server
	os.Exit(0)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" && r.URL.Path == "/v1/api/oauth20/access_token/generate" {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		grantType := r.Form.Get("grant_type")
		passwordType := r.Form.Get("password_type")
		userID = r.Form.Get("user_id")
		password = r.Form.Get("password")

		fmt.Println("Received POST request:")
		fmt.Println("Grant Type:", grantType)
		fmt.Println("User ID:", userID)
		fmt.Println("Password:", password)
		fmt.Println("Password Type:", passwordType)
		
		//response := "<OAuth20><access_token><token>1234567890abcdef1234567890abcdef</token><refresh_token>fedcba0987654321fedcba0987654321fedcba12</refresh_token><expires_in>3600</expires_in></access_token></OAuth20>"
		response := fmt.Sprintf("<OAuth20><access_token><token>%s</token><refresh_token>%s</refresh_token><expires_in>%s</expires_in></access_token></OAuth20>", AppConfig.AccessToken, AppConfig.RefreshToken, AppConfig.ExpiresIn)
		fmt.Println("\nSending Response : ", response)

		// Set response headers
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)

		// Write the XML response
		w.Write([]byte(response))
	} else if r.Method == "GET" && r.URL.Path == "/v1/api/provider/service_token/@me"{
		clientID := r.URL.Query().Get("client_id")

		fmt.Println("Received GET request:")
		fmt.Println("Client ID:", clientID)

		if AppConfig.StaticKey != "false" {
			token = AppConfig.Key
		} else {
			// Encode the client ID in base64
	  		token = generateToken(userID, password)
			token = strings.ToUpper(token)
		}

		response := fmt.Sprintf(`(<?xml version="1.0" encoding="UTF-8" standalone="yes"?><service_token><token>%s</token></service_token>)`, token)

		fmt.Println("\nGenerated Token : ", token)
		fmt.Println("Sending Response:", response)
		// Set response headers
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)

		// Write the XML response
		w.Write([]byte(response))
	} else if r.URL.Path == "/serverlist.xml" {
		response, err := http.Get("http://zerulight.cc/serverlist.xml")
		if err != nil {
		log.Fatal("Failed to fetch XML file:", err)
		}
		// Read the response body
		defer response.Body.Close()
		fmt.Println("Getting Serverlist.xml success")
		xmlData, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal("Failed to read XML file:", err)
		}

		// Set response headers
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)

		// Write the XML response
		w.Write([]byte(xmlData))

	} else {
		http.Error(w, "Invalid request method or URL", http.StatusNotFound)
		return
	}
}

func generateToken(text1, text2 string) string {
	combine := text1 + text2

	hash := sha256.Sum256([]byte(combine))
	token := hex.EncodeToString(hash[:32])
	return token
}

func loadConfig(filename string) error {
	// Read the JSON file
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Fail to load Config.json using default config")
		// Fallback: Manually input the configuration values
		AppConfig = Config{
			ListenAddress: "0.0.0.0",
			ListenPort:    "443",
			GetServerlist: "false",
			AccessToken:   "1234567890abcdef1234567890abcdef",
			RefreshToken:  "fedcba0987654321fedcba0987654321fedcba12",
			ExpiresIn:     "3600",
			StaticKey:     "yes",
			Key:    "U0VSVklDRVNFUlZJQ0VTRVJWSUNFU0VSVklDRVNFUlZJQ0VTRVJWSUNFU0VSVklDRVNFUlZJQ0VTRVI=",
		}
		// Return nil to indicate success (since we manually input the values)
		return nil
	}

	// Unmarshal the JSON data into the AppConfig variable
	err = json.Unmarshal(data, &AppConfig)
	if err != nil {
		fmt.Println("Failed to Load config into app")
		return nil
	}

	return nil
}
