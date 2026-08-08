# Libpostal HTTP wrapper

This package is an http wrapper that calls libpostal running in another service,
as provided by the
[pelias/libpostal-service](https://hub.docker.com/r/pelias/libpostal-service/tags)
docker image. More can be learned about this image at its
[github repository](https://github.com/pelias/libpostal-service) and the
[Who's on First](https://github.com/whosonfirst/go-whosonfirst-libpostal)
service it runs.

We do not believe that it makes sense for this library to directly or indirectly
require the installation of libpostal for the process using this library. This
http interface allows users to make use of libpostal without binding all users
go processes to it.
