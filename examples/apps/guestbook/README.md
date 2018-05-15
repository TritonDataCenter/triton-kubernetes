# Example: Guestbook application on Kubernetes

This example is based on the [PHP Guestbook application](https://github.com/kubernetes/examples/tree/master/guestbook) managed by Kubernetes org.

## Changes

To run this example, make sure you have copied and set up your `~/.kube/conf`.

```bash
# git the repo
git clone https://github.com/kubernetes/examples.git
# modify yaml to expose the app
sed -i -- 's/# type: LoadBalancer/type: LoadBalancer/g' ~/Downloads/examples/guestbook/all-in-one/guestbook-all-in-one.yaml
# run the app
kubectl create -f guestbook/all-in-one/guestbook-all-in-one.yaml
```

><sub>This example uses service type loadbalancer which might not be available on all clusters.</sub>