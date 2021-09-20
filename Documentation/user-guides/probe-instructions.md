## Probe

The `Probe` custom resource definition (CRD) allows to declarative define how groups of ingresses and static targets should be monitored. Besides the target, the `Probe` object requires a `prober` which is the service that monitors the target and provides metrics for Prometheus to scrape. The following document describes one of the application scenarios.



## Usage scenario

+ HTTPS certificate needs to be configured
+ Systems outside the k8s cluster, such as etcd
+ Static target

## Example

+ Prepare SSL certificate file, execute the following command and generate related resources

~~~shell
kubectl create secret generic etcd-ssl-ca --from-file=etcd-ssl/ca.pem --dry-run -oyaml > etcd-ssl.yaml
kubectl apply -f etcd-ssl.yaml -n monitoring

kubectl create secret generic etcd-ssl-cert --from-file=etcd-ssl/etcd.pem --dry-run -oyaml > etcd-ssl.yaml
kubectl apply -f etcd-ssl.yaml -n monitoring

kubectl create secret generic etcd-ssl-key --from-file=etcd-ssl/etcd-key.pem  --dry-run -oyaml > etcd-ssl.yaml
kubectl apply -f etcd-ssl.yaml -n monitoring
~~~

+ Create the following file, such as `probe-etcd.yaml`

~~~yaml
apiVersion: monitoring.coreos.com/v1
kind: Probe
metadata:
  labels:
    release: static  # need set ProbeSelector in Proemtheus CRD
  name: etcd
  namespace: monitoring
spec:
  jobName: etcd
  tlsConfig:
    ca:
      secret:
        key: ca.pem
        name: etcd-ssl-ca
    cert:
      secret:
        key: etcd.pem
        name: etcd-ssl-cert
    keySecret:
      key: etcd-key.pem
      name: etcd-ssl-key
  prober:
    url: "1.1.3.1:2379"   # one of static target is ok 
    path: "/metrics"
    scheme: "https"
  targets:
    staticConfig:
      relabelingConfigs:   
        - sourceLabels: [ __param_target ]
          separator: ;
          regex: (.*)
          targetLabel: __address__
          replacement: $1
          action: replace
      static:
        - 1.1.3.1:2379  
        - 1.1.3.2:2379
        - 1.1.3.3:2379
~~~

+ Execute the command to generate the `Probe` resource

~~~sh
kubectl apply -f probe-etcd.yaml
~~~

Wait for Prometheus to refresh the configuration file, and you will see your configuration take effect in target.

