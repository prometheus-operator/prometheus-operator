# Revamping Documentation Structure

- Owners:
  - [AshwinSriram11](https://github.com/AshwinSriram11)
- Related Tickets:
  - [#3553](https://github.com/prometheus-operator/prometheus-operator/issues/3553#issuecomment-726733177)
  - [#6046](https://github.com/prometheus-operator/prometheus-operator/issues/6046)
- Other docs:
  - N/A

This document proposes restructuring the documentation of the Prometheus-Operator website, aiming for better content organization and better experience for new users and old veterans.

# Why

Restructuring the documentation will help in improving user experience and may save time for a user to search for relevant information effectively. This will encourage a newcomer to get familiar with the project in an efficient manner. The end goal of a good documentation is to follow the industrial best practices and provide accurate information to a user.

I believe that by restructuring, we can group the related topics together in an organized way and use cross-references to link topics that are difficult to group but are relevant for understanding. This makes the documentation more user-friendly and also ensures uniform flow of content.

After we have a proper structure, it will become relatively easy to add information about new features. Maintainers will be able to save time on deciding the best place for adding new content. Also, improving documentation will help in improving Search Engine Optimization(SEO) and increase the engagement of the project.

# Pitfalls of the current solution

A good documentation is one that is easy to understand for a newcomer and provides the exact amount of information that is needed according to the need. But looking at the current documentation structure, a lot of topics seem misplaced. For example, there is no need for a **"Contributing"** page in the prologue section. Prologue section should only give the introduction and the prerequisites for the project. Due to this, a user might need to put more effort to search for relevant information and this might decrease the user's productivity.

Currently, there is some unnecessary information from the documentation creating misconceptions in the user's mind. For example, let us look at [#6046](https://github.com/prometheus-operator/prometheus-operator/issues/6046) which tells us that there is no Ingress Guide present in the current documentation. But, if we look at the website, there are links to the **“Ingress Guide”** on the [Getting-Started](https://prometheus-operator.dev/docs/user-guides/getting-started/#exposing-the-prometheus-service) and [Alerting page](https://prometheus-operator.dev/docs/user-guides/alerting/#exposing-the-alertmanager-service) in **User-Guide**. Due to this, many users will report the same issue and it will take time for maintainers to resolve them.

Incorporation of new topics is also difficult if the structure is not up to mark because more time and effort is needed to decide the best place to add a topic which can often lead to decrease in productivity of a maintainer. For example, in issue [#3553](https://github.com/prometheus-operator/prometheus-operator/issues/3553#issuecomment-726733177), it has been mentioned that basic architecture needs to be worked upon before adding the diagram as there is no section talking about **“namespace selection”** in the current documentation.

# Goals

- To improve readability and make the documentation easier to understand.
- To ensure a uniform flow of content.
- To remove irrelevant information from the documentation.
- To ensure that all new features get documented properly in the future.

# Audience

- Developers who rely on Prometheus Operator to facilitate the deployment, management and monitoring of Prometheus instances in their environment.
- Developers who want to set up alerts for their application.
- Administrators who want to monitor Prometheus and Alertmanager instances using Kubernetes tools.
- Maintainers who are responsible for optimizing Prometheus Operator.
- Contributors who want to engage and want to contribute effectively to the project.

# Non-Goals

- Adding extra documentation about new features, enhancements, troubleshooting, or other information that might be missing today.

# How

I believe a basic structure for the website should be as given below:

### 1. Getting Started

- **Introduction** - This section will introduce us to Prometheus-Operator and will talk about goals and the problems that are being solved by this project.
- **Installation Guide** - This section will contain all the methods of installation of Prometheus-Operator in the production environment.
- **Compatibility** - This section will provide information about compatibility of Prometheus-Operator with Kubernetes, Prometheus, Alertmanager and Thanos.
- **Design** - This section will describe the design of Prometheus-Operator focusing on the various custom resource definitions(CRDs) it manages.

### 2. API Reference

This section will provide detailed information about different fields of the Custom Resources in Prometheus-Operator(config, spec, status and other information).

### 3. User Guide

This section will provide focused information for developers who want to integrate Prometheus and Alertmanager using Prometheus-Operator in their applications along with examples and troubleshooting.

### 4. Administrator Guide

This section will provide deeper insights for administrators who want to set up a robust, secure monitoring infrastructure.

### 5. Community

- **Contributing**
- **Testing**
- **DCO**
- **License**
- **Code of Conduct**
- **Release**
- **Adopters**

The reasoning behind the new structure is given below -

- The **“Getting-Started”** section should only serve information on how to begin your journey with the project. So, the first thing that comes to mind should be a basic overview followed by the installation. It won’t make sense to have the installation section appear in the User-Guide(in the current model) because installation is the foremost important thing to begin working on a project. With installation, users would also like to know the different releases and their compatibility, thus it makes more sense to add this topic just below the installation section. At last, Design doc should be present to give an overview of how different components interact and what their purpose is in a concise manner.

- API Reference is of extreme importance and is something needed by developers to implement the functionalities they need. It is something needed by both a **User** and an **Administrator**. As it is common for both personas, I think rather than duplicating this page in both sections, we should have a separate section for it due to its relevance. Further, looking at the current API reference page, it contains too much information for a single page. This makes the page look overly crowded and searching for any particular field becomes a tedious task. We can convert this page into multiple pages by grouping sections and cross-reference properly which will provide a better experience for a user.

- I felt that User-Guide is placed before Administrator-Guide keeping in mind about the difficulties a newcomer faces when he/she starts working on a project. User-Guide contains beginner friendly examples and provides a beginner level experience. The topics discussed are mostly centered around how to work with different CRDs. But when we move on to the Administrator-Guide, we will find more advanced situations like security, storage which would be difficult to understand if a person is not able to cope up with basic concepts of different CRDs.
  Thus, this structure provides a uniform flow with context to difficulty level for a newcomer as well as segregates relevant information according to the persona.

- As mentioned in the audience, we also have people who want to contribute to the project. So, making a **“Community”** section serves the purpose of providing users to find how to contribute to the project. Also, it provides other community related information like release, adopters, and other topics which are not provided in the current model.

- We should remove the Kube-Prometheus section in the current model. We can mention the method of deploying Prometheus-Operator with Kube-Prometheus in the installation section due to its ease and popularity among users. But, as the helm-charts are not being maintained by Prometheus-Operator and have now been moved to Kube-Prometheus-Stack, it becomes difficult to manage and address the issues regarding Kube-Prometheus.

# Action Plan

1. Making sections and organizing files as described above.
2. If needed, splitting the content of one file into multiple files to improve readability or changing the location of content wherever required.
3. Removing content that is unnecessary for the project.

# Follow-Ups

After we successfully complete the Goals of this proposal, we can move on with the Non-Goals section. As mentioned, the next step would be to add the documentation for the topics that are currently missing.
