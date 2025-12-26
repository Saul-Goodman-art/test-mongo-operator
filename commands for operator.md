комнады для оператора (оператор в одном простр имен с прилож)


# 1. Создаем namespace для оператора (если он уже есть, команда ничего не сделает)
kubectl create namespace mongodb-operator

# 2. Устанавливаем Custom Resource Definition (глобально, без namespace)
kubectl apply -f https://raw.githubusercontent.com/mongodb/mongodb-kubernetes-operator/v0.8.0/config/crd/bases/mongodbcommunity.mongodb.com_mongodbcommunity.yaml

# 3. Устанавливаем ВСЕ RBAC и служебные ресурсы в НУЖНЫЙ namespace
# 3.1. Основные роли и привязки для оператора
kubectl apply -f https://raw.githubusercontent.com/mongodb/mongodb-kubernetes-operator/v0.8.0/config/rbac/role.yaml --namespace mongodb-operator
kubectl apply -f https://raw.githubusercontent.com/mongodb/mongodb-kubernetes-operator/v0.8.0/config/rbac/role_binding.yaml --namespace mongodb-operator
kubectl apply -f https://raw.githubusercontent.com/mongodb/mongodb-kubernetes-operator/v0.8.0/config/rbac/service_account.yaml --namespace mongodb-operator

# 3.2. Дополнительные роли и привязки для базы данных
kubectl apply -f https://raw.githubusercontent.com/mongodb/mongodb-kubernetes-operator/v0.8.0/config/rbac/role_database.yaml --namespace mongodb-operator
kubectl apply -f https://raw.githubusercontent.com/mongodb/mongodb-kubernetes-operator/v0.8.0/config/rbac/role_binding_database.yaml --namespace mongodb-operator
kubectl apply -f https://raw.githubusercontent.com/mongodb/mongodb-kubernetes-operator/v0.8.0/config/rbac/service_account_database.yaml --namespace mongodb-operator

# 4. Устанавливаем сам Deployment оператора (контроллер)
kubectl apply -f https://raw.githubusercontent.com/mongodb/mongodb-kubernetes-operator/v0.8.0/config/manager/manager.yaml --namespace mongodb-operator

# 5. Сразу применяем ВАШЕ критическое исправление для сервисного аккаунта базы данных
kubectl apply -f deploy/fix-database-sa-rbac.yaml

# 6. Проверяем, что оператор запустился (подождем 30 секунд и проверим)
Start-Sleep -Seconds 30
kubectl get pods -n mongodb-operator






kubectl get pods -n mongodb-operator -l name=mongodb-kubernetes-operator

kubectl get role,rolebinding,serviceaccount -n mongodb-operator

kubectl get role mongodb-database -n mongodb-operator

helm install orders-app ./helm-chart -n mongodb-operator

kubectl port-forward svc/orders-app-api 8000:8000 -n mongodb-operator