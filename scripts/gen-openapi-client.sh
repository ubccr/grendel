#!/bin/bash

OUT=pkg/client

java -jar $OPENAPI_CLI_JAR generate \
  -g go \
  -i api/openapi.json \
  -t api/openapi-templates/Go \
  -o $OUT \
  --package-name=client \
  --import-mappings=Host=github.com/ubccr/grendel/pkg/model,NetInterface=github.com/ubccr/grendel/pkg/model,BootImage=github.com/ubccr/grendel/pkg/model \
  --type-mappings=Host=model.Host,NetInterface=model.NetInterface,BootImage=model.BootImage

# TODO This is very hackish. Figure out how to properly support external models
# in Go
sed -i 's/(\[\]Host,/(model\.HostList,/g' $OUT/api_host.go
sed -i 's/localVarReturnValue  \[\]Host/localVarReturnValue  model\.HostList/g' $OUT/api_host.go
sed -i 's/body \[\]Host/body model\.HostList/g' $OUT/api_host.go
sed -i 's/(\[\]BootImage,/(model\.BootImageList,/g' $OUT/api_image.go
sed -i 's/localVarReturnValue  \[\]BootImage/localVarReturnValue  model\.BootImageList/g' $OUT/api_image.go
sed -i 's/body \[\]BootImage/body model\.BootImageList/g' $OUT/api_image.go

# TODO Again, figure out how to disable generating these. Wasn't obvious how to
# do this
rm -f $OUT/.travis.yml
rm -f $OUT/go.mod
rm -f $OUT/go.sum
rm -f $OUT/git_push.sh
