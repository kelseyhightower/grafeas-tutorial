


sleep 2
kubectl delete pods $(kubectl get pods -l app=image-signature-webhook -o jsonpath='{.items[0].metadata.name}')
sleep 10

kubectl delete pods $(kubectl get pods -l run=nginx -o jsonpath='{.items[0].metadata.name}')
sleep 10

