buf.gen:
	buf generate

run.log.idm: buf.gen
	go run main.go cloudwatch -group ecs-staging -stream idm-backend-staging-log-stream -proto ./proto/cloudwatch/log.proto
