<p align="center"><a href="#readme"><img src=".github/images/card.svg"/></a></p>

<p align="center">
  <a href="https://kaos.sh/w/jira-reindex-runner/ci"><img src="https://kaos.sh/w/jira-reindex-runner/ci.svg" alt="GitHub Actions CI Status" /></a>
  <a href="https://kaos.sh/w/jira-reindex-runner/codeql"><img src="https://kaos.sh/w/jira-reindex-runner/codeql.svg" alt="GitHub Actions CodeQL Status" /></a>
  <a href="#license"><img src=".github/images/license.svg"/></a>
</p>

<p align="center"><a href="#installation">Installation</a> • <a href="#usage">Usage</a> • <a href="#ci-status">CI Status</a> • <a href="#contributing">Contributing</a> • <a href="#license">License</a></p>

<br/>

`jira-reindex-runner` is an app for periodical running Jira re-index process.

Due to [the lack](https://jira.atlassian.com/browse/JRASERVER-70793) of Jira functionality, it is impossible to check if a re-index is required or not. For using this app, you must have [ScriptRunner add-on](https://marketplace.atlassian.com/apps/6820/scriptrunner-for-jira) installed on your Jira instance. Then you have to create a new REST endpoint in ScriptRunner with the following code:

```groovy
import com.onresolve.scriptrunner.runner.rest.common.CustomEndpointDelegate
import groovy.json.JsonBuilder
import groovy.transform.BaseScript
import com.atlassian.jira.component.ComponentAccessor
import com.atlassian.jira.config.DefaultReindexMessageManager

import javax.ws.rs.core.MultivaluedMap
import javax.ws.rs.core.Response

@BaseScript CustomEndpointDelegate delegate

reindexRequired(httpMethod: "GET", groups: ["jira-administrators"]) { MultivaluedMap queryParams, String body ->
  def rmm = ComponentAccessor.getComponent(DefaultReindexMessageManager.class)
  def msg = rmm.getMessageObject()

  if (msg == null) {
    return Response.ok(new JsonBuilder([required: false]).toString()).build();
  }

  return Response.ok(new JsonBuilder([required: true, user: msg.getUserName(), date: msg.getTime()]).toString()).build();
}
```

Using this endpoint, our app can check if re-index is required and run it.

### Installation

#### From source

Make sure you have a working [Go 1.22+](https://github.com/essentialkaos/.github/blob/master/GO-VERSION-SUPPORT.md) workspace ([instructions](https://go.dev/doc/install)), then:

```
go install github.com/essentialkaos/jira-reindex-runner@latest
```

#### From [ESSENTIAL KAOS Public Repository](https://kaos.sh/kaos-repo)

```bash
sudo dnf install -y https://pkgs.kaos.st/kaos-repo-latest.el$(grep 'CPE_NAME' /etc/os-release | tr -d '"' | cut -d':' -f5).noarch.rpm
sudo dnf install jira-reindex-runner
```

### CI Status

| Branch | Status |
|--------|----------|
| `master` | [![CI](https://kaos.sh/w/jira-reindex-runner/ci.svg?branch=master)](https://kaos.sh/w/jira-reindex-runner/ci?query=branch:master) |
| `develop` | [![CI](https://kaos.sh/w/jira-reindex-runner/ci.svg?branch=develop)](https://kaos.sh/w/jira-reindex-runner/ci?query=branch:develop) |

### Contributing

Before contributing to this project please read our [Contributing Guidelines](https://github.com/essentialkaos/contributing-guidelines#contributing-guidelines).

### License

[Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0)

<p align="center"><a href="https://essentialkaos.com"><img src="https://gh.kaos.st/ekgh.svg"/></a></p>
