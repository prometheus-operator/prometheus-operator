// imageName extracts the image name from a fully qualified image string. eg.
// quay.io/coreos/addon-resizer -> addon-resizer
// grafana/grafana -> grafana
local imageName(image) =
  local parts = std.split(image, '/');
  local len = std.length(parts);
  if len == 3 then
    # registry.com/org/image
    parts[2]
  else if len == 2 then
    # org/image
    parts[1]
  else if len == 1 then
    # image, ie. busybox
    parts[0]
  else
      error 'unknown image format: ' + image;

{
  imageName:: imageName,
}
