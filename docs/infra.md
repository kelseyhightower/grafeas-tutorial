# Infra

```
gcloud alpha container clusters create grafeas \
  --zone us-west1-b \
  --cluster-version=1.7.8 \
  --enable-kubernetes-alpha \
  --machine-type n1-standard-4 \
  --num-nodes 3
```
