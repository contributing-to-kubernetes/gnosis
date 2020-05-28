# Bootstrapping Kubernetes Clusters in GCP

# Creating a Cluster with the the `gcloud` CLI

Make sure you are logged in
```
$ gcloud auth login
```

Choose a project:
```
$ gcloud config set project <project ID>
```

And finally create a cluster:
```
gcloud container clusters create kube-up --region us-central1-a
```
If you use `--machine-type f1-micro` you'll get really cheap worker nodes.
If you want to specify `--cluster-version` you can get the available Kubernetes
versions by running `gcloud container get-server-config`.

You can delete this cluster by running:
```
$ gcloud container clusters delete kube-up --region us-central1-a
```
