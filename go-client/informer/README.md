# Example Informer

## Try It Out!

You can either use the provided skaffold file and run `skaffold dev` to run the
whole thing or you can do it yourself.
If you want to do it yourself, you will need to first off build and push the
container image for this app (make sure you modify the `push` recipe in the
[Makefile](./Makefile) and the `image` field in the Deployment inside of
[app.yaml](./app.yaml)):
```bash
make
```

Now apply it and see it work:
```bash
$ kubectl delete -f app.yaml 
serviceaccount "pod-manager" deleted
role.rbac.authorization.k8s.io "pod-manager" deleted
rolebinding.rbac.authorization.k8s.io "pod-manager" deleted
deployment.apps "pod-manager" deleted
```

```bash
$ kubectl get po -w
NAME                           READY   STATUS              RESTARTS   AGE
demo-pod                       0/1     ContainerCreating   0          3s
pod-manager-5d8c84d6d4-dp46v   1/1     Running             0          4s
```
