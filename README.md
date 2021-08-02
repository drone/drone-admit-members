# drone-admit-members

An admission extension to limit system access based on GitHub organization and team membership. Here is a summary of how the extension works:

* if user is organization member, grant access
* if user is organization admin, grant admin access
* if user is member of designated team, grant admin access (optional)
* if user is member of designated team, grant access to drone (optional)

## Installation

Create a shared secret:

```console
$ openssl rand -hex 16
bea26a2221fd8090ea38720fc445eca6
```

Download and run the plugin:

```console
$ docker run -d \
  --publish=3000:3000 \
  --env=DRONE_DEBUG=true \
  --env=DRONE_SECRET=bea26a2221fd8090ea38720fc445eca6 \
  --env=DRONE_GITHUB_TOKEN=3da541559918a808c2402bba5012f6c6 \
  --env=DRONE_GITHUB_ORG=acme \
  --env=DRONE_GITHUB_TEAM=admins \
  --env=DRONE_GITHUB_TEAM_ACESS=drone-team \
  --restart=always \
  --name=admitter drone/drone-admit-members
```

Update your Drone server configuration to include the plugin address and the shared secret.

```text
DRONE_ADMISSION_PLUGIN_ENDPOINT=http://1.2.3.4:3000
DRONE_ADMISSION_PLUGIN_SECRET=bea26a2221fd8090ea38720fc445eca6
```

## Testing

Test the admission extension using the command line tools. First you need to provide the command line tools with the extension endpoint and secret:

```console
export DRONE_ADMISSION_ENDPOINT=http://localhost:3000
export DRONE_ADMISSION_SECRET=bea26a2221fd8090ea38720fc445eca6
```

Use the following command to test account access:

```console
$ drone plugins admit octocat
admission: access denied
```
