DOCKER_REGISTRY=localhost:5000
IMAGE_VERSION=1.0

MICROSERVICES= \
    api-gateway \
    user-service \
    notification-service

.PHONY: help kind-create-cluster kind-delete-cluster \
	kind-deploy-metrics-server kind-deploy-nginx-ingress kind-delete-nginx-ingress \
	kind-build-images kind-push-images kind-deploy-services

# Utility for colored output
define PRINT_COLOR
\033[1;32m$(1)\033[0m
endef

help:
	@printf "$(call PRINT_COLOR,Available commands:\n)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "} {printf "%-30s %s\n", $$1, $$2}'

kind-create-cluster: ## Create KIND cluster, local registry, metrics service, namespaces
	@printf "$(call PRINT_COLOR,Creating Kind cluster\n)"
	@kind create cluster --config=./k8s/kind-config.yaml

	@printf "$(call PRINT_COLOR,Creating namespace\n)"
	@kubectl apply -f k8s/namespace.yaml

	$(MAKE) kind-deploy-metrics-server

	@printf "$(call PRINT_COLOR,Starting Docker registry container\n)"
	@docker run -d --rm --name kind-registry --net kind -p 5000:5000  registry:2 || true

	@printf "$(call PRINT_COLOR,Setting kubectl context to "kind-kind" and default namespace to "microservices"\n)"
	@kubectl config set-context kind-kind --namespace=microservices

kind-delete-cluster: ## Delete Kind cluster
	@printf "$(call PRINT_COLOR,Deleting Kind cluster\n)"
	@kind delete cluster || true
	@printf "$(call PRINT_COLOR,Stopping Docker registry container\n)"
	@docker stop kind-registry || true

kind-deploy-metrics-server: ## Install metrics server so we can see resource usage for pods etc
	@kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

	@printf "$(call PRINT_COLOR,Patch the metrics-server deployment to allow insecure TLS connections to the kubelet.\n)"
	@kubectl patch deployment metrics-server \
	   --namespace kube-system \
	   --type='json' \
	   -p='[{"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--kubelet-insecure-tls"}]'

	@printf "$(call PRINT_COLOR,Waiting for metrics server to be ready...\n)"
	@kubectl wait --namespace kube-system \
       --for=condition=Ready pod \
       --field-selector=status.phase=Running \
       --timeout=360s

kind-deploy-nginx-ingress: ## Install and run NGINX Ingress + cloud-provider-kind
	@printf "$(call PRINT_COLOR,Deploying NGINX Ingress Controller\n)"
	@kubectl apply -f https://kind.sigs.k8s.io/examples/ingress/deploy-ingress-nginx.yaml

	@kubectl wait --namespace ingress-nginx \
      --for=condition=ready pod \
      --selector=app.kubernetes.io/component=controller \
      --timeout=360s

	@nohup cloud-provider-kind >/tmp/cloud-provider-kind.log 2>&1 &
	@printf "$(call PRINT_COLOR,Started cloud-provider-kind in background - logs path: tmp/cloud-provider-kind.log \n)"

	@printf "$(call PRINT_COLOR,Waiting for External IP...\n)"
	@until kubectl get svc -n ingress-nginx ingress-nginx-controller -o=jsonpath='{.status.loadBalancer.ingress[0].ip}' | grep -Eo '([0-9]{1,3}\.){3}[0-9]{1,3}'; do \
	    sleep 5; \
	done

	@printf "$(call PRINT_COLOR,NGINX Ingress is available at: http://$$(kubectl get svc -n ingress-nginx ingress-nginx-controller -o=jsonpath='{.status.loadBalancer.ingress[0].ip}')\n)"

kind-build-images: ## Build Docker images for Microservices
	@printf "$(call PRINT_COLOR,Building Docker images\n)"

	@for service in $(MICROSERVICES); do \
		printf "$(call PRINT_COLOR,Building $$service image\n)"; \
		docker build --target production -t $(DOCKER_REGISTRY)/$$service:$(IMAGE_VERSION) ./$$service; \
	done

kind-push-images: ## Push Docker images to local registry
	@printf "$(call PRINT_COLOR,Pushing docker images to local registry: $(DOCKER_REGISTRY)\n)"

	@for service in $(MICROSERVICES); do \
		printf "$(call PRINT_COLOR,Pushing $$svc image\n)"; \
		docker push $(DOCKER_REGISTRY)/$$service:$(IMAGE_VERSION); \
	done

kind-deploy-services: ## Deploy all services to Kind
	@printf "$(call PRINT_COLOR,Deploying RabbitMQ\n)"
	@kubectl apply -f ./k8s/rabbitmq

	@printf "$(call PRINT_COLOR,Deploying Microservices\n)"

	@for service in $(MICROSERVICES); do \
		printf "$(call PRINT_COLOR,Deploying $$service\n)"; \
	  kubectl apply -f ./k8s/$$service; \
	done
