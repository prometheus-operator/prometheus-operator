# Graduate ScrapeConfig CRD To Beta

* Owners:
  * [mviswanathsai](https://github.com/mviswanathsai)
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

## Audience

- Prometheus Operator maintainers and contributors.
- Users and developers relying on the Prometheus Operator for creating monitoring configurations.
- Stakeholders interested in the evolution and improvement of the Prometheus Operator.

## Non-Goals

- Implementing the 28 Service Discovery configurations currently supported by Prometheus.
- To plan out a detailed graduation strategy.

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

- We will incorporate all fields present in Prometheus `scrape_config` into the ScrapeConfig CRD. However, there are specific exceptions:
  - **`job_name`**: To prevent issues related to non-unique job names, we have introduced an equivalent `JobName` field, which ensures safer and more predictable configurations.
  - **`body_size_limit`**: This field is managed globally through `CommonPrometheusFields`. Additionally, users can set `body_size_limit` at the ScrapeConfig CRD level for specific scrape jobs, providing flexibility without redundancy.
- Include missing fields such as:
  - `scrape_classic_histograms`
  - `follow_redirects`
  - `enable_http`
  - `oauth2`
  - `native_histogram_bucket_limit`
  - `native_histogram_min_bucket_factor`

### Improve API Consistency

The idea is to make the API as restrictive as possible against the user making wrong/redundant manifests. To that extent, the CRD contains a number
of inconsistencies, both in naming conventions and in field validations that need to be rectified. Some of the noted inconsistencies are:
- Missing validations on `URL`, `Host` fields.
- Missing validations on maximum and minimum int value acceptable for `Port` field.
- Missing Prometheus version check for various Service Discoveries.
- Missing length validation on multiple string fields.
- Multiple `Filter` fields present with identical code.

We propose that once the above mentioned Service discoveries which are planned to be implemented before graduation are added, we will start
restructuring the ScrapeConfig API one Service discovery at a time to achieve a tightly knit API surface. To that extent, the general rule
of thumb will be: "Make the API as strict as possible." This allows us to lower the level of restrictions in the future if need be, whereas
the converse might not always be feasible.

Through these efforts, we aim to achieve a 1:1 relationship with the Prometheus `scrape_config` (minus the Service Discoveries), enhancing the usability and completeness of the ScrapeConfig CRD. This alignment ensures that users have access to the full range of configurations offered by Prometheus, making the Prometheus Operator a more powerful and flexible tool for monitoring and observability.

### Graduation Strategy

#### Requirements for Graduation

We propose to graduate the CRD to beta when the following milestones are all achieved:
1. The Service Discoveries which we have listed are all supported.
2. There is consensus among the maintainers about the API consistency.
3. We are confident about the completeness of the test cases coverage for the API.

#### Path for Graduation

*Note that this is not meant to be a definitive, complete path for graduation. Rather, it can be viewed at as a discussion of the possible strategies.*

From past experience with the Alertmanager CRD, we suggest to avoid the implementation of a conversion webhook
if possible. To that extent, we feel that the option 1, which suggests that we move from `v1alpha1` -> `v1beta1` without a conversion webhook might
be a good fit. Once we are confident that the API has all the missing fields mentioned above and any Service discoveries mentioned above,
with necessary validations in place and void of any apparent inconsistencies, we will transition to `v1beta1`.

We can go a step further and introduce breaking changes (if any) in v1alpha1 and then just copy it over to v1beta1. This way, we do not have to
worry about conversion webhooks and the user does not need to perform any deliberate migration from `v1alpha1` to `v1beta1`. Further, currently the CRD is in alpha stage
and we believe that it is fine for us to introduce breaking changes in this stage and trust the user handle it than to introduce more complexity
into the code base.

This graduation strategy ensures a balanced approach, allowing us to refine the API while preparing for a more stable and well-supported v1beta1 release.

### Testing and Verification

- **Covering All Test Cases For Kubernetes Service Discovery**: Since Kubernetes is our main player, make sure all testcases for unit tests and e2e tests have been covered for the Kubernetes Service Discovery.

- **Implement Comprehensive Unit Tests**: Ensure that unit tests are added for all new and existing Service Discovery configurations to ensure that the expected configuration is generated and validations are in place.

### Miscellaneous Enhancements

1. **Consolidation of Monitor Resources**:
   - Explore the possibility of consolidating Monitor resources into ScrapeConfig instances, using feature gates to
     manage the transition. This approach aims to create a single source of truth for configuration generation, simplifying
     management and improving consistency. This is a change which affects a huge part of the user base, a separate design proposal
     and more time to investigate options would be required. It is not a primary focus for now, but something which we aim to get
     to once the main objectives of this proposal seem close.
     The work on this can be started either in `v1alpha1` or `v1beta1` stages with the following morale:
     1. In `v1beta1`, the API is expected to be fairly complete with all the features and fixes mentioned above. At this stage, it might
        be a good idea to play with the consolidation logic behind a feature gate.
     2. In `v1alpha1`, the API is still nascent as a result we have more freedom in introducing breaking changes to the API, if need be, for the
        consolidation logic (which we believe would need a design proposal of its own). Whereas, after graduation, we might have to introduce a conversion webhook
        for the same, but it is just speculation and we believe it is too early to judge this now.

2. **Quality-of-Life Features**:
   - Introduce additional features that enhance usability, such as frequently used relabeling configurations like metadata attachment
     for Kubernetes Service Discovery using the `attachMetadata` field. These enhancements are not necessary but aim to make it easier
     for users to configure and manage their monitoring setups.

## Alternatives

- **Direct Graduation to Beta**
  - Strategy: Move directly from `v1alpha1` to `v1beta1` without any intermediate steps.
  - Con: additional burden on maintainers since we may have to live with a sub-optimal API if not dealt with sufficient thoroughness.

- **No API Changes in Beta**
  - Strategy: Ensure no API changes from `v1alpha1` to `v1beta1`, avoiding the need for a conversion webhook.
  - Con: Might not be possible since new fields have been added since `v1alpha1` of the API was introduced.

- **Introduce v1alpha2 Before Beta**
  - Strategy: Create a `v1alpha2` version incorporating all necessary breaking changes and refinements. Transition from `v1alpha2` to `v1beta1` when ready.
  - Con: additional complexity for users without real benefit.

- **Implement v1alpha1 to v1beta1 Conversion Webhook**
  - Strategy: Graduate directly to `v1beta1` from `v1alpha1` and handle any breaking changes through a conversion webhook. This ensures that users can automatically transition their configurations without manual intervention.
  - Con: Greatly increases complexity for the maintainers as well as the users.

- **Avoid v1alpha1 to v1beta1 Conversion Webhook**
  - Strategy: Graduate directly to `v1beta1` and require users to manually handle any breaking changes or conversion tasks. Provide detailed documentation and support to guide users through the transition process.
  - Con: Additional complexity for existing users of the API.
