#!/bin/bash

OUT=pkg/client

java -jar $OPENAPI_CLI_JAR generate \
  -g go \
  -i api/openapi.json \
  -t api/openapi-templates/Go \
  -o $OUT \
  --package-name=client \
  --import-mappings=Host=github.com/ubccr/grendel/pkg/model,NetInterface=github.com/ubccr/grendel/pkg/model,BootImage=github.com/ubccr/grendel/pkg/model,User=github.com/ubccr/grendel/pkg/model,DataDump=github.com/ubccr/grendel/pkg/model \
  --type-mappings=User=model.User,Host=model.Host,NetInterface=model.NetInterface,BootImage=model.BootImage,DataDump=model.DataDump

# TODO This is very hackish. Figure out how to properly support external models
# in Go
sed -i 's/(\[\]Host,/(model\.HostList,/g' $OUT/api_host.go
sed -i 's/localVarReturnValue  \[\]Host/localVarReturnValue  model\.HostList/g' $OUT/api_host.go
sed -i 's/body \[\]Host/body model\.HostList/g' $OUT/api_host.go
sed -i 's/(\[\]BootImage,/(model\.BootImageList,/g' $OUT/api_image.go
sed -i 's/localVarReturnValue  \[\]BootImage/localVarReturnValue  model\.BootImageList/g' $OUT/api_image.go
sed -i 's/body \[\]BootImage/body model\.BootImageList/g' $OUT/api_image.go
sed -i 's/(\[\]User,/(\[\]model\.User,/g' $OUT/api_user.go
sed -i 's/localVarReturnValue  \[\]User/localVarReturnValue  \[\]model\.User/g' $OUT/api_user.go
sed -i 's/body \[\]User/body \[\]model\.User/g' $OUT/api_user.go
sed -i 's/(DataDump,/(model\.DataDump,/g' $OUT/api_restore.go
sed -i 's/localVarReturnValue  DataDump/localVarReturnValue  model\.DataDump/g' $OUT/api_restore.go
sed -i 's/body DataDump/body model\.DataDump/g' $OUT/api_restore.go

# TODO Again, figure out how to disable generating these. Wasn't obvious how to
# do this
rm -f $OUT/.travis.yml
rm -f $OUT/go.mod
rm -f $OUT/go.sum
rm -f $OUT/git_push.sh
