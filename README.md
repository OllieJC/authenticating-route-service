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

## Environment Variables

### GOOGLE_OAUTH_CLIENT_ID

```sh
cf set-env authenticating-route-service GOOGLE_OAUTH_CLIENT_ID xxx
cf restage authenticating-route-service
```

### GOOGLE_OAUTH_CLIENT_SECRET

```sh
cf set-env authenticating-route-service GOOGLE_OAUTH_CLIENT_SECRET xxx
cf restage authenticating-route-service
```

### SKIP_SSL_VALIDATION

Set this environment variable to true in order to skip the validation of SSL certificates.
By default the route service will attempt to validate certificates.

Example:

```sh
cf set-env authenticating-route-service SKIP_SSL_VALIDATION true
cf restart authenticating-route-service
```
