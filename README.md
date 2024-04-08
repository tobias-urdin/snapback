# Snapback

Snapback was a quick POC on implementing a binary wire protocol transported
over QUIC for UDP based fast, reliable and encrypted communication between
components.

**This was never finished and is NOT usable in its current state.**

The binary protocol uses CBOR (Concise Binary Object Representation, RFC8949)
for encoding the data structures when being transmitted over a QUIC stream.

In this case the POC was about transfering snapshots of Ceph RBD images
between clusters, where the Snapback Exporter component had read-only
access to the source cluster and the Snapback Importer runs remotely
on the backup cluster and reads data from the source cluster.

This way none of the sides has direct write-access to the other cluster
to improve the security posture.

## History

As the greatest lyricist of all time said.

    Snap back to reality, ope, there goes gravity
    - Eminem, Lose Yourself, 2002
