package main

import (
	"authorsbooks/metadata/internal/controller/metadata"
	handler "authorsbooks/metadata/internal/handler/http"
	"authorsbooks/metadata/internal/repository/memory"
	"authorsbooks/pkg/registry"
	"authorsbooks/pkg/registry/consul"
	"context" //para reg en el consul
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const serviceName = "metadata" //nombre del servicio

func getenv(k, def string) string { //función para leer las variables de entorno con valor por defecto
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	var port int
	flag.IntVar(&port, "port", 8082, "API handler port") //para que el programa sepa en qué puerto debe como arranca y que se pueda cambiar cuando lo ejecuto sin que tenga que cambiar le codigo completo
	flag.Parse()                                         //para que por si quiero poner metadata -port=8087 pueda hacerlo si quiero

	// Dentro de Docker: CONSUL_ADDR=consul:8500,
	// SERVICE_HOST=metadata
	// Local: CONSUL_ADDR=localhost:8500,
	// SERVICE_HOST=localhost
	consulAddr := getenv("CONSUL_ADDR", "localhost:8500") //depende del entorno
	serviceHost := getenv("SERVICE_HOST", "localhost")    // host con el que el consul va a decir el servicio
	hostPort := fmt.Sprintf("%s:%d", serviceHost, port)   //host + puerto

	ctx := context.Background()                //contexto base, sin deadline
	reg, err := consul.NewRegistry(consulAddr) //crea el cliente de consul aputnadno al consuladdr
	if err != nil {
		log.Fatalf("consul.NewRegistry error: %v", err) //muere literal muerte si falla
	}

	instanceID := registry.GenerateInstanceID(serviceName) //genera un ID único para la instancia
	if err := reg.Register(ctx, instanceID, serviceName, hostPort); err != nil {
		log.Fatalf("registry.Register error: %v", err)
	}
	log.Printf("Registered %s as %s at %s (Consul: %s)", serviceName, instanceID, hostPort, consulAddr)

	// Reporte de salud casa seg
	go func() { //avisa al consul que el servicio it's alive
		for {
			if err := reg.ReportHealthyState(instanceID, serviceName); err != nil {
				log.Printf("ReportHealthyState error: %v", err)
			}
			time.Sleep(1 * time.Second)
		}
	}()

	defer func() { //al salit del programa se desregistra del consul
		_ = reg.Deregister(ctx, instanceID, serviceName)
		log.Printf("Deregistered %s", instanceID)
	}()

	repo := memory.New()       //crea el repo en memeotia para metadata
	ctrl := metadata.New(repo) //crea el controller usando el repo
	h := handler.New(ctrl)     //crea los handlers usando controller

	mux := http.NewServeMux() //mapa mapa mapa rutas

	// Un solo handler para el path /metadata diferenciamos por método
	mux.HandleFunc("/metadata", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetMetadata(w, r) // espera ?id=<ID>
		case http.MethodPost:
			h.PostMetadata(w, r) // body JSON {id,title,description}
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	//servudor http
	addr := fmt.Sprintf(":%d", port) //donde se va a escuhcar
	log.Printf("%s listening on %s", serviceName, addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("ListenAndServe error: %v", err)
	}
}
