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

Ehem, by which I mean, each proposal is just a tiny file in a Google Cloud
Storage bucket, and a tiny Go server to submit proposals and fetch them by
checksum, deployed to App Engine.

## Concerns

There shouldn't be any personally identifiable information in the proposals
submitted, since we're not yet supporting free-form comments. There's no user
identity at all -- submissions are anonymous, we don't even hold touch IP
addresses. If somebody decides to DDOS the API, GAE probably has some basic
abuse mitigation somewhere.

## Development notes

I haven't worked out how to glue the Go client library up to the same local
credentials that `gsutil` uses, so no local development. `gcloud add deploy`
straight to prod like a boss!

Peek in at the activity: `gsutil ls -R gs://aorta-routes.appspot.com`
