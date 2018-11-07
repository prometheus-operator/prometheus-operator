local kp = import 'kube-prometheus/kube-prometheus.libsonnet';
local l = import 'kube-prometheus/lib/lib.libsonnet';
local config = kp._config;

local makeImages(config) = [
  {
    name: config.imageRepos[image],
    tag: config.versions[image],
  }
  for image in std.objectFields(config.imageRepos)
];

local upstreamImage(image) = '%s:%s' % [image.name, image.tag];
local downstreamImage(registry, image) = '%s/%s:%s' % [registry, l.imageName(image.name), image.tag];

local pullPush(image, newRegistry) = [
  'docker pull %s' % upstreamImage(image),
  'docker tag %s %s' % [upstreamImage(image), downstreamImage(newRegistry, image)],
  'docker push %s' % downstreamImage(newRegistry, image),
];

local images = makeImages(config);

local output(repository) = std.flattenArrays([
  pullPush(image, repository)
  for image in images
]);

function(repository='my-registry.com/repository')
  std.join('\n', output(repository))
