# usersvc
User service for Studiously

usersvc functions as both an identity provider for Hydra (OAuth2) and as a client app, responsible for user account settings. In the end, it should amount to the functional equivalent of Google's auth service.

## Run (Quick)

```
export FORCE_ROOT_CLIENT_CREDENTIALS=demo:demo
export CONSENT_URL=http://localhost:8080/consent
export DATABASE_URL=memory
hydra host --dangerous-force-http
```

then

```
hydra clients import client.json
hydra policies create -f policy.json
hydra policies create -f policy_2.json
```

## TODO

- [ ] Finish
- [ ] Complete Refactor
- [ ] Unit Tests
- [ ] System/Integration Tests