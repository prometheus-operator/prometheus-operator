# Prometheus Operator CLI

- Owners:
  - [Nicolas Takashi](https://github.com/nicolastakashi)
- Related Tickets
  - [#6423](https://github.com/prometheus-operator/prometheus-operator/issues/6423)
- Other docs:
  - N/A

This document proposes a new CLI tool to manage Prometheus Operator resources, allowing users to create, update, and delete Prometheus Operator resources using a simple CLI interface as well as troubleshoot and debug Prometheus Operator resources.

# Why

Throughout the years, we received feedback from users (through workshops, support tickets, slack, email, etc) that managing Prometheus Operator resources is difficult and error-prone. Users reported that they often encounter issues when creating, updating, and deleting Prometheus Operator resources, and that troubleshooting and debugging Prometheus Operator resources is challenging.

# Pitfalls of the current solution

At the moment people are struggling to manage Prometheus Operator resources, they have to manually create, update, and delete Prometheus Operator resources using `kubectl` or other tools, and troubleshooting and debugging Prometheus Operator resources is challenging.

People were struggling to troubleshoot why their Prometheus pods were not being created. They were not sure if the issue was with the Prometheus Operator or with the Prometheus resource itself.
After long investigation, they found out that the Prometheus Operator were only watching the namespace where the Prometheus Operator was deployed, and the Prometheus resource was being created in a different namespace.

Troubleshooting Alertmanager, Prometheus and Thanos Ruler pod creation is a manual work and requires domain knowledge about what RBAC permissions are needed, below you can find few examples of the issues that users reported:
- https://github.com/prometheus-operator/prometheus-operator/issues/5874
- https://github.com/prometheus-operator/prometheus-operator/issues/2641

Target discovery was also a common issue, users were not sure how to configure the Prometheus to discover the targets that they wanted to monitor, below you can find few examples of the issues that users reported:
- https://github.com/prometheus-operator/prometheus-operator/issues/4428
- https://github.com/prometheus-operator/prometheus-operator/issues/4701
- https://github.com/prometheus-operator/prometheus-operator/issues/5386

# Goals

The main goal of this proposal is to reduce the manual effort and domain knowledge needed to manage and operate Prometheus Operator resources.

# Non-Goals

- Replace any existing config management tools such as Kustomize or Helm.

# Audience

- Platform engineers/DevOps/SREs that want to extend their CI validations to prometheus-operator resources.
- Beginners trying out Prometheus-Operator for the first time.
- Users that prefer a CLI-focused approach to manage and troubleshoot prometheus-operator resources.

# How

We propose to create a new CLI tool allowing users to create, update, and delete Prometheus Operator resources using a simple CLI interface as well as troubleshoot and debug Prometheus Operator resources. As a source of inspiration, the new cli could work as the example below, for different commands:

## Deploy a Kube-Prometheus Stack

The CLI could provide a command to deploy a Kube-Prometheus stack, where users can deploy the Prometheus Operator, Alertmanager, Node-Exporter and others with a single command.

```bash
poctl create stack --version 0.81.0
```

## Analyze

The CLI could provide an analyze command, where users can check if the Prometheus Operator resources are correctly configured, the troubleshoot command will analyze the given resource and provide suggestions on how to fix the issues.

```bash
poctl analyze --kind prometheus --name prometheus --namespace monitoring
```

## Where the code will live?

The prometheus-operator-cli code will be placed in a dedicated repository under the [Prometheus-Operator organization](https://github.com/prometheus-operator), this will allow us to have a separate lifecycle for the Prometheus Operator and the Prometheus Operator CLI and the following benefits:

- We can have a dedicated team to maintain the Prometheus Operator CLI
- We can use the Prometheus Operator client as a customer and be more careful with the changes that we make in the Prometheus Operator codebase
- We can have a dedicated release cycle for the Prometheus Operator CLI

### Linting and Validation

Allow users to validate Prometheus Operator resources before creating them, the CLI should check if the resource manifest is valid.

This is especially useful for CI pipelines, where users can validate the Prometheus Operator resources before applying them to the cluster.

# Action Plan

1. Create a new repository under the Prometheus-Operator organization.
2. Create a new CLI startup project.
3. Implement scaffolding feature for the main Prometheus Operator resources.
4. Implement linting and validation feature.
5. Implement an interactive mode for the scaffolding feature.
