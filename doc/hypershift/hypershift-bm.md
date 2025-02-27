# Provision HyperShift Hosted Cluster on Bare Metal

This document is used to provision HyperShift hosted cluster on bare metal with hypershift-addon in ACM 2.5+.

## Prerequisites

1. Set the following environment variables that will be used later:

```bash
export OPENSHIFT_PULL_SECRET_FILE=<openshift-pull-secret-file>
export OPENSHIFT_PULL_SECRET_NAME=<pull-secret-name>
export OPENSHIFT_PULL_SECRET_NAMESPACE=<pull-secret-namespace>
```

2. Create OpenShift pull secret by the following commands:

```bash
oc create ns ${OPENSHIFT_PULL_SECRET_NAMESPACE}
export PS64=$(cat ${OPENSHIFT_PULL_SECRET_FILE} | base64 -w0)
oc apply -f - <<EOF
apiVersion: v1
data:
  .dockerconfigjson: ${PS64}
kind: Secret
metadata:
  name: ${OPENSHIFT_PULL_SECRET_NAME}
  namespace: ${OPENSHIFT_PULL_SECRET_NAMESPACE}
type: kubernetes.io/dockerconfigjson
EOF
```

## Create HyperShift Hosted Cluster on Bare Metal

1. Set the following environment variables for `HypershiftDeployment` resource:

```bash
export HYPERSHIFT_MGMT_CLUSTER=hypermgt
export HYPERSHIFT_HOSTING_NAMESPACE=clusters
export OPENSHIFT_RELEASE_IMAGE=quay.io/openshift-release-dev/ocp-release:4.10.15-x86_64
export OPENSHIFT_BASE_DOMAIN=example.com
export HYPERSHIFT_DEPLOYMENT_NAME=<hypershiftdeployment-name>
```

2. Following the [guide](https://hypershift-docs.netlify.app/how-to/none/create-none-cluster/#requisites) to configure the DNS for the HyperShift hosted cluster on bare metal.

_Note_: If the worker node of the HyperShift management cluster doesn't have external IPs, then the HyperShift hosted cluster created in the following steps can not be accessed from external, but can be accessed from internal for local dev/test.

3. Create `HypershiftDeployment` resource to provision hosted cluster on bare metal:

```bash
oc apply -f - <<EOF
apiVersion: cluster.open-cluster-management.io/v1alpha1
kind: HypershiftDeployment
metadata:
  name: ${HYPERSHIFT_DEPLOYMENT_NAME}
  namespace: ${OPENSHIFT_PULL_SECRET_NAMESPACE}
spec:
  hostingCluster: ${HYPERSHIFT_MGMT_CLUSTER}
  hostingNamespace: ${HYPERSHIFT_HOSTING_NAMESPACE}
  hostedClusterSpec:
    infraID: ${HYPERSHIFT_DEPLOYMENT_NAME}
    platform:
      type: None
    pullSecret:
      name: ${OPENSHIFT_PULL_SECRET_NAME}
    release:
      image: ${OPENSHIFT_RELEASE_IMAGE}
    networking:
      machineCIDR: 192.168.123.0/24
      podCIDR: 10.132.0.0/14
      serviceCIDR: 172.31.0.0/16
    dns:
      baseDomain: ${OPENSHIFT_BASE_DOMAIN}
    services:
    - service: APIServer
      servicePublishingStrategy:
        type: NodePort
        nodePort:
          address: api.${HYPERSHIFT_DEPLOYMENT_NAME}.${OPENSHIFT_BASE_DOMAIN}
    - service: OAuthServer
      servicePublishingStrategy:
        type: Route
    - service: Konnectivity
      servicePublishingStrategy:
        type: Route
    - service: Ignition
      servicePublishingStrategy:
        type: Route
    sshKey: {}
  nodePools:
  - name: ${HYPERSHIFT_DEPLOYMENT_NAME}-workers
    spec:
      clusterName: ${HYPERSHIFT_DEPLOYMENT_NAME}
      nodeCount: 0
      platform:
        type: None
      management:
        upgradeType: Replace
      release:
        image: ${OPENSHIFT_RELEASE_IMAGE}
  infra-id: ${HYPERSHIFT_DEPLOYMENT_NAME}
  infrastructure:
    configure: false
EOF
```

4. Get managed cluster name of the hypershift hosted cluster created in last step and wait until managed cluster is available:

```bash
export HYPERSHIFT_MANAGED_CLUSTER_NAME=$(oc get managedcluster | grep ${HYPERSHIFT_DEPLOYMENT_NAME} | awk '{print $1}')
oc wait --for=condition=ManagedClusterConditionAvailable managedcluster/${HYPERSHIFT_MANAGED_CLUSTER_NAME} --timeout=600s
```

5. Retrieve kubeconfig for the HyperShift hosted cluster:

```bash
oc -n ${HYPERSHIFT_MGMT_CLUSTER} get secret ${HYPERSHIFT_MANAGED_CLUSTER_NAME}-admin-kubeconfig -o jsonpath="{.data.kubeconfig}" | base64 -d
```

  _Note_: If the worker node of the HyperShift management cluster doesn't have external IPs, then the HyperShift hosted cluster created in the following steps can not be accessed from external, but can be accessed from internal for dev/test. Add a new DNS entry `<ip-of-local-machine> api.${HYPERSHIFT_DEPLOYMENT_NAME}.${OPENSHIFT_BASE_DOMAIN}` to `/etc/hosts` of the local machine and then execute the following commands to forward the local traffic to kube-apiserver service.

  ```bash
  export LOCAL_MACHINE_IP=<IP-for-local-machine>
  export HYPERSHIFT_MGMT_CLUSTER_KUBECONFIG=<kubeconfig-path-to-hypershift-management-cluster>
  ```
  ```bash
  API_SERVICE_NODEPORT=$(oc --kubeconfig=${HYPERSHIFT_MGMT_CLUSTER_KUBECONFIG} -n ${HYPERSHIFT_HOSTING_NAMESPACE}-${HYPERSHIFT_DEPLOYMENT_NAME} get svc/kube-apiserver -o jsonpath='{.spec.ports[?(@.port==6443)].nodePort}')
  oc --kubeconfig=${HYPERSHIFT_MGMT_CLUSTER_KUBECONFIG} -n ${HYPERSHIFT_HOSTING_NAMESPACE}-${HYPERSHIFT_DEPLOYMENT_NAME} port-forward svc/kube-apiserver ${API_SERVICE_NODEPORT}:6443 --address=${LOCAL_MACHINE_IP}
  ```

  _Note_: If you want to import a managed cluster to the HyperShift hosted cluster without external access, you have to make sure the managed cluster can access the local machine.
