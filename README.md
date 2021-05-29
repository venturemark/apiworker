# apiworker

Worker process for managing api resources asynchronously.

### Secret

The `apiworker` process is responsible for sending email reminders. Right now
there is a secret manually deployed to the Kubernetes cluster which requires the
following structure. More information about Postmark can be found at https://postmarkapp.com.

```
data:
  postmark.token.account: <token>
  postmark.token.server: <token>
```
