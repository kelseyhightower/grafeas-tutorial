#!/bin/bash

cfssl gencert -initca ca-csr.json | cfssljson -bare ca
cfssl gencert \
  -ca=ca.pem \
  -ca-key=ca-key.pem \
  -config=ca-config.json \
  -hostname=127.0.0.1,image-signature-webhook,image-signature-webhook.kube-system,image-signature-webhook.default,image-signature-webhook.default.svc \
  -profile=default \
  image-signature-webhook-csr.json | cfssljson -bare image-signature-webhook

kubectl create secret tls tls-image-signature-webhook \
  --cert=image-signature-webhook.pem \
  --key=image-signature-webhook-key.pem
