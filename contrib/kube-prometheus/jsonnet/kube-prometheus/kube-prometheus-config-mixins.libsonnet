local l = import 'lib/lib.libsonnet';

// withImageRepository is a mixin that replaces all images prefixes by repository. eg.
// quay.io/coreos/addon-resizer -> $repository/addon-resizer
// grafana/grafana -> grafana $repository/grafana
local withImageRepository(repository) = {
  local oldRepos = super._config.imageRepos,
  local substituteRepository(image, repository) =
    if repository == null then image else repository + '/' + l.imageName(image),
  _config+:: {
    imageRepos:: {
      [field]: substituteRepository(oldRepos[field], repository),
      for field in std.objectFields(oldRepos)
    }
  },
};

{
  withImageRepository:: withImageRepository,
}
