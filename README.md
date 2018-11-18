# registry-sync

Most sync tools use Docker to pull the image and then push it again, which is not very efficient. This also means you need Docker installed.

registry-sync is a tool for synchronizing Docker images from one Docker registry to another, directly.  It transfers the Docker layers directly from the source registry to the destination, skipping layers that are already there, so it's efficient.

# Example Usage
~~~~
registry-sync -source <registry-url> -destination <registry-url> -image example/image -tag 1.0

registry-sync -source <registry-url> -source-user <username> -source-pass <password> -destination <registry-url> -destination-user <username> -destination-pass <password> -image example/image -tag 1.0
~~~~
