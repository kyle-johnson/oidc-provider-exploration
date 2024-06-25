# Demo: _OpenID Connect_ (OIDC) Provider

This is a bare-bones OIDC provider demonstrating the minimum required for use with [AWS OIDC federation](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_oidc.html).

## Basic Use

Create a JSON Web Key Set ("jwks") and private key:

```shell
go run create-jwks/main.go jwks.json cert.pem
```

The server has to be accessible _and_ be served over HTTPS. I'm using a [Cloudflare Quick Tunnel](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/do-more-with-tunnels/trycloudflare/):

```shell
cloudflared tunnel --url http://localhost:8688 --loglevel debug
```

Start the server (`PREFIX` is optional):

```shell
PREFIX=/oidc2/ BASE_URL=https://<quick tunnel domain>/oidc2/ PORT=8688 AUD=sts.amazonaws.com go run . jwks.json cert.pem
```

Get a token:

```shell
curl http://localhost:8688/issue-token
```

Use it with AWS (see the section below on AWS setup):

```shell
aws sts assume-role-with-web-identity --role-arn <role> --role-session-name <something friendly for logs> --web-identity-token <token>
```

### AWS Setup

Setup an "Identity Provider" in IAM as type "OpenID Connect." Set the audtio to whatever you set as `AUD` on the server.

Then, you can create a role with a trust relationship to the provider. It will look something like:

```json
{
  "Version": "2012-10-17",
  "Statement": [
      {
          "Effect": "Allow",
          "Action": "sts:AssumeRoleWithWebIdentity",
          "Principal": {
              "Federated": "arn:aws:iam::976079455550:oidc-provider/<quick tunnel domain>/oidc2/"
          },
          "Condition": {
              "StringEquals": {
                  "<quick tunnel domain>/oidc2/:aud": "sts.amazonaws.com",
                  ...
              }
          }
      }
  ]
}
```
