package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	addr := getenv("ADDR", ":8081") //puerto donde escuchará este servicio

	mux := http.NewServeMux() //crea el tipo router, así como en el microservicio de autores, siento que se van a repetir varias cosas

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { //ruta para salud, igual que en authors
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("/api/v1/books", func(w http.ResponseWriter, r *http.Request) { //endopoint para crar libros
		if r.Method != http.MethodPost { //solo admite post
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed) //va a devolver el 406 si no es post
			return
		}
		//esto es para leer y guardar temporalmente los datos que llegan en el cuerpo de una petición http a un json
		//uso esto ahorita para probar, ACUERDATE DE QUITARLO depsués y usar el dto
		var in struct { //para parsear el JSon de entrada
			AuthorID int64  `json:"authorID"`
			Title    string `json:"title"`
			Year     int    `json:"year"`
			Genre    string `json:"genre"`
			Language string `json:"language"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}

		out := map[string]any{
			"id":       1,
			"title":    in.Title,
			"year":     in.Year,
			"genre":    in.Genre,
			"language": in.Language,
		}
		w.Header().Set("Content-Type", "application/json") //indica la respuesta del JSON
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(out)
	})

	consulAddr := getenv("CONSUL_ADDR", "consul:8500")
	serviceName := getenv("SERVICE_HOST", "books") //nombre con el que se registrará el servicio
	serviceID := serviceName + "-" + strings.TrimPrefix(addr, ":")

	_, p, err := net.SplitHostPort(addr) //separa host y el puerto de addr
	if err != nil {                      //si la sddr no es válida entonces error
		fmt.Println("[books] invalid ADDR:", err)
		os.Exit(1)
	}
	port, _ := strconv.Atoi(p) //convierte el puerto a un entero

	registerAddress := getenv("REGISTER_ADDRESS", "books")            // dirección que el consul va a usar para llamar al servicio
	healthHost := getenv("HEALTH_HOST", registerAddress)              //que use el health check
	healthURL := fmt.Sprintf("http://%s:%d/health", healthHost, port) //url completa del health check

	go func() { //otra rutina así como con authors para registrar consul con reintentos hasta lograrlo
		for {
			if err := consulRegister(consulAddr, serviceName, serviceID, registerAddress, port, healthURL); err != nil {
				fmt.Println("[books] consul register error:", err)
				time.Sleep(2 * time.Second)
				continue
			}
			fmt.Println("[books] registered in consul:", serviceID)
			break
		}
	}()

	go func() { //rutina para cuando se recibe una señal de salida
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		<-ch //esperar a que llegue la señal
		_ = consulDeregister(consulAddr, serviceID)
		os.Exit(0)
	}()

	fmt.Println("[books] listening on", addr) //LOG (donde se está escuhando el servidor)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Println("[books] http error:", err)
	}
}

func consulRegister(consulAddr, name, id, address string, port int, healthURL string) error { //registra el servicio en el consul
	body := map[string]any{ //construye el JSON que consul espera para registrar u servicio
		"Name":    name,
		"ID":      id,
		"Address": address,
		"Port":    port,
		"Tags":    []string{"authorsbooks"}, //esto es para agrupación / filtrado
		"Check": map[string]any{ //va a checar que esté bien
			"HTTP":                           healthURL,
			"Interval":                       "10s",
			"Timeout":                        "2s",
			"DeregisterCriticalServiceAfter": "1m",
		},
	}
	b, _ := json.Marshal(body)                                                                                       //serializa el mapa a JSON
	req, _ := http.NewRequest(http.MethodPut, "http://"+consulAddr+"/v1/agent/service/register", bytes.NewReader(b)) //put para registrar el servicio
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req) //envía la pet http por defecto
	if err != nil {                         //si no, error
		return err
	}
	defer resp.Body.Close() //cierra el body al terminar, si el status no es 2xx entonces hay error
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("bad status: %s", resp.Status)
	}
	return nil
}

func consulDeregister(consulAddr, id string) error { //desregistra el servicio en consul
	req, _ := http.NewRequest(http.MethodPut, "http://"+consulAddr+"/v1/agent/service/deregister/"+id, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	return nil
}
