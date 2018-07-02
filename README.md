# hrbrmaster

## Description

hrbrmster (Harbour Master) is a CLI tool to get information about Docker Image Tags from Docker Hub.

It's a prety basic implementation and the motivation behind it was mainly to explore Golang more.

Definitely needs some improvements, especially with performance. Next stage is to implement channel behaviour


## Usage

Usage instructions with code examples

```shell
# Install locally if you have Go installed
# This will go to $GPATH/bin so make sure that is in your $PATH
go install

# To Run Test Suite on imageinfo
go test -v imageinfo

# To get some info about the Golang Docker image
hrbrmster library/golang
```

## TODO

- [ ] Improve Performance by Implementing Concurrency
- [ ] Update main.go to use flags. Maybe extend to use cobra in future
- [ ] Add Mocking to imageinfo_test.go

## License

[MIT](./LICENSE)
