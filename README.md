# Grafeas Tutorial

This tutorial will guide you through testing Grafeas.  In it, you will create a Kubernetes cluster configured to only allow container images signed by a specific key, configurable via a configmap.  Container image signatures will be stored in Grafeas.  To make sure only signed images are allowed, you will start an admission plugin service which finds signatures in Grafeas and verifies them.

## Tutorial

### Infrastructure

A Kubernetes 1.8+ cluster is required with support for the [external admission webhooks](https://kubernetes.io/docs/admin/extensible-admission-controllers/#external-admission-webhooks) alpha feature enabled. The `image-signature-webhook` external admission webhook will be used to enforce only signed images are allowed to be deployed to the cluster.

Create a 1.8 Kubernetes cluster:

```
gcloud alpha container clusters create grafeas \
  --enable-kubernetes-alpha \
  --cluster-version 1.8.0-gke.1
```

### Deploy the Grafeas Server

[Grafeas](http://grafeas.io/about) is an open artifact metadata API to audit and govern your software supply chain. In this tutorial Grafeas will be used to store container image signatures. 

Create the Grafeas server deployment:

```
kubectl apply -f kubernetes/grafeas.yaml
```

> While in early alpha the Grafeas server leverages an in-memory data store. If the Grafeas server is ever restarted, all image signature must be repopulated.

### Create Image Signature 

Install gpg:

```
brew install gpg2
```

Generate a signing key:

```
gpg --quick-generate-key --yes image.signer@example.com 
```

List the keys and store the key ID:

```
gpg --list-keys --keyid-format short
```

Store the gpg key ID in the `GPG_KEY_ID` env var:

```
GPG_KEY_ID="0CD9D96F"
```

Export the image signer's public key:

```
gpg --armor --export image.signer@example.com > ${GPG_KEY_ID}.pub
```

Sign the image digest text file:

```
gpg -u image.signer@example.com \
  --armor \
  --clearsign \
  --output=signature.gpg \
  image-digest.txt
```

Verify the signature:

```
gpg --output - --verify signature.gpg
```

### Create a pgpSignedAttestation Occurrence

In a new terminal create a secure tunnel to the grafeas server:

```
kubectl port-forward \
  $(kubectl get pods -l app=grafeas -o jsonpath='{.items[0].metadata.name}') \
  8080:8080
```

Create the `production` attestationAuthority note:

```
curl -X POST \
  "http://127.0.0.1:8080/v1alpha1/projects/image-signing/notes?noteId=production" \
  -d @note.json
```

Generate an pgpSignedAttestation occurrence:

```
GPG_SIGNATURE=$(cat signature.gpg | base64)
```

```
RESOURCE_URL="https://gcr.io/hightowerlabs/echod@sha256:aba48d60ba4410ec921f9d2e8169236c57660d121f9430dc9758d754eec8f887"
```

```
cat > occurrence.json <<EOF
{
  "resourceUrl": "${RESOURCE_URL}",
  "noteName": "projects/image-signing/notes/production",
  "attestation": {
    "pgpSignedAttestation": {
       "signature": "${GPG_SIGNATURE}",
       "contentType": "application/vnd.gcr.image.url.v1",
       "pgpKeyId": "${GPG_KEY_ID}"
    }
  }
}
EOF
```

Post the pgpSignedAttestation occurrence:

```
curl -X POST \
  'http://127.0.0.1:8080/v1alpha1/projects/image-signing/occurrences' \
  -d @occurrence.json
```

### Deploy the Image Signature Webhook

Create the `image-signature-webhook` configmap and store the image signer's public key: 

```
kubectl create configmap image-signature-webhook \
  --from-file ${GPG_KEY_ID}.pub
```

```
kubectl get configmap image-signature-webhook -o yaml
```

Create the `tls-image-signature-webhook` secret and store the TLS certs:

```
kubectl create secret tls tls-image-signature-webhook \
  --key pki/image-signature-webhook-key.pem \
  --cert pki/image-signature-webhook.pem
```

Create the `image-signature-webhook` deployment:

```
kubectl apply -f kubernetes/image-signature-webhook.yaml 
```

Create the `image-signature-webook` ExternalAdmissionHookConfiguration:

```
kubectl apply -f kubernetes/admission-hook-configuration.yam
```

### Testing the Admission Webhook

```
kubectl apply -f pods/nginx.yaml
```

```
Error from server: error when creating "pods/nginx.yaml": admission webhook "image-signature.hightowerlabs.com" denied the request without explanation
```

```
kubectl apply -f pods/echod.yaml 
```
```
pod "echod" created
```
