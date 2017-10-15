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
gpg --quick-generate-key image.signer@example.com
```

```
gpg --list-keys --keyid-format short
```

```
gcloud container images describe gcr.io/hightowerlabs/echo \
  --format json > image-summary.json
```

```
gpg -u image.signer@example.com \
  --sign \
  --armor \
  --output=signature.gpg \
  image-summary.json
```

```
gpg --export image.signer@example.com > keyring.pub
```

```
gpg --no-default-keyring \
  --keyring ./keyring.pub \
  --output test.json \
  --verify signature.gpg
```


```
curl -X POST \
  "http://127.0.0.1:8080/v1alpha1/projects/image-signing/notes?noteId=production" \
  -d @note.json
```

```
curl -s http://127.0.0.1:8080/v1alpha1/projects/image-signing/notes/production | jq
```

```
curl -X POST \
  'http://127.0.0.1:8080/v1alpha1/projects/image-signing/occurrences' \
  -d @occurrence.json
```

```
curl -s 'http://127.0.0.1:8080/v1alpha1/projects/image-signing/occurrences' | jq
```

```
IMAGE_SIGNATURE=$(curl -s http://127.0.0.1:8080/v1alpha1/projects/image-signing/notes/jira | \
  jq -r '.buildType.signature.signature' | \
  base64 -D -)
```

```
KEY_ID=$(curl -s http://127.0.0.1:8080/v1alpha1/projects/image-signing/notes/jira | \
  jq -r '.buildType.signature.keyId')
```
