package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"math/big"

	"flag"
	quic "github.com/lucas-clemente/quic-go"
	"time"
)

var (
	addr     string
	port     string
	duration int
	size     int
	message  []byte
	quiet    bool
)

func init() {
	flag.StringVar(&port, "s", "", "Server mode. Set address and port to listen on")
	flag.StringVar(&addr, "c", "", "Client mode. Specify address and port to connect to")
	flag.IntVar(&duration, "d", 10, "Send for d seconds")
	flag.IntVar(&size, "p", 1000, "Payload buffer size in bytes")
	flag.BoolVar(&quiet, "q", false, "Quiet, suppress echo output")
	flag.Parse()
}

// We start a server echoing data on the first stream the client opens,
// then connect with a client, send the message, and wait for its receipt.
func main() {
	message = make([]byte, size)
	if port != "" {
		log.Println("Started listening...")
		log.Fatal(echoServer())
	} else if addr != "" {
		err := clientMain()
		if err != nil {
			panic(err)
		}
		log.Println("Send/receive ended, time is up")
	}

}

// Start a server that echos all data on the first stream opened by the client
func echoServer() error {
	listener, err := quic.ListenAddr(port, generateTLSConfig(), nil)
	if err != nil {
		return err
	}
	sess, err := listener.Accept()
	if err != nil {
		return err
	}
	log.Printf("Connect from %s\n", sess.RemoteAddr())
	stream, err := sess.AcceptStream()
	if err != nil {
		panic(err)
	}
	// Echo message
	for err == nil {
		_, err = io.Copy(loggingWriter{stream}, stream)
	}
	return err
}

func clientMain() error {
	session, err := quic.DialAddr(addr, &tls.Config{InsecureSkipVerify: true}, nil)
	if err != nil {
		return err
	}
	log.Printf("Connected to %v\n", addr)
	stream, err := session.OpenStreamSync()
	if err != nil {
		return err
	}

	go func() { sendMessage(stream) }()

	go func() { receiveMessage(stream) }()

	time.Sleep(time.Duration(duration) * time.Second)

	return nil
}

func sendMessage(stream quic.Stream) {
	var err error
	for err == nil {
		_, err = stream.Write([]byte(message))
		if !quiet {
			fmt.Print("s")
		}
	}
	log.Println(err)
}

func receiveMessage(stream quic.Stream) {
	var err error
	buf := make([]byte, len(message))
	for err == nil {
		_, err = io.ReadFull(stream, buf)
		if !quiet {
			fmt.Print("r")
		}
	}
	log.Println(err)
}

// A wrapper for io.Writer that also logs the message.
type loggingWriter struct{ io.Writer }

func (w loggingWriter) Write(b []byte) (int, error) {
	if !quiet {
		fmt.Print(".")
	}
	return w.Writer.Write(b)
}

// Setup a bare-bones TLS config for the server
func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{Certificates: []tls.Certificate{tlsCert}}
}
