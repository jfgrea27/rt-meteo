# 001 Project initial assumptions

This Architecture Decision Log (ADL) includes initial assumptions of the project. These are subject to change and a new ADL will be written in such case.

## Assumptions

### Assumption 1:

By realtime in this context, we mean every hour. Weather doesn't change that fast and this is deemed a suitable time window.

### Assumption 2:

Users will be able to access this project through the web securely. We will inforce API rate limiting & authz to secure the application.

### Assumption 3:

Since this is a learning project, this project will use certain cloud services which might be overkill but which will educate me. This may include AWS EKS, SQS, etc.

### Assumption 4:

The project will assume a datalake pattern in case we want to provide further analytics or data science workflows in the future. This means we will store the raw weather data in some object storage as well as the sanitized data in some database for application workloads.
