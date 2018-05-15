# Setting up HTTP Load Balancing with Ingress

This example shows how to run an web application behind an HTTP load balancer by configuring the Ingress resource.

## Deploy a web application

Create a Deployment using the sample web application container image that listens on a HTTP server on port 8080:

```
$kubectl run web --image=gcr.io/google-samples/hello-app:1.0 --port=8080

```
Verify the Service was created and a node port was allocated:
```
$kubectl get service web
NAME      TYPE       CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
web       NodePort   10.96.135.94   <none>        8080:32640/TCP   1m
```

## Expose your Deployment as a Service internally

Create a Service resource to make the web deployment reachable within your container cluster:

```
kubectl expose deployment web --target-port=8080 --type=NodePort
```

Verify the Service was created and a node port was allocated:
```
kubectl get service web
NAME      TYPE       CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
web       NodePort   10.96.135.94   <none>        8080:32640/TCP   2m
```
In the sample output above, the node port for the web Service is 32640. Also, note that there is no external IP allocated for this Service. Since the Kubernetes Engine nodes are not externally accessible by default, creating this Service does not make your application accessible from the Internet.

To make your HTTP(S) web server application publicly accessible, you need to create an Ingress resource.

## Create an Ingress resource

To deploy this Ingress resource run:
```
kubectl apply -f examples/apps/ingress/ingress.yaml
```

## Visit your application

Find out the external IP address of the load balancer serving your application by running:

```
$kubectl get ingress basic-ingress

NAME            HOSTS     ADDRESS         PORTS     AGE
basic-ingress   *         165.225.128.1    80        7m
````
To see all the external IP addresses run:

```
 kubectl get ingress basic-ingress -o yaml
```