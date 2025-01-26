package db

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
    "github.com/joho/godotenv"
  "os"
	mysqlDriver "github.com/go-sql-driver/mysql" // Explicit import for MySQL driver
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	// Path to the CA certificate
	caCertPath := "ca.pem" // Replace with the actual path to your ca.pem file

	// Load the CA certificate
	caCert, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		log.Fatal("Failed to read CA certificate:", err)
	}

	// Create a certificate pool and add the CA certificate
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		log.Fatal("Failed to append CA certificate to the pool")
	}

	// Configure TLS
	tlsConfig := &tls.Config{
		RootCAs: certPool,
	}

	// Register the custom TLS config with the name "custom"
	err = mysqlDriver.RegisterTLSConfig("custom", tlsConfig)
	if err != nil {
		log.Fatal("Failed to register custom TLS config:", err)
	}

	// Define DSN with the custom TLS configuration
	dsn := os.Getenv("DSN")
	if dsn == "" {
        log.Fatal("DSN is not set in the environment")
    }
	log.Println("Database DSN loaded successfully")

	// Connect to the database
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("Database connected!")
}
