# Go Project US@

This module implements the [Project US@ standard](https://asapnet.org/wp-content/uploads/2022/03/Project_US_FINAL_Technical_Specification_Version_1.0.pdf) (which is an extension of the USPS publication 28 standard) for US Address Normalization directly in Go.

## Alternatives and related technologies

There are several related technologies that could be used instead of or in conjunction with Go Project US@.

- [ProjectUsNormalizer](https://github.com/ica-carealign/project-us-normalizer): "C# library to facilitate address normalization in accordance with Project US@, the "Unified Specification for Addresses in Health Care." Portobello couldn't use this directly since we want a Go implementation.
- [golang-address](https://github.com/kminehart/golang-address) is 10 years old and recommends using [gopostal](https://github.com/openvenues/gopostal). It also only parses the street line of adresses.
- [gopostal](https://github.com/openvenues/gopostal) uses machine learning to select an appropriate parser for international addresses. It is built on [libpostal](https://github.com/openvenues/libpostal) and requires that library to be installed. libpostal is broadly deployed and used by several projects. Employing it through [libpostal-rest](https://github.com/johnlonganecker/libpostal-rest) on a [docker image](https://github.com/johnlonganecker/libpostal-rest-docker) might be a great way to tap in to that community.
- [Boostport address](https://github.com/Boostport/address) does address validation rather than normalization. It might pair well with a normalization library.
- [Pelias](https://pelias.io/) and their [docker](https://github.com/pelias/docker/) deployment could be interesting if you wanted to geocode two addresses and compare the location information or geocode an address and then reverse-geocode it to potentially get more information about it.

Ultimately, Portobello wanted a native Go implementation of the [Project US@ standard](https://asapnet.org/wp-content/uploads/2022/03/Project_US_FINAL_Technical_Specification_Version_1.0.pdf) available that didn't involve any external library installs or calls to other services. It isn't intended to do anything more or less than that.
