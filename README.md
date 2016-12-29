# Autoscaler

## Build

Clone this repo, then ensure glide is installed 

OSX: `brew install glide`

Install all the dependencies (`/vendor`):

`glide install`

## Testing

Run tests like this: `go test ./manager/...`

## Configuration

Copy the `config.yaml.dist` file to `config.yaml`