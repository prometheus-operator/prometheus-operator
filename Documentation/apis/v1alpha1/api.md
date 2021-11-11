# API Reference

Packages:

- [monitoring.coreos.com/v1alpha1](#monitoringcoreoscomv1alpha1)

# monitoring.coreos.com/v1alpha1

Resource Types:

- [AlertmanagerConfig](#alertmanagerconfig)




## AlertmanagerConfig
<sup><sup>[↩ Parent](#monitoringcoreoscomv1alpha1 )</sup></sup>






AlertmanagerConfig defines a namespaced AlertmanagerConfig to be aggregated across multiple namespaces configuring one Alertmanager cluster.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>monitoring.coreos.com/v1alpha1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>AlertmanagerConfig</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspec">spec</a></b></td>
        <td>object</td>
        <td>
          AlertmanagerConfigSpec is a specification of the desired behavior of the Alertmanager configuration. By definition, the Alertmanager configuration only applies to alerts for which the `namespace` label is equal to the namespace of the AlertmanagerConfig resource.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec
<sup><sup>[↩ Parent](#alertmanagerconfig)</sup></sup>



AlertmanagerConfigSpec is a specification of the desired behavior of the Alertmanager configuration. By definition, the Alertmanager configuration only applies to alerts for which the `namespace` label is equal to the namespace of the AlertmanagerConfig resource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecinhibitrulesindex">inhibitRules</a></b></td>
        <td>[]object</td>
        <td>
          List of inhibition rules. The rules will only apply to alerts matching the resource’s namespace.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindex">receivers</a></b></td>
        <td>[]object</td>
        <td>
          List of receivers.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecroute">route</a></b></td>
        <td>object</td>
        <td>
          The Alertmanager route definition for alerts matching the resource’s namespace. If present, it will be added to the generated Alertmanager configuration as a first-level route.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.inhibitRules[index]
<sup><sup>[↩ Parent](#alertmanagerconfigspec)</sup></sup>



InhibitRule defines an inhibition rule that allows to mute alerts when other alerts are already firing. See https://prometheus.io/docs/alerting/latest/configuration/#inhibit_rule

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>equal</b></td>
        <td>[]string</td>
        <td>
          Labels that must have an equal value in the source and target alert for the inhibition to take effect.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecinhibitrulesindexsourcematchindex">sourceMatch</a></b></td>
        <td>[]object</td>
        <td>
          Matchers for which one or more alerts have to exist for the inhibition to take effect. The operator enforces that the alert matches the resource’s namespace.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecinhibitrulesindextargetmatchindex">targetMatch</a></b></td>
        <td>[]object</td>
        <td>
          Matchers that have to be fulfilled in the alerts to be muted. The operator enforces that the alert matches the resource’s namespace.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.inhibitRules[index].sourceMatch[index]
<sup><sup>[↩ Parent](#alertmanagerconfigspecinhibitrulesindex)</sup></sup>



Matcher defines how to match on alert's labels.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Label to match.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>regex</b></td>
        <td>boolean</td>
        <td>
          Whether to match on equality (false) or regular-expression (true).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Label value to match.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.inhibitRules[index].targetMatch[index]
<sup><sup>[↩ Parent](#alertmanagerconfigspecinhibitrulesindex)</sup></sup>



Matcher defines how to match on alert's labels.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Label to match.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>regex</b></td>
        <td>boolean</td>
        <td>
          Whether to match on equality (false) or regular-expression (true).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Label value to match.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index]
<sup><sup>[↩ Parent](#alertmanagerconfigspec)</sup></sup>



Receiver defines one or more notification integrations.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the receiver. Must be unique across all items from the list.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexemailconfigsindex">emailConfigs</a></b></td>
        <td>[]object</td>
        <td>
          List of Email configurations.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexopsgenieconfigsindex">opsgenieConfigs</a></b></td>
        <td>[]object</td>
        <td>
          List of OpsGenie configurations.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpagerdutyconfigsindex">pagerdutyConfigs</a></b></td>
        <td>[]object</td>
        <td>
          List of PagerDuty configurations.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpushoverconfigsindex">pushoverConfigs</a></b></td>
        <td>[]object</td>
        <td>
          List of Pushover configurations.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexslackconfigsindex">slackConfigs</a></b></td>
        <td>[]object</td>
        <td>
          List of Slack configurations.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexvictoropsconfigsindex">victoropsConfigs</a></b></td>
        <td>[]object</td>
        <td>
          List of VictorOps configurations.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexwebhookconfigsindex">webhookConfigs</a></b></td>
        <td>[]object</td>
        <td>
          List of webhook configurations.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexwechatconfigsindex">wechatConfigs</a></b></td>
        <td>[]object</td>
        <td>
          List of WeChat configurations.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].emailConfigs[index]
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindex)</sup></sup>



EmailConfig configures notifications via Email.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>authIdentity</b></td>
        <td>string</td>
        <td>
          The identity to use for authentication.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexemailconfigsindexauthpassword">authPassword</a></b></td>
        <td>object</td>
        <td>
          The secret's key that contains the password to use for authentication. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexemailconfigsindexauthsecret">authSecret</a></b></td>
        <td>object</td>
        <td>
          The secret's key that contains the CRAM-MD5 secret. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>authUsername</b></td>
        <td>string</td>
        <td>
          The username to use for authentication.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>from</b></td>
        <td>string</td>
        <td>
          The sender address.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexemailconfigsindexheadersindex">headers</a></b></td>
        <td>[]object</td>
        <td>
          Further headers email header key/value pairs. Overrides any headers previously set by the notification implementation.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>hello</b></td>
        <td>string</td>
        <td>
          The hostname to identify to the SMTP server.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>html</b></td>
        <td>string</td>
        <td>
          The HTML body of the email notification.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>requireTLS</b></td>
        <td>boolean</td>
        <td>
          The SMTP TLS requirement. Note that Go does not support unencrypted connections to remote SMTP endpoints.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sendResolved</b></td>
        <td>boolean</td>
        <td>
          Whether or not to notify about resolved alerts.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>smarthost</b></td>
        <td>string</td>
        <td>
          The SMTP host and port through which emails are sent. E.g. example.com:25<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>text</b></td>
        <td>string</td>
        <td>
          The text body of the email notification.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexemailconfigsindextlsconfig">tlsConfig</a></b></td>
        <td>object</td>
        <td>
          TLS configuration<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>to</b></td>
        <td>string</td>
        <td>
          The email address to send notifications to.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].emailConfigs[index].authPassword
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexemailconfigsindex)</sup></sup>



The secret's key that contains the password to use for authentication. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].emailConfigs[index].authSecret
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexemailconfigsindex)</sup></sup>



The secret's key that contains the CRAM-MD5 secret. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].emailConfigs[index].headers[index]
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexemailconfigsindex)</sup></sup>



KeyValue defines a (key, value) tuple.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          Key of the tuple.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Value of the tuple.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].emailConfigs[index].tlsConfig
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexemailconfigsindex)</sup></sup>



TLS configuration

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexemailconfigsindextlsconfigca">ca</a></b></td>
        <td>object</td>
        <td>
          Struct containing the CA cert to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexemailconfigsindextlsconfigcert">cert</a></b></td>
        <td>object</td>
        <td>
          Struct containing the client cert file for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>insecureSkipVerify</b></td>
        <td>boolean</td>
        <td>
          Disable target certificate validation.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexemailconfigsindextlsconfigkeysecret">keySecret</a></b></td>
        <td>object</td>
        <td>
          Secret containing the client key file for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>serverName</b></td>
        <td>string</td>
        <td>
          Used to verify the hostname for the targets.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].emailConfigs[index].tlsConfig.ca
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexemailconfigsindextlsconfig)</sup></sup>



Struct containing the CA cert to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexemailconfigsindextlsconfigcaconfigmap">configMap</a></b></td>
        <td>object</td>
        <td>
          ConfigMap containing data to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexemailconfigsindextlsconfigcasecret">secret</a></b></td>
        <td>object</td>
        <td>
          Secret containing data to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].emailConfigs[index].tlsConfig.ca.configMap
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexemailconfigsindextlsconfigca)</sup></sup>



ConfigMap containing data to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key to select.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the ConfigMap or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].emailConfigs[index].tlsConfig.ca.secret
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexemailconfigsindextlsconfigca)</sup></sup>



Secret containing data to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].emailConfigs[index].tlsConfig.cert
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexemailconfigsindextlsconfig)</sup></sup>



Struct containing the client cert file for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexemailconfigsindextlsconfigcertconfigmap">configMap</a></b></td>
        <td>object</td>
        <td>
          ConfigMap containing data to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexemailconfigsindextlsconfigcertsecret">secret</a></b></td>
        <td>object</td>
        <td>
          Secret containing data to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].emailConfigs[index].tlsConfig.cert.configMap
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexemailconfigsindextlsconfigcert)</sup></sup>



ConfigMap containing data to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key to select.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the ConfigMap or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].emailConfigs[index].tlsConfig.cert.secret
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexemailconfigsindextlsconfigcert)</sup></sup>



Secret containing data to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].emailConfigs[index].tlsConfig.keySecret
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexemailconfigsindextlsconfig)</sup></sup>



Secret containing the client key file for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].opsgenieConfigs[index]
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindex)</sup></sup>



OpsGenieConfig configures notifications via OpsGenie. See https://prometheus.io/docs/alerting/latest/configuration/#opsgenie_config

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexopsgenieconfigsindexapikey">apiKey</a></b></td>
        <td>object</td>
        <td>
          The secret's key that contains the OpsGenie API key. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>apiURL</b></td>
        <td>string</td>
        <td>
          The URL to send OpsGenie API requests to.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>description</b></td>
        <td>string</td>
        <td>
          Description of the incident.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexopsgenieconfigsindexdetailsindex">details</a></b></td>
        <td>[]object</td>
        <td>
          A set of arbitrary key/value pairs that provide further detail about the incident.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexopsgenieconfigsindexhttpconfig">httpConfig</a></b></td>
        <td>object</td>
        <td>
          HTTP client configuration.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          Alert text limited to 130 characters.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>note</b></td>
        <td>string</td>
        <td>
          Additional alert note.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>priority</b></td>
        <td>string</td>
        <td>
          Priority level of alert. Possible values are P1, P2, P3, P4, and P5.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexopsgenieconfigsindexrespondersindex">responders</a></b></td>
        <td>[]object</td>
        <td>
          List of responders responsible for notifications.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sendResolved</b></td>
        <td>boolean</td>
        <td>
          Whether or not to notify about resolved alerts.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>source</b></td>
        <td>string</td>
        <td>
          Backlink to the sender of the notification.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>tags</b></td>
        <td>string</td>
        <td>
          Comma separated list of tags attached to the notifications.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].opsgenieConfigs[index].apiKey
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexopsgenieconfigsindex)</sup></sup>



The secret's key that contains the OpsGenie API key. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].opsgenieConfigs[index].details[index]
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexopsgenieconfigsindex)</sup></sup>



KeyValue defines a (key, value) tuple.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          Key of the tuple.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Value of the tuple.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].opsgenieConfigs[index].httpConfig
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexopsgenieconfigsindex)</sup></sup>



HTTP client configuration.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexopsgenieconfigsindexhttpconfigauthorization">authorization</a></b></td>
        <td>object</td>
        <td>
          Authorization header configuration for the client. This is mutually exclusive with BasicAuth and is only available starting from Alertmanager v0.22+.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexopsgenieconfigsindexhttpconfigbasicauth">basicAuth</a></b></td>
        <td>object</td>
        <td>
          BasicAuth for the client. This is mutually exclusive with Authorization. If both are defined, BasicAuth takes precedence.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexopsgenieconfigsindexhttpconfigbearertokensecret">bearerTokenSecret</a></b></td>
        <td>object</td>
        <td>
          The secret's key that contains the bearer token to be used by the client for authentication. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>proxyURL</b></td>
        <td>string</td>
        <td>
          Optional proxy URL.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexopsgenieconfigsindexhttpconfigtlsconfig">tlsConfig</a></b></td>
        <td>object</td>
        <td>
          TLS configuration for the client.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].opsgenieConfigs[index].httpConfig.authorization
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexopsgenieconfigsindexhttpconfig)</sup></sup>



Authorization header configuration for the client. This is mutually exclusive with BasicAuth and is only available starting from Alertmanager v0.22+.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexopsgenieconfigsindexhttpconfigauthorizationcredentials">credentials</a></b></td>
        <td>object</td>
        <td>
          The secret's key that contains the credentials of the request<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          Set the authentication type. Defaults to Bearer, Basic will cause an error<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].opsgenieConfigs[index].httpConfig.authorization.credentials
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexopsgenieconfigsindexhttpconfigauthorization)</sup></sup>



The secret's key that contains the credentials of the request

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].opsgenieConfigs[index].httpConfig.basicAuth
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexopsgenieconfigsindexhttpconfig)</sup></sup>



BasicAuth for the client. This is mutually exclusive with Authorization. If both are defined, BasicAuth takes precedence.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexopsgenieconfigsindexhttpconfigbasicauthpassword">password</a></b></td>
        <td>object</td>
        <td>
          The secret in the service monitor namespace that contains the password for authentication.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexopsgenieconfigsindexhttpconfigbasicauthusername">username</a></b></td>
        <td>object</td>
        <td>
          The secret in the service monitor namespace that contains the username for authentication.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].opsgenieConfigs[index].httpConfig.basicAuth.password
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexopsgenieconfigsindexhttpconfigbasicauth)</sup></sup>



The secret in the service monitor namespace that contains the password for authentication.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].opsgenieConfigs[index].httpConfig.basicAuth.username
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexopsgenieconfigsindexhttpconfigbasicauth)</sup></sup>



The secret in the service monitor namespace that contains the username for authentication.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].opsgenieConfigs[index].httpConfig.bearerTokenSecret
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexopsgenieconfigsindexhttpconfig)</sup></sup>



The secret's key that contains the bearer token to be used by the client for authentication. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].opsgenieConfigs[index].httpConfig.tlsConfig
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexopsgenieconfigsindexhttpconfig)</sup></sup>



TLS configuration for the client.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexopsgenieconfigsindexhttpconfigtlsconfigca">ca</a></b></td>
        <td>object</td>
        <td>
          Struct containing the CA cert to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexopsgenieconfigsindexhttpconfigtlsconfigcert">cert</a></b></td>
        <td>object</td>
        <td>
          Struct containing the client cert file for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>insecureSkipVerify</b></td>
        <td>boolean</td>
        <td>
          Disable target certificate validation.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexopsgenieconfigsindexhttpconfigtlsconfigkeysecret">keySecret</a></b></td>
        <td>object</td>
        <td>
          Secret containing the client key file for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>serverName</b></td>
        <td>string</td>
        <td>
          Used to verify the hostname for the targets.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].opsgenieConfigs[index].httpConfig.tlsConfig.ca
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexopsgenieconfigsindexhttpconfigtlsconfig)</sup></sup>



Struct containing the CA cert to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexopsgenieconfigsindexhttpconfigtlsconfigcaconfigmap">configMap</a></b></td>
        <td>object</td>
        <td>
          ConfigMap containing data to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexopsgenieconfigsindexhttpconfigtlsconfigcasecret">secret</a></b></td>
        <td>object</td>
        <td>
          Secret containing data to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].opsgenieConfigs[index].httpConfig.tlsConfig.ca.configMap
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexopsgenieconfigsindexhttpconfigtlsconfigca)</sup></sup>



ConfigMap containing data to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key to select.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the ConfigMap or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].opsgenieConfigs[index].httpConfig.tlsConfig.ca.secret
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexopsgenieconfigsindexhttpconfigtlsconfigca)</sup></sup>



Secret containing data to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].opsgenieConfigs[index].httpConfig.tlsConfig.cert
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexopsgenieconfigsindexhttpconfigtlsconfig)</sup></sup>



Struct containing the client cert file for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexopsgenieconfigsindexhttpconfigtlsconfigcertconfigmap">configMap</a></b></td>
        <td>object</td>
        <td>
          ConfigMap containing data to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexopsgenieconfigsindexhttpconfigtlsconfigcertsecret">secret</a></b></td>
        <td>object</td>
        <td>
          Secret containing data to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].opsgenieConfigs[index].httpConfig.tlsConfig.cert.configMap
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexopsgenieconfigsindexhttpconfigtlsconfigcert)</sup></sup>



ConfigMap containing data to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key to select.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the ConfigMap or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].opsgenieConfigs[index].httpConfig.tlsConfig.cert.secret
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexopsgenieconfigsindexhttpconfigtlsconfigcert)</sup></sup>



Secret containing data to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].opsgenieConfigs[index].httpConfig.tlsConfig.keySecret
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexopsgenieconfigsindexhttpconfigtlsconfig)</sup></sup>



Secret containing the client key file for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].opsgenieConfigs[index].responders[index]
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexopsgenieconfigsindex)</sup></sup>



OpsGenieConfigResponder defines a responder to an incident. One of `id`, `name` or `username` has to be defined.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          Type of responder.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>id</b></td>
        <td>string</td>
        <td>
          ID of the responder.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the responder.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>username</b></td>
        <td>string</td>
        <td>
          Username of the responder.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pagerdutyConfigs[index]
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindex)</sup></sup>



PagerDutyConfig configures notifications via PagerDuty. See https://prometheus.io/docs/alerting/latest/configuration/#pagerduty_config

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>class</b></td>
        <td>string</td>
        <td>
          The class/type of the event.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>client</b></td>
        <td>string</td>
        <td>
          Client identification.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>clientURL</b></td>
        <td>string</td>
        <td>
          Backlink to the sender of notification.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>component</b></td>
        <td>string</td>
        <td>
          The part or component of the affected system that is broken.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>description</b></td>
        <td>string</td>
        <td>
          Description of the incident.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpagerdutyconfigsindexdetailsindex">details</a></b></td>
        <td>[]object</td>
        <td>
          Arbitrary key/value pairs that provide further detail about the incident.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>group</b></td>
        <td>string</td>
        <td>
          A cluster or grouping of sources.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpagerdutyconfigsindexhttpconfig">httpConfig</a></b></td>
        <td>object</td>
        <td>
          HTTP client configuration.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpagerdutyconfigsindexroutingkey">routingKey</a></b></td>
        <td>object</td>
        <td>
          The secret's key that contains the PagerDuty integration key (when using Events API v2). Either this field or `serviceKey` needs to be defined. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sendResolved</b></td>
        <td>boolean</td>
        <td>
          Whether or not to notify about resolved alerts.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpagerdutyconfigsindexservicekey">serviceKey</a></b></td>
        <td>object</td>
        <td>
          The secret's key that contains the PagerDuty service key (when using integration type "Prometheus"). Either this field or `routingKey` needs to be defined. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>severity</b></td>
        <td>string</td>
        <td>
          Severity of the incident.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>url</b></td>
        <td>string</td>
        <td>
          The URL to send requests to.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pagerdutyConfigs[index].details[index]
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpagerdutyconfigsindex)</sup></sup>



KeyValue defines a (key, value) tuple.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          Key of the tuple.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Value of the tuple.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pagerdutyConfigs[index].httpConfig
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpagerdutyconfigsindex)</sup></sup>



HTTP client configuration.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpagerdutyconfigsindexhttpconfigauthorization">authorization</a></b></td>
        <td>object</td>
        <td>
          Authorization header configuration for the client. This is mutually exclusive with BasicAuth and is only available starting from Alertmanager v0.22+.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpagerdutyconfigsindexhttpconfigbasicauth">basicAuth</a></b></td>
        <td>object</td>
        <td>
          BasicAuth for the client. This is mutually exclusive with Authorization. If both are defined, BasicAuth takes precedence.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpagerdutyconfigsindexhttpconfigbearertokensecret">bearerTokenSecret</a></b></td>
        <td>object</td>
        <td>
          The secret's key that contains the bearer token to be used by the client for authentication. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>proxyURL</b></td>
        <td>string</td>
        <td>
          Optional proxy URL.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpagerdutyconfigsindexhttpconfigtlsconfig">tlsConfig</a></b></td>
        <td>object</td>
        <td>
          TLS configuration for the client.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pagerdutyConfigs[index].httpConfig.authorization
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpagerdutyconfigsindexhttpconfig)</sup></sup>



Authorization header configuration for the client. This is mutually exclusive with BasicAuth and is only available starting from Alertmanager v0.22+.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpagerdutyconfigsindexhttpconfigauthorizationcredentials">credentials</a></b></td>
        <td>object</td>
        <td>
          The secret's key that contains the credentials of the request<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          Set the authentication type. Defaults to Bearer, Basic will cause an error<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pagerdutyConfigs[index].httpConfig.authorization.credentials
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpagerdutyconfigsindexhttpconfigauthorization)</sup></sup>



The secret's key that contains the credentials of the request

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pagerdutyConfigs[index].httpConfig.basicAuth
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpagerdutyconfigsindexhttpconfig)</sup></sup>



BasicAuth for the client. This is mutually exclusive with Authorization. If both are defined, BasicAuth takes precedence.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpagerdutyconfigsindexhttpconfigbasicauthpassword">password</a></b></td>
        <td>object</td>
        <td>
          The secret in the service monitor namespace that contains the password for authentication.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpagerdutyconfigsindexhttpconfigbasicauthusername">username</a></b></td>
        <td>object</td>
        <td>
          The secret in the service monitor namespace that contains the username for authentication.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pagerdutyConfigs[index].httpConfig.basicAuth.password
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpagerdutyconfigsindexhttpconfigbasicauth)</sup></sup>



The secret in the service monitor namespace that contains the password for authentication.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pagerdutyConfigs[index].httpConfig.basicAuth.username
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpagerdutyconfigsindexhttpconfigbasicauth)</sup></sup>



The secret in the service monitor namespace that contains the username for authentication.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pagerdutyConfigs[index].httpConfig.bearerTokenSecret
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpagerdutyconfigsindexhttpconfig)</sup></sup>



The secret's key that contains the bearer token to be used by the client for authentication. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pagerdutyConfigs[index].httpConfig.tlsConfig
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpagerdutyconfigsindexhttpconfig)</sup></sup>



TLS configuration for the client.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpagerdutyconfigsindexhttpconfigtlsconfigca">ca</a></b></td>
        <td>object</td>
        <td>
          Struct containing the CA cert to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpagerdutyconfigsindexhttpconfigtlsconfigcert">cert</a></b></td>
        <td>object</td>
        <td>
          Struct containing the client cert file for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>insecureSkipVerify</b></td>
        <td>boolean</td>
        <td>
          Disable target certificate validation.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpagerdutyconfigsindexhttpconfigtlsconfigkeysecret">keySecret</a></b></td>
        <td>object</td>
        <td>
          Secret containing the client key file for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>serverName</b></td>
        <td>string</td>
        <td>
          Used to verify the hostname for the targets.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pagerdutyConfigs[index].httpConfig.tlsConfig.ca
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpagerdutyconfigsindexhttpconfigtlsconfig)</sup></sup>



Struct containing the CA cert to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpagerdutyconfigsindexhttpconfigtlsconfigcaconfigmap">configMap</a></b></td>
        <td>object</td>
        <td>
          ConfigMap containing data to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpagerdutyconfigsindexhttpconfigtlsconfigcasecret">secret</a></b></td>
        <td>object</td>
        <td>
          Secret containing data to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pagerdutyConfigs[index].httpConfig.tlsConfig.ca.configMap
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpagerdutyconfigsindexhttpconfigtlsconfigca)</sup></sup>



ConfigMap containing data to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key to select.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the ConfigMap or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pagerdutyConfigs[index].httpConfig.tlsConfig.ca.secret
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpagerdutyconfigsindexhttpconfigtlsconfigca)</sup></sup>



Secret containing data to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pagerdutyConfigs[index].httpConfig.tlsConfig.cert
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpagerdutyconfigsindexhttpconfigtlsconfig)</sup></sup>



Struct containing the client cert file for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpagerdutyconfigsindexhttpconfigtlsconfigcertconfigmap">configMap</a></b></td>
        <td>object</td>
        <td>
          ConfigMap containing data to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpagerdutyconfigsindexhttpconfigtlsconfigcertsecret">secret</a></b></td>
        <td>object</td>
        <td>
          Secret containing data to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pagerdutyConfigs[index].httpConfig.tlsConfig.cert.configMap
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpagerdutyconfigsindexhttpconfigtlsconfigcert)</sup></sup>



ConfigMap containing data to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key to select.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the ConfigMap or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pagerdutyConfigs[index].httpConfig.tlsConfig.cert.secret
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpagerdutyconfigsindexhttpconfigtlsconfigcert)</sup></sup>



Secret containing data to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pagerdutyConfigs[index].httpConfig.tlsConfig.keySecret
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpagerdutyconfigsindexhttpconfigtlsconfig)</sup></sup>



Secret containing the client key file for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pagerdutyConfigs[index].routingKey
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpagerdutyconfigsindex)</sup></sup>



The secret's key that contains the PagerDuty integration key (when using Events API v2). Either this field or `serviceKey` needs to be defined. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pagerdutyConfigs[index].serviceKey
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpagerdutyconfigsindex)</sup></sup>



The secret's key that contains the PagerDuty service key (when using integration type "Prometheus"). Either this field or `routingKey` needs to be defined. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pushoverConfigs[index]
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindex)</sup></sup>



PushoverConfig configures notifications via Pushover. See https://prometheus.io/docs/alerting/latest/configuration/#pushover_config

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>expire</b></td>
        <td>string</td>
        <td>
          How long your notification will continue to be retried for, unless the user acknowledges the notification.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>html</b></td>
        <td>boolean</td>
        <td>
          Whether notification message is HTML or plain text.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpushoverconfigsindexhttpconfig">httpConfig</a></b></td>
        <td>object</td>
        <td>
          HTTP client configuration.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          Notification message.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>priority</b></td>
        <td>string</td>
        <td>
          Priority, see https://pushover.net/api#priority<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>retry</b></td>
        <td>string</td>
        <td>
          How often the Pushover servers will send the same notification to the user. Must be at least 30 seconds.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sendResolved</b></td>
        <td>boolean</td>
        <td>
          Whether or not to notify about resolved alerts.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sound</b></td>
        <td>string</td>
        <td>
          The name of one of the sounds supported by device clients to override the user's default sound choice<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>title</b></td>
        <td>string</td>
        <td>
          Notification title.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpushoverconfigsindextoken">token</a></b></td>
        <td>object</td>
        <td>
          The secret's key that contains the registered application’s API token, see https://pushover.net/apps. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>url</b></td>
        <td>string</td>
        <td>
          A supplementary URL shown alongside the message.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>urlTitle</b></td>
        <td>string</td>
        <td>
          A title for supplementary URL, otherwise just the URL is shown<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpushoverconfigsindexuserkey">userKey</a></b></td>
        <td>object</td>
        <td>
          The secret's key that contains the recipient user’s user key. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pushoverConfigs[index].httpConfig
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpushoverconfigsindex)</sup></sup>



HTTP client configuration.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpushoverconfigsindexhttpconfigauthorization">authorization</a></b></td>
        <td>object</td>
        <td>
          Authorization header configuration for the client. This is mutually exclusive with BasicAuth and is only available starting from Alertmanager v0.22+.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpushoverconfigsindexhttpconfigbasicauth">basicAuth</a></b></td>
        <td>object</td>
        <td>
          BasicAuth for the client. This is mutually exclusive with Authorization. If both are defined, BasicAuth takes precedence.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpushoverconfigsindexhttpconfigbearertokensecret">bearerTokenSecret</a></b></td>
        <td>object</td>
        <td>
          The secret's key that contains the bearer token to be used by the client for authentication. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>proxyURL</b></td>
        <td>string</td>
        <td>
          Optional proxy URL.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpushoverconfigsindexhttpconfigtlsconfig">tlsConfig</a></b></td>
        <td>object</td>
        <td>
          TLS configuration for the client.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pushoverConfigs[index].httpConfig.authorization
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpushoverconfigsindexhttpconfig)</sup></sup>



Authorization header configuration for the client. This is mutually exclusive with BasicAuth and is only available starting from Alertmanager v0.22+.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpushoverconfigsindexhttpconfigauthorizationcredentials">credentials</a></b></td>
        <td>object</td>
        <td>
          The secret's key that contains the credentials of the request<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          Set the authentication type. Defaults to Bearer, Basic will cause an error<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pushoverConfigs[index].httpConfig.authorization.credentials
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpushoverconfigsindexhttpconfigauthorization)</sup></sup>



The secret's key that contains the credentials of the request

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pushoverConfigs[index].httpConfig.basicAuth
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpushoverconfigsindexhttpconfig)</sup></sup>



BasicAuth for the client. This is mutually exclusive with Authorization. If both are defined, BasicAuth takes precedence.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpushoverconfigsindexhttpconfigbasicauthpassword">password</a></b></td>
        <td>object</td>
        <td>
          The secret in the service monitor namespace that contains the password for authentication.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpushoverconfigsindexhttpconfigbasicauthusername">username</a></b></td>
        <td>object</td>
        <td>
          The secret in the service monitor namespace that contains the username for authentication.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pushoverConfigs[index].httpConfig.basicAuth.password
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpushoverconfigsindexhttpconfigbasicauth)</sup></sup>



The secret in the service monitor namespace that contains the password for authentication.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pushoverConfigs[index].httpConfig.basicAuth.username
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpushoverconfigsindexhttpconfigbasicauth)</sup></sup>



The secret in the service monitor namespace that contains the username for authentication.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pushoverConfigs[index].httpConfig.bearerTokenSecret
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpushoverconfigsindexhttpconfig)</sup></sup>



The secret's key that contains the bearer token to be used by the client for authentication. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pushoverConfigs[index].httpConfig.tlsConfig
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpushoverconfigsindexhttpconfig)</sup></sup>



TLS configuration for the client.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpushoverconfigsindexhttpconfigtlsconfigca">ca</a></b></td>
        <td>object</td>
        <td>
          Struct containing the CA cert to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpushoverconfigsindexhttpconfigtlsconfigcert">cert</a></b></td>
        <td>object</td>
        <td>
          Struct containing the client cert file for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>insecureSkipVerify</b></td>
        <td>boolean</td>
        <td>
          Disable target certificate validation.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpushoverconfigsindexhttpconfigtlsconfigkeysecret">keySecret</a></b></td>
        <td>object</td>
        <td>
          Secret containing the client key file for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>serverName</b></td>
        <td>string</td>
        <td>
          Used to verify the hostname for the targets.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pushoverConfigs[index].httpConfig.tlsConfig.ca
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpushoverconfigsindexhttpconfigtlsconfig)</sup></sup>



Struct containing the CA cert to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpushoverconfigsindexhttpconfigtlsconfigcaconfigmap">configMap</a></b></td>
        <td>object</td>
        <td>
          ConfigMap containing data to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpushoverconfigsindexhttpconfigtlsconfigcasecret">secret</a></b></td>
        <td>object</td>
        <td>
          Secret containing data to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pushoverConfigs[index].httpConfig.tlsConfig.ca.configMap
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpushoverconfigsindexhttpconfigtlsconfigca)</sup></sup>



ConfigMap containing data to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key to select.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the ConfigMap or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pushoverConfigs[index].httpConfig.tlsConfig.ca.secret
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpushoverconfigsindexhttpconfigtlsconfigca)</sup></sup>



Secret containing data to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pushoverConfigs[index].httpConfig.tlsConfig.cert
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpushoverconfigsindexhttpconfigtlsconfig)</sup></sup>



Struct containing the client cert file for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpushoverconfigsindexhttpconfigtlsconfigcertconfigmap">configMap</a></b></td>
        <td>object</td>
        <td>
          ConfigMap containing data to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexpushoverconfigsindexhttpconfigtlsconfigcertsecret">secret</a></b></td>
        <td>object</td>
        <td>
          Secret containing data to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pushoverConfigs[index].httpConfig.tlsConfig.cert.configMap
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpushoverconfigsindexhttpconfigtlsconfigcert)</sup></sup>



ConfigMap containing data to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key to select.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the ConfigMap or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pushoverConfigs[index].httpConfig.tlsConfig.cert.secret
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpushoverconfigsindexhttpconfigtlsconfigcert)</sup></sup>



Secret containing data to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pushoverConfigs[index].httpConfig.tlsConfig.keySecret
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpushoverconfigsindexhttpconfigtlsconfig)</sup></sup>



Secret containing the client key file for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pushoverConfigs[index].token
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpushoverconfigsindex)</sup></sup>



The secret's key that contains the registered application’s API token, see https://pushover.net/apps. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].pushoverConfigs[index].userKey
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexpushoverconfigsindex)</sup></sup>



The secret's key that contains the recipient user’s user key. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].slackConfigs[index]
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindex)</sup></sup>



SlackConfig configures notifications via Slack. See https://prometheus.io/docs/alerting/latest/configuration/#slack_config

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexslackconfigsindexactionsindex">actions</a></b></td>
        <td>[]object</td>
        <td>
          A list of Slack actions that are sent with each notification.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexslackconfigsindexapiurl">apiURL</a></b></td>
        <td>object</td>
        <td>
          The secret's key that contains the Slack webhook URL. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>callbackId</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>channel</b></td>
        <td>string</td>
        <td>
          The channel or user to send notifications to.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>color</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>fallback</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexslackconfigsindexfieldsindex">fields</a></b></td>
        <td>[]object</td>
        <td>
          A list of Slack fields that are sent with each notification.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>footer</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexslackconfigsindexhttpconfig">httpConfig</a></b></td>
        <td>object</td>
        <td>
          HTTP client configuration.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>iconEmoji</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>iconURL</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>imageURL</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>linkNames</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>mrkdwnIn</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>pretext</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sendResolved</b></td>
        <td>boolean</td>
        <td>
          Whether or not to notify about resolved alerts.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>shortFields</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>text</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>thumbURL</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>title</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>titleLink</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>username</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].slackConfigs[index].actions[index]
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexslackconfigsindex)</sup></sup>



SlackAction configures a single Slack action that is sent with each notification. See https://api.slack.com/docs/message-attachments#action_fields and https://api.slack.com/docs/message-buttons for more information.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>text</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexslackconfigsindexactionsindexconfirm">confirm</a></b></td>
        <td>object</td>
        <td>
          SlackConfirmationField protect users from destructive actions or particularly distinguished decisions by asking them to confirm their button click one more time. See https://api.slack.com/docs/interactive-message-field-guide#confirmation_fields for more information.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>style</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>url</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].slackConfigs[index].actions[index].confirm
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexslackconfigsindexactionsindex)</sup></sup>



SlackConfirmationField protect users from destructive actions or particularly distinguished decisions by asking them to confirm their button click one more time. See https://api.slack.com/docs/interactive-message-field-guide#confirmation_fields for more information.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>text</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>dismissText</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>okText</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>title</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].slackConfigs[index].apiURL
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexslackconfigsindex)</sup></sup>



The secret's key that contains the Slack webhook URL. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].slackConfigs[index].fields[index]
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexslackconfigsindex)</sup></sup>



SlackField configures a single Slack field that is sent with each notification. Each field must contain a title, value, and optionally, a boolean value to indicate if the field is short enough to be displayed next to other fields designated as short. See https://api.slack.com/docs/message-attachments#fields for more information.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>title</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>short</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].slackConfigs[index].httpConfig
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexslackconfigsindex)</sup></sup>



HTTP client configuration.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexslackconfigsindexhttpconfigauthorization">authorization</a></b></td>
        <td>object</td>
        <td>
          Authorization header configuration for the client. This is mutually exclusive with BasicAuth and is only available starting from Alertmanager v0.22+.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexslackconfigsindexhttpconfigbasicauth">basicAuth</a></b></td>
        <td>object</td>
        <td>
          BasicAuth for the client. This is mutually exclusive with Authorization. If both are defined, BasicAuth takes precedence.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexslackconfigsindexhttpconfigbearertokensecret">bearerTokenSecret</a></b></td>
        <td>object</td>
        <td>
          The secret's key that contains the bearer token to be used by the client for authentication. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>proxyURL</b></td>
        <td>string</td>
        <td>
          Optional proxy URL.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexslackconfigsindexhttpconfigtlsconfig">tlsConfig</a></b></td>
        <td>object</td>
        <td>
          TLS configuration for the client.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].slackConfigs[index].httpConfig.authorization
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexslackconfigsindexhttpconfig)</sup></sup>



Authorization header configuration for the client. This is mutually exclusive with BasicAuth and is only available starting from Alertmanager v0.22+.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexslackconfigsindexhttpconfigauthorizationcredentials">credentials</a></b></td>
        <td>object</td>
        <td>
          The secret's key that contains the credentials of the request<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          Set the authentication type. Defaults to Bearer, Basic will cause an error<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].slackConfigs[index].httpConfig.authorization.credentials
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexslackconfigsindexhttpconfigauthorization)</sup></sup>



The secret's key that contains the credentials of the request

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].slackConfigs[index].httpConfig.basicAuth
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexslackconfigsindexhttpconfig)</sup></sup>



BasicAuth for the client. This is mutually exclusive with Authorization. If both are defined, BasicAuth takes precedence.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexslackconfigsindexhttpconfigbasicauthpassword">password</a></b></td>
        <td>object</td>
        <td>
          The secret in the service monitor namespace that contains the password for authentication.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexslackconfigsindexhttpconfigbasicauthusername">username</a></b></td>
        <td>object</td>
        <td>
          The secret in the service monitor namespace that contains the username for authentication.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].slackConfigs[index].httpConfig.basicAuth.password
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexslackconfigsindexhttpconfigbasicauth)</sup></sup>



The secret in the service monitor namespace that contains the password for authentication.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].slackConfigs[index].httpConfig.basicAuth.username
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexslackconfigsindexhttpconfigbasicauth)</sup></sup>



The secret in the service monitor namespace that contains the username for authentication.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].slackConfigs[index].httpConfig.bearerTokenSecret
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexslackconfigsindexhttpconfig)</sup></sup>



The secret's key that contains the bearer token to be used by the client for authentication. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].slackConfigs[index].httpConfig.tlsConfig
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexslackconfigsindexhttpconfig)</sup></sup>



TLS configuration for the client.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexslackconfigsindexhttpconfigtlsconfigca">ca</a></b></td>
        <td>object</td>
        <td>
          Struct containing the CA cert to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexslackconfigsindexhttpconfigtlsconfigcert">cert</a></b></td>
        <td>object</td>
        <td>
          Struct containing the client cert file for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>insecureSkipVerify</b></td>
        <td>boolean</td>
        <td>
          Disable target certificate validation.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexslackconfigsindexhttpconfigtlsconfigkeysecret">keySecret</a></b></td>
        <td>object</td>
        <td>
          Secret containing the client key file for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>serverName</b></td>
        <td>string</td>
        <td>
          Used to verify the hostname for the targets.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].slackConfigs[index].httpConfig.tlsConfig.ca
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexslackconfigsindexhttpconfigtlsconfig)</sup></sup>



Struct containing the CA cert to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexslackconfigsindexhttpconfigtlsconfigcaconfigmap">configMap</a></b></td>
        <td>object</td>
        <td>
          ConfigMap containing data to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexslackconfigsindexhttpconfigtlsconfigcasecret">secret</a></b></td>
        <td>object</td>
        <td>
          Secret containing data to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].slackConfigs[index].httpConfig.tlsConfig.ca.configMap
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexslackconfigsindexhttpconfigtlsconfigca)</sup></sup>



ConfigMap containing data to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key to select.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the ConfigMap or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].slackConfigs[index].httpConfig.tlsConfig.ca.secret
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexslackconfigsindexhttpconfigtlsconfigca)</sup></sup>



Secret containing data to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].slackConfigs[index].httpConfig.tlsConfig.cert
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexslackconfigsindexhttpconfigtlsconfig)</sup></sup>



Struct containing the client cert file for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexslackconfigsindexhttpconfigtlsconfigcertconfigmap">configMap</a></b></td>
        <td>object</td>
        <td>
          ConfigMap containing data to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexslackconfigsindexhttpconfigtlsconfigcertsecret">secret</a></b></td>
        <td>object</td>
        <td>
          Secret containing data to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].slackConfigs[index].httpConfig.tlsConfig.cert.configMap
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexslackconfigsindexhttpconfigtlsconfigcert)</sup></sup>



ConfigMap containing data to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key to select.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the ConfigMap or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].slackConfigs[index].httpConfig.tlsConfig.cert.secret
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexslackconfigsindexhttpconfigtlsconfigcert)</sup></sup>



Secret containing data to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].slackConfigs[index].httpConfig.tlsConfig.keySecret
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexslackconfigsindexhttpconfigtlsconfig)</sup></sup>



Secret containing the client key file for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].victoropsConfigs[index]
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindex)</sup></sup>



VictorOpsConfig configures notifications via VictorOps. See https://prometheus.io/docs/alerting/latest/configuration/#victorops_config

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexvictoropsconfigsindexapikey">apiKey</a></b></td>
        <td>object</td>
        <td>
          The secret's key that contains the API key to use when talking to the VictorOps API. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>apiUrl</b></td>
        <td>string</td>
        <td>
          The VictorOps API URL.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexvictoropsconfigsindexcustomfieldsindex">customFields</a></b></td>
        <td>[]object</td>
        <td>
          Additional custom fields for notification.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>entityDisplayName</b></td>
        <td>string</td>
        <td>
          Contains summary of the alerted problem.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexvictoropsconfigsindexhttpconfig">httpConfig</a></b></td>
        <td>object</td>
        <td>
          The HTTP client's configuration.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>messageType</b></td>
        <td>string</td>
        <td>
          Describes the behavior of the alert (CRITICAL, WARNING, INFO).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>monitoringTool</b></td>
        <td>string</td>
        <td>
          The monitoring tool the state message is from.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>routingKey</b></td>
        <td>string</td>
        <td>
          A key used to map the alert to a team.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sendResolved</b></td>
        <td>boolean</td>
        <td>
          Whether or not to notify about resolved alerts.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>stateMessage</b></td>
        <td>string</td>
        <td>
          Contains long explanation of the alerted problem.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].victoropsConfigs[index].apiKey
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexvictoropsconfigsindex)</sup></sup>



The secret's key that contains the API key to use when talking to the VictorOps API. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].victoropsConfigs[index].customFields[index]
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexvictoropsconfigsindex)</sup></sup>



KeyValue defines a (key, value) tuple.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          Key of the tuple.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Value of the tuple.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].victoropsConfigs[index].httpConfig
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexvictoropsconfigsindex)</sup></sup>



The HTTP client's configuration.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexvictoropsconfigsindexhttpconfigauthorization">authorization</a></b></td>
        <td>object</td>
        <td>
          Authorization header configuration for the client. This is mutually exclusive with BasicAuth and is only available starting from Alertmanager v0.22+.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexvictoropsconfigsindexhttpconfigbasicauth">basicAuth</a></b></td>
        <td>object</td>
        <td>
          BasicAuth for the client. This is mutually exclusive with Authorization. If both are defined, BasicAuth takes precedence.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexvictoropsconfigsindexhttpconfigbearertokensecret">bearerTokenSecret</a></b></td>
        <td>object</td>
        <td>
          The secret's key that contains the bearer token to be used by the client for authentication. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>proxyURL</b></td>
        <td>string</td>
        <td>
          Optional proxy URL.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexvictoropsconfigsindexhttpconfigtlsconfig">tlsConfig</a></b></td>
        <td>object</td>
        <td>
          TLS configuration for the client.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].victoropsConfigs[index].httpConfig.authorization
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexvictoropsconfigsindexhttpconfig)</sup></sup>



Authorization header configuration for the client. This is mutually exclusive with BasicAuth and is only available starting from Alertmanager v0.22+.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexvictoropsconfigsindexhttpconfigauthorizationcredentials">credentials</a></b></td>
        <td>object</td>
        <td>
          The secret's key that contains the credentials of the request<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          Set the authentication type. Defaults to Bearer, Basic will cause an error<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].victoropsConfigs[index].httpConfig.authorization.credentials
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexvictoropsconfigsindexhttpconfigauthorization)</sup></sup>



The secret's key that contains the credentials of the request

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].victoropsConfigs[index].httpConfig.basicAuth
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexvictoropsconfigsindexhttpconfig)</sup></sup>



BasicAuth for the client. This is mutually exclusive with Authorization. If both are defined, BasicAuth takes precedence.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexvictoropsconfigsindexhttpconfigbasicauthpassword">password</a></b></td>
        <td>object</td>
        <td>
          The secret in the service monitor namespace that contains the password for authentication.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexvictoropsconfigsindexhttpconfigbasicauthusername">username</a></b></td>
        <td>object</td>
        <td>
          The secret in the service monitor namespace that contains the username for authentication.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].victoropsConfigs[index].httpConfig.basicAuth.password
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexvictoropsconfigsindexhttpconfigbasicauth)</sup></sup>



The secret in the service monitor namespace that contains the password for authentication.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].victoropsConfigs[index].httpConfig.basicAuth.username
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexvictoropsconfigsindexhttpconfigbasicauth)</sup></sup>



The secret in the service monitor namespace that contains the username for authentication.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].victoropsConfigs[index].httpConfig.bearerTokenSecret
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexvictoropsconfigsindexhttpconfig)</sup></sup>



The secret's key that contains the bearer token to be used by the client for authentication. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].victoropsConfigs[index].httpConfig.tlsConfig
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexvictoropsconfigsindexhttpconfig)</sup></sup>



TLS configuration for the client.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexvictoropsconfigsindexhttpconfigtlsconfigca">ca</a></b></td>
        <td>object</td>
        <td>
          Struct containing the CA cert to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexvictoropsconfigsindexhttpconfigtlsconfigcert">cert</a></b></td>
        <td>object</td>
        <td>
          Struct containing the client cert file for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>insecureSkipVerify</b></td>
        <td>boolean</td>
        <td>
          Disable target certificate validation.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexvictoropsconfigsindexhttpconfigtlsconfigkeysecret">keySecret</a></b></td>
        <td>object</td>
        <td>
          Secret containing the client key file for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>serverName</b></td>
        <td>string</td>
        <td>
          Used to verify the hostname for the targets.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].victoropsConfigs[index].httpConfig.tlsConfig.ca
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexvictoropsconfigsindexhttpconfigtlsconfig)</sup></sup>



Struct containing the CA cert to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexvictoropsconfigsindexhttpconfigtlsconfigcaconfigmap">configMap</a></b></td>
        <td>object</td>
        <td>
          ConfigMap containing data to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexvictoropsconfigsindexhttpconfigtlsconfigcasecret">secret</a></b></td>
        <td>object</td>
        <td>
          Secret containing data to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].victoropsConfigs[index].httpConfig.tlsConfig.ca.configMap
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexvictoropsconfigsindexhttpconfigtlsconfigca)</sup></sup>



ConfigMap containing data to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key to select.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the ConfigMap or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].victoropsConfigs[index].httpConfig.tlsConfig.ca.secret
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexvictoropsconfigsindexhttpconfigtlsconfigca)</sup></sup>



Secret containing data to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].victoropsConfigs[index].httpConfig.tlsConfig.cert
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexvictoropsconfigsindexhttpconfigtlsconfig)</sup></sup>



Struct containing the client cert file for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexvictoropsconfigsindexhttpconfigtlsconfigcertconfigmap">configMap</a></b></td>
        <td>object</td>
        <td>
          ConfigMap containing data to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexvictoropsconfigsindexhttpconfigtlsconfigcertsecret">secret</a></b></td>
        <td>object</td>
        <td>
          Secret containing data to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].victoropsConfigs[index].httpConfig.tlsConfig.cert.configMap
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexvictoropsconfigsindexhttpconfigtlsconfigcert)</sup></sup>



ConfigMap containing data to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key to select.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the ConfigMap or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].victoropsConfigs[index].httpConfig.tlsConfig.cert.secret
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexvictoropsconfigsindexhttpconfigtlsconfigcert)</sup></sup>



Secret containing data to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].victoropsConfigs[index].httpConfig.tlsConfig.keySecret
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexvictoropsconfigsindexhttpconfigtlsconfig)</sup></sup>



Secret containing the client key file for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].webhookConfigs[index]
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindex)</sup></sup>



WebhookConfig configures notifications via a generic receiver supporting the webhook payload. See https://prometheus.io/docs/alerting/latest/configuration/#webhook_config

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexwebhookconfigsindexhttpconfig">httpConfig</a></b></td>
        <td>object</td>
        <td>
          HTTP client configuration.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>maxAlerts</b></td>
        <td>integer</td>
        <td>
          Maximum number of alerts to be sent per webhook message. When 0, all alerts are included.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sendResolved</b></td>
        <td>boolean</td>
        <td>
          Whether or not to notify about resolved alerts.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>url</b></td>
        <td>string</td>
        <td>
          The URL to send HTTP POST requests to. `urlSecret` takes precedence over `url`. One of `urlSecret` and `url` should be defined.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexwebhookconfigsindexurlsecret">urlSecret</a></b></td>
        <td>object</td>
        <td>
          The secret's key that contains the webhook URL to send HTTP requests to. `urlSecret` takes precedence over `url`. One of `urlSecret` and `url` should be defined. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].webhookConfigs[index].httpConfig
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexwebhookconfigsindex)</sup></sup>



HTTP client configuration.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexwebhookconfigsindexhttpconfigauthorization">authorization</a></b></td>
        <td>object</td>
        <td>
          Authorization header configuration for the client. This is mutually exclusive with BasicAuth and is only available starting from Alertmanager v0.22+.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexwebhookconfigsindexhttpconfigbasicauth">basicAuth</a></b></td>
        <td>object</td>
        <td>
          BasicAuth for the client. This is mutually exclusive with Authorization. If both are defined, BasicAuth takes precedence.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexwebhookconfigsindexhttpconfigbearertokensecret">bearerTokenSecret</a></b></td>
        <td>object</td>
        <td>
          The secret's key that contains the bearer token to be used by the client for authentication. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>proxyURL</b></td>
        <td>string</td>
        <td>
          Optional proxy URL.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexwebhookconfigsindexhttpconfigtlsconfig">tlsConfig</a></b></td>
        <td>object</td>
        <td>
          TLS configuration for the client.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].webhookConfigs[index].httpConfig.authorization
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexwebhookconfigsindexhttpconfig)</sup></sup>



Authorization header configuration for the client. This is mutually exclusive with BasicAuth and is only available starting from Alertmanager v0.22+.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexwebhookconfigsindexhttpconfigauthorizationcredentials">credentials</a></b></td>
        <td>object</td>
        <td>
          The secret's key that contains the credentials of the request<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          Set the authentication type. Defaults to Bearer, Basic will cause an error<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].webhookConfigs[index].httpConfig.authorization.credentials
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexwebhookconfigsindexhttpconfigauthorization)</sup></sup>



The secret's key that contains the credentials of the request

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].webhookConfigs[index].httpConfig.basicAuth
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexwebhookconfigsindexhttpconfig)</sup></sup>



BasicAuth for the client. This is mutually exclusive with Authorization. If both are defined, BasicAuth takes precedence.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexwebhookconfigsindexhttpconfigbasicauthpassword">password</a></b></td>
        <td>object</td>
        <td>
          The secret in the service monitor namespace that contains the password for authentication.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexwebhookconfigsindexhttpconfigbasicauthusername">username</a></b></td>
        <td>object</td>
        <td>
          The secret in the service monitor namespace that contains the username for authentication.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].webhookConfigs[index].httpConfig.basicAuth.password
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexwebhookconfigsindexhttpconfigbasicauth)</sup></sup>



The secret in the service monitor namespace that contains the password for authentication.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].webhookConfigs[index].httpConfig.basicAuth.username
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexwebhookconfigsindexhttpconfigbasicauth)</sup></sup>



The secret in the service monitor namespace that contains the username for authentication.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].webhookConfigs[index].httpConfig.bearerTokenSecret
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexwebhookconfigsindexhttpconfig)</sup></sup>



The secret's key that contains the bearer token to be used by the client for authentication. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].webhookConfigs[index].httpConfig.tlsConfig
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexwebhookconfigsindexhttpconfig)</sup></sup>



TLS configuration for the client.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexwebhookconfigsindexhttpconfigtlsconfigca">ca</a></b></td>
        <td>object</td>
        <td>
          Struct containing the CA cert to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexwebhookconfigsindexhttpconfigtlsconfigcert">cert</a></b></td>
        <td>object</td>
        <td>
          Struct containing the client cert file for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>insecureSkipVerify</b></td>
        <td>boolean</td>
        <td>
          Disable target certificate validation.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexwebhookconfigsindexhttpconfigtlsconfigkeysecret">keySecret</a></b></td>
        <td>object</td>
        <td>
          Secret containing the client key file for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>serverName</b></td>
        <td>string</td>
        <td>
          Used to verify the hostname for the targets.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].webhookConfigs[index].httpConfig.tlsConfig.ca
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexwebhookconfigsindexhttpconfigtlsconfig)</sup></sup>



Struct containing the CA cert to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexwebhookconfigsindexhttpconfigtlsconfigcaconfigmap">configMap</a></b></td>
        <td>object</td>
        <td>
          ConfigMap containing data to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexwebhookconfigsindexhttpconfigtlsconfigcasecret">secret</a></b></td>
        <td>object</td>
        <td>
          Secret containing data to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].webhookConfigs[index].httpConfig.tlsConfig.ca.configMap
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexwebhookconfigsindexhttpconfigtlsconfigca)</sup></sup>



ConfigMap containing data to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key to select.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the ConfigMap or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].webhookConfigs[index].httpConfig.tlsConfig.ca.secret
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexwebhookconfigsindexhttpconfigtlsconfigca)</sup></sup>



Secret containing data to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].webhookConfigs[index].httpConfig.tlsConfig.cert
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexwebhookconfigsindexhttpconfigtlsconfig)</sup></sup>



Struct containing the client cert file for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexwebhookconfigsindexhttpconfigtlsconfigcertconfigmap">configMap</a></b></td>
        <td>object</td>
        <td>
          ConfigMap containing data to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexwebhookconfigsindexhttpconfigtlsconfigcertsecret">secret</a></b></td>
        <td>object</td>
        <td>
          Secret containing data to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].webhookConfigs[index].httpConfig.tlsConfig.cert.configMap
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexwebhookconfigsindexhttpconfigtlsconfigcert)</sup></sup>



ConfigMap containing data to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key to select.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the ConfigMap or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].webhookConfigs[index].httpConfig.tlsConfig.cert.secret
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexwebhookconfigsindexhttpconfigtlsconfigcert)</sup></sup>



Secret containing data to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].webhookConfigs[index].httpConfig.tlsConfig.keySecret
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexwebhookconfigsindexhttpconfigtlsconfig)</sup></sup>



Secret containing the client key file for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].webhookConfigs[index].urlSecret
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexwebhookconfigsindex)</sup></sup>



The secret's key that contains the webhook URL to send HTTP requests to. `urlSecret` takes precedence over `url`. One of `urlSecret` and `url` should be defined. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].wechatConfigs[index]
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindex)</sup></sup>



WeChatConfig configures notifications via WeChat. See https://prometheus.io/docs/alerting/latest/configuration/#wechat_config

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>agentID</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexwechatconfigsindexapisecret">apiSecret</a></b></td>
        <td>object</td>
        <td>
          The secret's key that contains the WeChat API key. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>apiURL</b></td>
        <td>string</td>
        <td>
          The WeChat API URL.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>corpID</b></td>
        <td>string</td>
        <td>
          The corp id for authentication.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexwechatconfigsindexhttpconfig">httpConfig</a></b></td>
        <td>object</td>
        <td>
          HTTP client configuration.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          API request data as defined by the WeChat API.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>messageType</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sendResolved</b></td>
        <td>boolean</td>
        <td>
          Whether or not to notify about resolved alerts.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>toParty</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>toTag</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>toUser</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].wechatConfigs[index].apiSecret
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexwechatconfigsindex)</sup></sup>



The secret's key that contains the WeChat API key. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].wechatConfigs[index].httpConfig
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexwechatconfigsindex)</sup></sup>



HTTP client configuration.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexwechatconfigsindexhttpconfigauthorization">authorization</a></b></td>
        <td>object</td>
        <td>
          Authorization header configuration for the client. This is mutually exclusive with BasicAuth and is only available starting from Alertmanager v0.22+.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexwechatconfigsindexhttpconfigbasicauth">basicAuth</a></b></td>
        <td>object</td>
        <td>
          BasicAuth for the client. This is mutually exclusive with Authorization. If both are defined, BasicAuth takes precedence.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexwechatconfigsindexhttpconfigbearertokensecret">bearerTokenSecret</a></b></td>
        <td>object</td>
        <td>
          The secret's key that contains the bearer token to be used by the client for authentication. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>proxyURL</b></td>
        <td>string</td>
        <td>
          Optional proxy URL.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexwechatconfigsindexhttpconfigtlsconfig">tlsConfig</a></b></td>
        <td>object</td>
        <td>
          TLS configuration for the client.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].wechatConfigs[index].httpConfig.authorization
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexwechatconfigsindexhttpconfig)</sup></sup>



Authorization header configuration for the client. This is mutually exclusive with BasicAuth and is only available starting from Alertmanager v0.22+.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexwechatconfigsindexhttpconfigauthorizationcredentials">credentials</a></b></td>
        <td>object</td>
        <td>
          The secret's key that contains the credentials of the request<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          Set the authentication type. Defaults to Bearer, Basic will cause an error<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].wechatConfigs[index].httpConfig.authorization.credentials
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexwechatconfigsindexhttpconfigauthorization)</sup></sup>



The secret's key that contains the credentials of the request

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].wechatConfigs[index].httpConfig.basicAuth
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexwechatconfigsindexhttpconfig)</sup></sup>



BasicAuth for the client. This is mutually exclusive with Authorization. If both are defined, BasicAuth takes precedence.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexwechatconfigsindexhttpconfigbasicauthpassword">password</a></b></td>
        <td>object</td>
        <td>
          The secret in the service monitor namespace that contains the password for authentication.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexwechatconfigsindexhttpconfigbasicauthusername">username</a></b></td>
        <td>object</td>
        <td>
          The secret in the service monitor namespace that contains the username for authentication.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].wechatConfigs[index].httpConfig.basicAuth.password
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexwechatconfigsindexhttpconfigbasicauth)</sup></sup>



The secret in the service monitor namespace that contains the password for authentication.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].wechatConfigs[index].httpConfig.basicAuth.username
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexwechatconfigsindexhttpconfigbasicauth)</sup></sup>



The secret in the service monitor namespace that contains the username for authentication.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].wechatConfigs[index].httpConfig.bearerTokenSecret
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexwechatconfigsindexhttpconfig)</sup></sup>



The secret's key that contains the bearer token to be used by the client for authentication. The secret needs to be in the same namespace as the AlertmanagerConfig object and accessible by the Prometheus Operator.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].wechatConfigs[index].httpConfig.tlsConfig
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexwechatconfigsindexhttpconfig)</sup></sup>



TLS configuration for the client.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexwechatconfigsindexhttpconfigtlsconfigca">ca</a></b></td>
        <td>object</td>
        <td>
          Struct containing the CA cert to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexwechatconfigsindexhttpconfigtlsconfigcert">cert</a></b></td>
        <td>object</td>
        <td>
          Struct containing the client cert file for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>insecureSkipVerify</b></td>
        <td>boolean</td>
        <td>
          Disable target certificate validation.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexwechatconfigsindexhttpconfigtlsconfigkeysecret">keySecret</a></b></td>
        <td>object</td>
        <td>
          Secret containing the client key file for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>serverName</b></td>
        <td>string</td>
        <td>
          Used to verify the hostname for the targets.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].wechatConfigs[index].httpConfig.tlsConfig.ca
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexwechatconfigsindexhttpconfigtlsconfig)</sup></sup>



Struct containing the CA cert to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexwechatconfigsindexhttpconfigtlsconfigcaconfigmap">configMap</a></b></td>
        <td>object</td>
        <td>
          ConfigMap containing data to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexwechatconfigsindexhttpconfigtlsconfigcasecret">secret</a></b></td>
        <td>object</td>
        <td>
          Secret containing data to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].wechatConfigs[index].httpConfig.tlsConfig.ca.configMap
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexwechatconfigsindexhttpconfigtlsconfigca)</sup></sup>



ConfigMap containing data to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key to select.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the ConfigMap or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].wechatConfigs[index].httpConfig.tlsConfig.ca.secret
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexwechatconfigsindexhttpconfigtlsconfigca)</sup></sup>



Secret containing data to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].wechatConfigs[index].httpConfig.tlsConfig.cert
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexwechatconfigsindexhttpconfigtlsconfig)</sup></sup>



Struct containing the client cert file for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexwechatconfigsindexhttpconfigtlsconfigcertconfigmap">configMap</a></b></td>
        <td>object</td>
        <td>
          ConfigMap containing data to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecreceiversindexwechatconfigsindexhttpconfigtlsconfigcertsecret">secret</a></b></td>
        <td>object</td>
        <td>
          Secret containing data to use for the targets.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].wechatConfigs[index].httpConfig.tlsConfig.cert.configMap
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexwechatconfigsindexhttpconfigtlsconfigcert)</sup></sup>



ConfigMap containing data to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key to select.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the ConfigMap or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].wechatConfigs[index].httpConfig.tlsConfig.cert.secret
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexwechatconfigsindexhttpconfigtlsconfigcert)</sup></sup>



Secret containing data to use for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.receivers[index].wechatConfigs[index].httpConfig.tlsConfig.keySecret
<sup><sup>[↩ Parent](#alertmanagerconfigspecreceiversindexwechatconfigsindexhttpconfigtlsconfig)</sup></sup>



Secret containing the client key file for the targets.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.route
<sup><sup>[↩ Parent](#alertmanagerconfigspec)</sup></sup>



The Alertmanager route definition for alerts matching the resource’s namespace. If present, it will be added to the generated Alertmanager configuration as a first-level route.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>continue</b></td>
        <td>boolean</td>
        <td>
          Boolean indicating whether an alert should continue matching subsequent sibling nodes. It will always be overridden to true for the first-level route by the Prometheus operator.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>groupBy</b></td>
        <td>[]string</td>
        <td>
          List of labels to group by.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>groupInterval</b></td>
        <td>string</td>
        <td>
          How long to wait before sending an updated notification. Must match the regular expression `[0-9]+(ms|s|m|h)` (milliseconds seconds minutes hours).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>groupWait</b></td>
        <td>string</td>
        <td>
          How long to wait before sending the initial notification. Must match the regular expression `[0-9]+(ms|s|m|h)` (milliseconds seconds minutes hours).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#alertmanagerconfigspecroutematchersindex">matchers</a></b></td>
        <td>[]object</td>
        <td>
          List of matchers that the alert’s labels should match. For the first level route, the operator removes any existing equality and regexp matcher on the `namespace` label and adds a `namespace: <object namespace>` matcher.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>receiver</b></td>
        <td>string</td>
        <td>
          Name of the receiver for this route. If not empty, it should be listed in the `receivers` field.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>repeatInterval</b></td>
        <td>string</td>
        <td>
          How long to wait before repeating the last notification. Must match the regular expression `[0-9]+(ms|s|m|h)` (milliseconds seconds minutes hours).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>routes</b></td>
        <td>[]JSON</td>
        <td>
          Child routes.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### AlertmanagerConfig.spec.route.matchers[index]
<sup><sup>[↩ Parent](#alertmanagerconfigspecroute)</sup></sup>



Matcher defines how to match on alert's labels.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Label to match.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>regex</b></td>
        <td>boolean</td>
        <td>
          Whether to match on equality (false) or regular-expression (true).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Label value to match.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>