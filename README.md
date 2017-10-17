# Grafeas Tutorial

This tutorial will guide you through testing Grafeas.

## Tutorial

```
gcloud alpha container clusters create grafeas \
  --enable-kubernetes-alpha \
  --cluster-version 1.8.0-gke.1
```

```
brew install gpg2
```

```
gpg --quick-generate-key --yes image.signer@example.com 
```

```
gpg --list-keys --keyid-format short
```

```
gpg --armor --export image.signer@example.com > key.pub
```

```
gpg -u image.signer@example.com \
  --armor \
  --clearsign \
  --output=signature.gpg \
  image.txt
```

```
gpg --output - --verify signature.gpg
```

```
curl -X POST \
  "http://127.0.0.1:8080/v1alpha1/projects/image-signing/notes?noteId=production" \
  -d @note.json
```

```
curl -X POST \
  'http://127.0.0.1:8080/v1alpha1/projects/image-signing/occurrences' \
  -d @occurrence.json
```

```
kubectl create configmap image-signature-webhook \
  --from-file D2D2E339.pub=key.pub
```
