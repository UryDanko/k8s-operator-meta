# Metacontroller
Metacontroller is an add-on for Kubernetes that makes it easy to write and deploy custom controllers. 
Although the open-source project was started at Google, the add-on works the same in any Kubernetes cluster.

Here I created a POC to try the technology. It is a simple operator based on Composite MetaController. 
Currently it only deploys two pods with nginx and creates a service.

# Prerequisite
* Install metacontroller following https://metacontroller.github.io/metacontroller/guide/install.html
`kubectl apply -k https://github.com/metacontroller/metacontroller/manifests/production`
or 
`make init-metacontroller`

# Deploy
Just run `make restart` - it'll rebuild docker image, push it to an inventory and redeploy everything. 

# Debug.
* Make sure that local version is actually deployed before debugging. Use make goal "restart" (debug docker image will be built).
* Run `kubectl port-forward --namespace metacontroller $(kubectl get pod --namespace metacontroller --selector="app=sandbox-controller" --output jsonpath='{.items[0].metadata.name}') 40000:40000`
* Run debug from IDE.
* Don't forget to build separately Dockerfile.delve (I extracted it to speed up main docker image build)

# Known issues
* method: RollingRecreate - doesn't really work with pods.

# TODO:
* Read pod properties from sandbox.yaml instead of hardcodedю 
* Update sandbox.yaml with service details. 


# Cleanup
`make undeploy-sandbox undeploy-sandbox-controller undeploy-metacontroller`