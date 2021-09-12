# metacontroller

Something based on Metacontroller

# prerequisite
* Install metacontroller following https://metacontroller.github.io/metacontroller/guide/install.html
`kubectl apply -k https://github.com/metacontroller/metacontroller/manifests/production`

#Implementation

Something simple based on Composite MetaController.

# Debug.
Make sure that local version is actually deployed before debugging. Use make goal "restart" 

Steps: 
* Create N pods
* Assign labels to pods
* 