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
gpg --batch --gen-key gpg-batch
```

```
gpg --no-default-keyring --keyring ./grafeas.pub --list-keys --keyid-format long
```

```
gpg: /Users/khightower/.gnupg/trustdb.gpg: trustdb created
./grafeas.pub
-------------
pub   rsa2048 2017-10-14 [SCEA]
      C08F26EE4E2CDB05C9663CC486C27B9650AEAF07
uid           [ unknown] Container Image Signer (docker image signing key) <images@example.com>
sub   rsa2048 2017-10-14 [SEA]
```

```
gpg --no-default-keyring \
--keyring ./grafeas.pub \
--edit-key C08F26EE4E2CDB05C9663CC486C27B9650AEAF07 \
trust
```

```
gpg (GnuPG) 2.2.1; Copyright (C) 2017 Free Software Foundation, Inc.
This is free software: you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.

Secret key is available.

sec  rsa2048/86C27B9650AEAF07
     created: 2017-10-14  expires: never       usage: SCEA
     trust: unknown       validity: unknown
ssb  rsa2048/4DB95B55B2AE4A84
     created: 2017-10-14  expires: never       usage: SEA 
[ unknown] (1). Container Image Signer (docker image signing key) <images@example.com>

sec  rsa2048/86C27B9650AEAF07
     created: 2017-10-14  expires: never       usage: SCEA
     trust: unknown       validity: unknown
ssb  rsa2048/4DB95B55B2AE4A84
     created: 2017-10-14  expires: never       usage: SEA 
[ unknown] (1). Container Image Signer (docker image signing key) <images@example.com>

Please decide how far you trust this user to correctly verify other users' keys
(by looking at passports, checking fingerprints from different sources, etc.)

  1 = I don't know or won't say
  2 = I do NOT trust
  3 = I trust marginally
  4 = I trust fully
  5 = I trust ultimately
  m = back to the main menu

Your decision? 5
Do you really want to set this key to ultimate trust? (y/N) y
                                                             
sec  rsa2048/86C27B9650AEAF07
     created: 2017-10-14  expires: never       usage: SCEA
     trust: ultimate      validity: unknown
ssb  rsa2048/4DB95B55B2AE4A84
     created: 2017-10-14  expires: never       usage: SEA 
[ unknown] (1). Container Image Signer (docker image signing key) <images@example.com>
Please note that the shown key validity is not necessarily correct
unless you restart the program.

gpg> quit
```

```
gpg \
  --no-default-keyring \
  --keyring ./grafeas.pub \
  --export images@example.com > keyring.pub
```

```
gpg -u images@example.com \
  --no-default-keyring \
  --keyring ./grafeas.pub \
  --armor \
  --clearsign \
  -o image-signature.pgp \
  image-signature.json
```

```
gpg --no-default-keyring \
  --armor \
  --keyring ./trust.pub \
  --import ./pubkeys.gpg
```

```
gpg --no-default-keyring \
  --keyring ./trust.pub \
  --verify image-signature.pgp
```


```
curl -X POST \
  "http://127.0.0.1:8080/v1alpha1/projects/image-signing/notes?noteId=jira" \
  -d @note.json
```

```
curl http://127.0.0.1:8080/v1alpha1/projects/image-signing/notes/jira
```

```
gcloud container images describe gcr.io/hightowerlabs/echo --format json > image-summary.json
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
