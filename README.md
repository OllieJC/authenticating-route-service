# Authenticating Route Service

An authenticating route service for Cloud Foundry.

## Route Service Overview

The Route Service feature is currently in development, the proposal can be found in this [Google Doc](https://docs.google.com/document/d/1bGOQxiKkmaw6uaRWGd-sXpxL0Y28d3QihcluI15FiIA/edit#heading=h.8djffzes9pnb).

This example route service uses the new headers/features that have been added to the GoRouter. For example:

- `X-CF-Forwarded-Url`: A header that contains the original URL that the GoRouter received.
- `X-CF-Proxy-Signature`: A header that the GoRouter uses to determine if a request has gone through the route service.

## Getting Started

- Download this repository and `cf push` to your chosen CF deployment.
- Push your app which will be associated with the route service.
- Create a user-provided route service ([see docs](http://docs.cloudfoundry.org/services/route-services.html#user-provided))
- Bind the route service to the route (domain/hostname)
- Go to "/" and you should get redirected to "/auth/login"

## Configuration

See [configurator](config/README.md).

## Adding route service to an app

```
# ensure the space has a user-provided-service reference
cf create-user-provided-service authenticating-route-service -r https://xxx
# where xxx is the route service endpoint

# bind the app in the space to the user-provided-service in that space
cf bind-route-service DOMAIN --hostname APP_HOSTNAME authenticating-route-service
```
