# Graduate ScrapeConfig CRD To Beta

* Owners:
  * [mviswanathsai](https://github.com/mviswanathsai)
* Status:
  * `Accepted`
* Related Tickets:
  * [Graduate The `ScrapeConfig` CRD To `v1beta1`](https://github.com/prometheus-operator/prometheus-operator/issues/6697)
* Other docs:
  * [ScrapeConfig Design Proposal](https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/proposals/202212-scrape-config.md)
  * [Kubernetes API versioning](https://kubernetes.io/docs/reference/using-api/#api-versioning)

## Why

The goal of this proposal is to pave the way to graduate the ScrapeConfig CRD in the Prometheus Operator to beta. We aim to do this by building
a 1:1 relationship with the Prometheus `scrape_config`. By enhancing the Service Discovery support, aligning the CRD fields, and standardizing
configurations.

## Pitfalls of the Current Solution

- The CRD is currently in v1alpha1 version, this hinders its adoption among the users of Prometheus-Operator.

## Goals

- Pave the way for graduating the ScrapeConfig CRD to v1beta1 by aligning on the graduation criteria.
- Enhance the Service Discovery support by adding more Service Discovery configurations.
- Ensure all fields in Prometheus `scrape_config` are supported in the ScrapeConfig CRD.
- Maintain naming conventions and field consistency across existing and new Service Discoveries.
- Identify inconsistencies in validations.
- Outline the migration strategy from `v1alpha1` to `v1beta1`.

## Audience

- Prometheus Operator maintainers and contributors.
- Users and developers relying on the Prometheus Operator for creating monitoring configurations.
- Stakeholders interested in the evolution and improvement of the Prometheus Operator.

## Non-Goals

- Implementing the 28 Service Discovery configurations currently supported by Prometheus.
- To plan out a detailed graduation strategy.
- Convert existing monitor objects to low-level ScrapeConfig objects.

## How

In order to provide a more comprehensive and versatile monitoring solution, enhancing the
Service Discovery support in the Prometheus Operator is crucial. This will address current
limitations and better serve diverse user needs. The following steps outline our approach to achieving this goal:

### Support Statement For Service Discovery Mechanisms

By adding more Service Discovery configurations, we increase the flexibility and utility of the CRD for various user scenarios.
Kubernetes being the sole target for Prometheus-Operator, we think it is fitting to provide complete support for Kubernetes service discovery, this includes
supporting all fields present in the Prometheus configuration, example use cases and keeping the Service Discovery robustly maintained in general. We believe that the Kubernetes service discovery
along with the other existing service discoveries that we offer should be sufficient for most user's needs and so implementing all the Service discoveries
is not a priority. However, this does not imply service discovery support is limited to what exists today.
It would be a better use of time to add the service discoveries that the users need as we get feature requests for them.

The following is the list of Service Discoveries which we want to support before graduation:

- *`azure_sd_config`*
- *`consul_sd_config`*
- *`digitalocean_sd_config`*
- *`docker_sd_config`*
- *`dockerswarm_sd_config`*
- *`dns_sd_config`*
- *`ec2_sd_config`*
- *`openstack_sd_config`*
- *`puppetdb_sd_config`*
- *`file_sd_config`*
- *`gce_sd_config`*
- *`hetzner_sd_config`*
- *`http_sd_config`*
- *`kubernetes_sd_config`*
- *`kuma_sd_config`*
- *`lightsail_sd_config`*
- *`linode_sd_config`*
- *`nomad_sd_config`*
- *`eureka_sd_config`*
- *`ovhcloud_sd_config`*
- *`scaleway_sd_config`*
- *`ionos_sd_config`*

If we were to categorize the Service Discoveries based on the amount of effort we are willing to put in maintaining them:

**Tier-1:**
Project maintainers fully support the service discovery in this group. It includes the Kubernetes SD since the Operator requires a Kubernetes control plane to run. It also includes core service discoveries based on well-established protocols.
- Kubernetes Service Discovery
- File Service Discovery
- Static Config
- DNS Service Discovery
- HTTP Service Discovery

**Tier-2:**
This group includes service discoveries which are related to Kubernetes, cloud-native environments and widely used solutions. The project maintainers don't actively support them but they are happy to review issues and pull requests.
- DigitalOcean Service Discovery
- Consul Service Discovery
- Azure Service Discovery
- EC2 Service Discovery
- Lightsail Service Discovery
- Kuma Service Discovery
- OVHCloud Service Discovery
- Scaleway Service Discovery
- Ionos Service Discovery
- OpenStack Service Discovery
- GCE Service Discovery

The project maintainers do not commit to actively maintaining any service discoveries that are not listed above. We don't mean that other service discoveries are ignored, but
they are not a priority. They will be supported on a best-effort basis, meaning they will be maintained as time and resources allow, without a firm commitment from the
maintainers.

At the time of writing this document, the following Service Discoveries are not supported but may be added in the future on user requests and contributions:

- **`uyuni_sd_config`**
- **`vultr_sd_config`**

We don't plan to support the following Service Discoveries, due to them being deprecated or inactive:

- **`marathon_sd_config`**
- **`nerve_sd_config`**
- **`serverset_sd_config`**
- **`triton_sd_config`**

### Fill Existing Gaps From The Prometheus Configuration

We intend to support all fields that the Prometheus `scrape_config` contains, in the ScrapeConfig CRD. However, there might be exceptions like `job_name` for example which need to be
implemented in a slightly different manner to prevent issues related to non-unique job names.

### Improve API Consistency

The idea is to make the API as restrictive as possible against the user making wrong/redundant configurations. To that extent, the CRD contains a number
of inconsistencies both in naming conventions and field validations that need to be rectified. Some of the noted inconsistencies are:
- Missing validations on `URL` and `Host` fields.
- Missing validations on maximum and minimum value acceptable for `Port` field.
- Missing Prometheus version check for various Service Discoveries.
- Missing length validation on multiple string fields.

We propose that once the above mentioned Service discoveries which are planned to be implemented before graduation are added, we
restructur the ScrapeConfig API one Service discovery at a time to achieve a tightly knit API surface. To that extent, the general rule
of thumb will be: "Make the API as strict as possible." This allows us to lower the level of restrictions in the future if need be, whereas
the converse might not always be feasible.

Through these efforts, we aim to achieve a 1:1 relationship with the Prometheus `scrape_config` (minus the Service Discoveries), enhancing the usability and completeness of the ScrapeConfig CRD. This alignment ensures that users have access to the full range of configurations offered by Prometheus, making the Prometheus Operator a more powerful and flexible tool for monitoring and observability beyond Kubernetes.

### Graduation Strategy

#### Requirements for Graduation

We propose to graduate the CRD to beta when the following milestones are all achieved:
1. The Service Discoveries which we have listed are all supported.
2. There is consensus among the maintainers about the API consistency.
3. We are confident about the completeness of the test cases coverage for the API.

#### Path for Graduation

From past experience with the graduation of the `AlertmanagerConfig` CRD, we believe that the cost of implementing and maintaining a conversion webhook is too much to bear
and we would like to avoid it when possible.
Keeping this in mind, we recommend that we make all the breaking changes in `v1alpha` and
once there is a consensus in the community about the "readiness"/"completeness" of the`v1alpha1`, we graduate the ScrapeConfig CRD to `v1beta1`.
Note: In this strategy, both the `v1alpha1` and `v1beta1` APIs are expected to be identical to eachother, thus barring the need for a conversion webhook.

From the [Kuberenetes CRD docs](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definition-versioning/#specify-multiple-versions),
the `CustomResourceDefinition` would contain the following lines:

```
  conversion:
    # None conversion assumes the same schema for all versions and only sets the apiVersion
    # field of custom resources to the proper value
    strategy: None
```

### Testing and Verification

- **Covering All Test Cases For Kubernetes Service Discovery**: Since Kubernetes is our main player, make sure all testcases for unit tests and e2e tests have been covered for the Kubernetes Service Discovery.

- **Implement Comprehensive Unit Tests**: Ensure that unit tests are added for all new and existing Service Discovery configurations to ensure that the expected configuration is generated and validations are in place.

## Alternatives

- **Introduce v1alpha2 Before Beta**
  - Strategy: Create a `v1alpha2` version incorporating all necessary breaking changes and refinements. Transition from `v1alpha2` to `v1beta1` when ready.
  - Con: additional complexity for users without real benefit.

- **Implement v1alpha1 to v1beta1 Conversion Webhook**
  - Strategy: Graduate to `v1beta1` from `v1alpha1` with all the necessary changes/improvements and handle any breaking changes in the API between the two versions
    with a conversion webhook. This ensures that users can automatically transition their configurations without manual intervention.
  - Con: Greatly increases complexity for the maintainers as well as the users.
