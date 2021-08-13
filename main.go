package main

import (
	"flag"
	"fmt"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"opa-admission-controller/internal"
	"os"
	"time"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()

	port := flag.Int("port", 8443, "port")
	noSSL := flag.Bool("no-ssl", false, "don't use ssl")
	configFile := flag.String("config", "/config/config.yaml", "path to config file")
	flag.Parse()

	yamlFile, err := os.Open(*configFile)
	defer yamlFile.Close()
	if err != nil {
		sugar.Fatal(err)
	}

	byteValue, _ := ioutil.ReadAll(yamlFile)

	mutations := make([]internal.Mutation, 0)
	err = yaml.Unmarshal(byteValue, &mutations)
	if err != nil {
		sugar.Fatalf("Error unmarshalling config yaml %s",err)
	}

	controller := internal.Controller{Sugar: sugar, Mutations: mutations}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", controller.HandleHealth)
	mux.HandleFunc("/mutate", controller.HandleMutate)

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", *port),
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1048576
	}

	sugar.Infof("starting https server on :%d", *port)

	if *noSSL {
		sugar.Fatal(s.ListenAndServe())
	} else {
		sugar.Fatal(s.ListenAndServeTLS("/etc/tls/tls.crt", "/etc/tls/tls.key"))
	}
}
