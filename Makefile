
NAMESPACE=authorsbooks
K8S_FILE=k8s/authorsbooks.yaml

.PHONY: help \
	docker-build docker-up docker-down docker-reset \
	k8s-start k8s-up k8s-destroy k8s-url \
	grpc-smoke k6-test


# DOCKER
docker-build:
	@echo ">> Exportando variables de entorno de Docker dentro de Minikube..."
	@eval $$(minikube docker-env) && \
	  echo ">> Construyendo authorsbooks-authors-grpc..." && \
	  docker build -t authorsbooks-authors-grpc \
	    -f ./authors/Dockerfile --target authors-grpc . && \
	  echo ">> Construyendo authorsbooks-books-grpc..." && \
	  docker build -t authorsbooks-books-grpc \
	    -f ./books/Dockerfile --target books-grpc . && \
	  echo ">> Construyendo authorsbooks-metadata-grpc..." && \
	  docker build -t authorsbooks-metadata-grpc \
	    -f ./metadata/Dockerfile --target metadata-grpc .
	@echo "✔ Imágenes Docker construidas dentro de Minikube."

docker-up:
	docker compose up -d
	docker compose ps

docker-down:
	docker compose down

docker-reset:
	docker compose down || true
	docker image rm authorsbooks-authors-grpc authorsbooks-books-grpc authorsbooks-metadata-grpc -f || true
	docker volume prune -f || true
	docker system prune -f || true
	@echo "✔ Docker limpiado para este proyecto."


# KUBERNETES

k8s-start:
	minikube start --driver=docker
	kubectl config use-context minikube
	@echo "✔ Minikube listo y contexto 'minikube' seleccionado."

k8s-up:
	@echo ">>> Creando namespace $(NAMESPACE) si no existe..."
	@kubectl get ns $(NAMESPACE) >/dev/null 2>&1 || kubectl create namespace $(NAMESPACE)

	@echo ">>> Aplicando manifests $(K8S_FILE)..."
	kubectl apply -f $(K8S_FILE)

	@echo ">>> Esperando pods..."
	sleep 7

	@echo "=== Pods ==="
	kubectl get pods -n $(NAMESPACE)
	@echo "=== Services ==="
	kubectl get svc -n $(NAMESPACE)
	@echo "=== HPA ==="
	kubectl get hpa -n $(NAMESPACE)
	@echo "✔ Cluster Kubernetes del proyecto desplegado."

k8s-destroy:
	kubectl delete namespace $(NAMESPACE) --ignore-not-found=true
	@echo "✔ Namespace $(NAMESPACE) eliminado."

k8s-url:
	@echo "Obteniendo URL del LoadBalancer (authors-grpc):"
	minikube service authors-grpc -n $(NAMESPACE) --url


# PRUEBA DE COMUNICACIÓN GRPC

grpc-smoke:
	@if [ -z "$(AUTHORS_ADDR)" ]; then \
	  echo "ERROR: Debes pasar AUTHORS_ADDR, por ejemplo:"; \
	  echo "  make grpc-smoke AUTHORS_ADDR=127.0.0.1:54456"; \
	  exit 1; \
	fi; \
	ADDR="$(AUTHORS_ADDR)"; \
	ADDR="$${ADDR#http://}"; \
	ADDR="$${ADDR#https://}"; \
	echo "Usando AUTHORS_ADDR = $$ADDR"; \
	echo ""; \
	echo "=== 1) Creando autora en Authors ==="; \
	grpcurl -plaintext \
	  -d '{"name":"Laura Esquivel"}' \
	  $$ADDR authors.v1.AuthorsService/CreateAuthor || true; \
	echo ""; \
	echo "=== 2) Agregando libro (Authors → Metadata + Books internos) ==="; \
	grpcurl -plaintext \
	  -d '{"author_id":1,"title":"Como agua para chocolate","year":1989,"genre":"Romance","language":"es"}' \
	  $$ADDR authors.v1.AuthorsService/AddBookToAuthor || true; \
	echo ""; \
	echo "=== 3) Consultando autora con libros ==="; \
	grpcurl -plaintext \
	  -d '{"id":1}' \
	  $$ADDR authors.v1.AuthorsService/GetAuthor || true; \
	echo ""; \
	echo "✔ Comunicación entre los 3 microservicios verificada vía Authors (gRPC dentro del cluster)."

# SCALE TEST CON K6

k6-test:
	@if [ -z "$(AUTHORS_ADDR)" ]; then \
	  echo "ERROR: Debes pasar AUTHORS_ADDR, por ejemplo:"; \
	  echo "  make k6-test AUTHORS_ADDR=127.0.0.1:54456"; \
	  exit 1; \
	fi; \
	ADDR="$(AUTHORS_ADDR)"; \
	ADDR="$${ADDR#http://}"; \
	ADDR="$${ADDR#https://}"; \
	echo "Usando AUTHORS_ADDR = $$ADDR para k6..."; \
	k6 run -e AUTHORS_ADDR=$$ADDR k6/authors_grpc_load.js
