# yimbyhoodlum

Your friendly local YIMBYhoodlum hosts proposals created in A/B Street, letting
people share them without the hassle of manually printing and posting flyers
with JSON blobs on utility poles.

## Goals

1) Let people share a URL with their proposal.
2) Let people browse proposals that have passed manual moderation.

Non-goals:

1) No adding comments/feedback to a proposal from within A/B Street. People can
   share a URL on Twitter/wherever else with their thoughts, or further modify the
   proposal and share their own. This avoids huge moderation/spam/abuse problems.
2) All proposals are submitted anonymously and are immutable. No editing or
   deleting or proving who created a proposal. It's too early figure out user
   management. Will have a clear warning before submission -- releasing
   anonymously, under a creative commons license.

## Architecture

Proposals are recorded onto a blockchain and simulated using homomorphic
encryption, with the results being implemented in a digital twin governed by
smart contracts.

Ehem, by which I mean, there's just a single-table SQL database and a tiny Go
server to submit proposals, fetch them by ID, and browse moderated ones.

For local development, just using SQLite. Probably going to deploy on App
Engine with Cloud SQL, or equivalent.

## Concerns

There shouldn't be any personally identifiable information in the proposals
submitted, since we're not yet supporting free-form comments. There's no user
identity at all -- submissions are anonymous, we don't even hold touch IP
addresses. If somebody decides to DDOS the API, GAE probably has some basic
abuse mitigation somewhere.

## Development notes

To run locally, you need Go and MySQL running somewhere. Docker is easiest for
running the DB. You also need the MySQL client.

- Start MySQL: `docker run -p 3306:3306 --name yimbysql -e MYSQL_ROOT_PASSWORD=password -d mysql:5.7`
- Add a blank DB: `mysql -u root -ppassword -h 0.0.0.0 -P 3306 -e 'CREATE DATABASE dev'`
- Start the GO server: `MYSQL_URI='root:password@tcp(0.0.0.0:3306)/dev' go run main.go`
- To debug, you can grab container logs: `docker logs yimbysql`
- To teardown: `docker rm --force yimbysql`

I wish setting up GAE and Cloud SQL was more declarative / reproducible. I
didn't hit any snags following various bits of documentation and using the
console, so not saying much here. `gcloud app deploy`, then fiddle around with
logs. The only trick is that app.yaml isn't under version control, because
seemingly I have to copy the MySQL root password there? There's definitely a
way to make IAM work out here, but it's not obvious from docs.
