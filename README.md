# Grafeas Tutorial

This tutorial will guide you through testing Grafeas.  In it, you will create a Kubernetes cluster configured to only allow container images signed by a specific key, configurable via a configmap.  Container image signatures will be stored in Grafeas.  To make sure only signed images are allowed, you will start an admission plugin service which finds signatures in Grafeas and verifies them.

## Tutorial

### Infrastructure

A Kubernetes 1.8+ cluster is required with support for the [external admission webhooks](https://kubernetes.io/docs/admin/extensible-admission-controllers/#external-admission-webhooks) alpha feature enabled.

If you have access to [Google Container Engine](https://cloud.google.com/container-engine/) use the gcloud command to create a 1.8 Kubernetes cluster:

```
gcloud alpha container clusters create grafeas \
  --enable-kubernetes-alpha \
  --cluster-version 1.8.0-gke.1
```

> Any Kubernetes 1.8 cluster with support for external admission webhooks will work. 

### Deploy the Grafeas Server

[Grafeas](http://grafeas.io/about) is an open artifact metadata API to audit and govern your software supply chain. In this tutorial Grafeas will be used to store container image signatures. 

Create the Grafeas server deployment:

```
kubectl apply -f kubernetes/grafeas.yaml
```

> While in early alpha the Grafeas server leverages an in-memory data store. If the Grafeas server is ever restarted all image signature must be repopulated.

### Generating GPG Signing Keys

In this section you will generate a [gpg keypair](https://www.gnupg.org/gph/en/manual.html#INTRO) suitable for signing container image metadata.

Install gpg for you platform:

#### OS X

```
brew install gpg2
```

#### Linux

```
apt-get install gnupg
```

Once gpg has been installed generate a signing key:

```
gpg --quick-generate-key --yes image.signer@example.com 
```

Retrive the ID of the signing key:

```
gpg --list-keys --keyid-format short
```

```
------------------------------------
pub   rsa2048/0CD9D96F 2017-10-17 [SC] [expires: 2019-10-17]
      510CE141B559A243439EB18926CE52D30CD9D96F
uid         [ultimate] image.signer@example.com
sub   rsa2048/2C216B83 2017-10-17 [E]
```

> Based on the above output the key ID is 0CD9D96F. Your key ID will be different.

Store the ID of your signing key in the `GPG_KEY_ID` env var:

```
GPG_KEY_ID="0CD9D96F"
```
#### Signing Container Image Metadata

Container images tend to range in size from a few megabytes to multiple gigabytes. Signing and distributing container images can be quite resource intensive so we are going to opt for signing the [image digest](https://cloud.google.com/container-registry/docs/concepts/image-formats#content_addressability) which uniquely identifies a container image.

In this tutorial the `gcr.io/hightowerlabs/echod` container image will be used for testing. Instead of trusting an image tag such `0.0.1`, which can be reused and point to a different container image later, we are going to trust the image digest. 

```
cat image-digest.txt
```
```
sha256:aba48d60ba4410ec921f9d2e8169236c57660d121f9430dc9758d754eec8f887
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

```
sha256:aba48d60ba4410ec921f9d2e8169236c57660d121f9430dc9758d754eec8f887
gpg: Signature made Tue Oct 17 09:11:53 2017 PDT
gpg:                using RSA key 510CE141B559A243439EB18926CE52D30CD9D96F
gpg:                issuer "image.signer@example.com"
gpg: Good signature from "image.signer@example.com" [ultimate]
```

In order for others to verify signed images they must trust and have access to the image signer's public key. Export the image signer's public key:

```
gpg --armor --export image.signer@example.com > ${GPG_KEY_ID}.pub
```

### Create a pgpSignedAttestation Occurrence

Now that we have a signed container image, and a public key to verify it, we need to create a [pgpSignedAttestation occurrence](https://github.com/Grafeas/Grafeas/blob/master/samples/server/go-server/api/docs/PgpSignedAttestation.md) using the Grafeas API.

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

At this point the `gcr.io/hightowerlabs/echod` container image can be verified through the Grafeas API.

> Only the `gcr.io/hightowerlabs/echod` container image identified by the `sha256:aba48d60ba4410ec921f9d2e8169236c57660d121f9430dc9758d754eec8f887` image digest and be verified by the Grafeas API. Additional images require a new occurrence. 

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
kubectl apply -f kubernetes/admission-hook-configuration.yaml
```

### Testing the Admission Webhook

Attempt to run the `nginx:1.13` container image which does not have an pgpSignedAttestation occurrence in the Grafeas API. Create the `nginx` pod:

```
kubectl apply -f pods/nginx.yaml
```

Notice the `nginx` pod was not created and the follow error was returned: 

```
The  "" is invalid: : No matched signatures for container image: nginx:1.13
```

Attempt to run the `gcr.io/hightowerlabs/echod@sha256:aba48d60ba4410ec921f9d2e8169236c57660d121f9430dc9758d754eec8f887` container image which has an pgpSignedAttestation occurrence in the Grafeas API.

```
kubectl apply -f pods/echod.yaml 
```
```
pod "echod" created
```

At this point the following pods should be running in your cluster:

```
kubectl get pods
```
```
NAME                                       READY     STATUS    RESTARTS   AGE
echod                                      1/1       Running   0          5m
grafeas-5b5759cbcf-lx8r5                   1/1       Running   0          12m
image-signature-webhook-6cc7d6bd74-55blt   1/1       Running   0          8m
```

> Notice the `nginx` pod was not created because the `nginx:1.13` container image was not verified by the image signature webhook.

## Cleanup

```
kubectl delete deployments grafeas image-signature-webhook
kubectl delete pods echod
kubectl delete svc grafeas image-signature-webhook
kubectl delete secrets tls-image-signature-webhook
kubectl delete configmap image-signature-webhook
```
