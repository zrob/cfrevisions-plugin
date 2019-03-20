# cfrevisions-plugin
A CF cli plugin to view revisions and rollback

## Installation
1. git clone the repo to your desktop
1. In the repo, run `go build` to compile a binary
1. run `cf install-plugin <path-to-binary>`

## Known Issue
The CF cli has a [bug](https://github.com/cloudfoundry/cli/issues/1108) that causes the user token to periodically expire. This will manifest as not found errors for
resources that exist. To resolve run a normal cli command and then rerun the command from this plugin.


## Usage

### Enable revisions for an app (since it is an experimental feature)
```
cf enable-revisions my-app
```

### View all revisions
```
cf revisions my-app
```

### View more details for a revision
```
cf revisions my-app 3
```

### Rollback to a revision
```
cf rollback my-app 1
```
