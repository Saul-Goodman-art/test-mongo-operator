# Orders-k8s-go-app

Развертывание: API (Go) + Logger (Go) + Kafka + MongoDB в kind (`kind-control-plane`).

## Требования
- Docker Desktop (с kind integration)
- kind (кластер `kind-control-plane`)
- kubectl
- helm


# создать кластер в контрол плэйн со своим именем:
C:\Users\l1bxc>kind create cluster --name my-gpt-cluster

## 1) Сборка Docker-образов (PowerShell / WSL)
```powershell

примечание:
используем тэг latest
!!

# из корня проекта
docker build -t orders-api:0.4 ./api
docker build -t orders-logger:0.4 ./logger

# Загрузить образы в kind
kind load docker-image orders-api:0.4 --name my-7-gpt-cluster
kind load docker-image orders-logger:0.4 --name my-7-gpt-cluster
```
# помним, что теперь мы отказались от битнами зависимостей, и 
# теперь образы монги , кафки, зоокипера надо самому грузить в кластер

## 2) Установка chart
```powershell
cd helm-chart
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update


helm dependency update

## (примечание: нужно использовать последние версии чартов для кафки и монги в файле
## chart.yaml   А то были конфликты, и следующая внизу команда не срабатывала)


helm install orders-app . --namespace orders --create-namespace

## примечание:
        если нужно удалить и переустанвить:
        helm uninstall order-service
        helm install order-service .
        
        или еще такой вариант команды:
        helm uninstall orders-app -n orders
        helm install orders-app . --namespace orders --create-namespace

```

## 3) Проверка
```powershell
kubectl get pods -n orders
kubectl logs -l app=orders-api -n orders
kubectl logs -l app=orders-logger -n orders
kubectl get jobs -n orders
```

## 4) Тест API
```powershell

[//]: # (kubectl port-forward svc/orders-app-api 8080:80 -n orders)

kubectl port-forward svc/orders-app-api 8000:8000 -n orders

# В другом окне
curl -X POST http://localhost:8080/orders -H "Content-Type: application/json" -d '{"customer":"Test","item":"Toy","amount":9.9}'
curl http://localhost:8080/orders
```

## Примечания
- В values.yaml включены `mongodb.auth.enabled=false` и `persistence.enabled=false` для упрощённой локальной отладки.
- Job миграции вставляет три заказа при первом деплое.
- Для production включите авторизацию и персистентность в values.yaml и настройте секреты.
```

---
