local v = importstr '../../VERSION';

{
  namespace: 'default',
  version: std.strReplace(v, '\n', ''),
  image: 'quay.io/prometheus-operator/prometheus-operator:v' + self.version,
  configReloaderImage: 'quay.io/prometheus-operator/prometheus-config-reloader:v' + self.version,
}
