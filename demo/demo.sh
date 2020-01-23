go build ../src/main.go
./main ./openflow_dump
java -DPLANTUML_LIMIT_SIZE=8192 -jar plantuml.jar  -o ./ out.puml